package sse

import (
	"bufio"
	"context"
	"kubecloud/internal"
	"kubecloud/internal/mocks"
	"kubecloud/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandleSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var currentManager *SSEManager
	var currentUserID int

	router := gin.New()
	router.Use(func(c *gin.Context) {
		if currentUserID != 0 {
			c.Set("user_id", currentUserID)
		}
		c.Next()
	})
	router.GET("/events", func(c *gin.Context) {
		currentManager.HandleSSE(c)
	})
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Valid token from query parameter", func(t *testing.T) {
		mockTM, manager := setupTestManager(ctrl)
		currentManager = manager
		currentUserID = 123

		claims := setupValidClaims(123, "testuser", 1*time.Hour)
		mockTM.EXPECT().VerifyToken("valid-token").Return(claims, nil)

		// Make SSE request
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events?token=valid-token", nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")

		// Verify connection established
		time.Sleep(100 * time.Millisecond)
		assertClientCount(t, manager, 123, 1)

		// Send notification
		manager.Notify(123, "test", models.NotificationSeverityInfo,
			map[string]string{"message": "test msg", "status": "ok"}, "msg-1")

		// Read SSE message
		scanner := bufio.NewScanner(resp.Body)
		messageFound := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "test msg") {
				messageFound = true
				break
			}
		}
		assert.True(t, messageFound, "Should receive notification message")

		manager.Stop()
	})

	t.Run("Missing user_id returns unauthorized", func(t *testing.T) {
		_, manager := setupTestManager(ctrl)
		currentManager = manager
		currentUserID = 0 // No user_id

		resp, err := http.Get(server.URL + "/events?token=valid-token")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Missing token returns unauthorized", func(t *testing.T) {
		_, manager := setupTestManager(ctrl)
		currentManager = manager
		currentUserID = 123

		resp, err := http.Get(server.URL + "/events")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Token with no expiry still allows connection", func(t *testing.T) {
		mockTM, manager := setupTestManager(ctrl)
		currentManager = manager
		currentUserID = 456

		claims := &internal.TokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: nil, // No expiry
			},
			UserID:   456,
			Username: "noexpiry",
		}
		mockTM.EXPECT().VerifyToken("no-expiry-token").Return(claims, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events?token=no-expiry-token", nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(50 * time.Millisecond)
		assertClientCount(t, manager, 456, 1)

		manager.Stop()
	})

	t.Run("Invalid token returns unauthorized", func(t *testing.T) {
		mockTM, manager := setupTestManager(ctrl)
		currentManager = manager
		currentUserID = 789

		mockTM.EXPECT().VerifyToken("invalid-token").Return(nil, assert.AnError)

		resp, err := http.Get(server.URL + "/events?token=invalid-token")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Token expiry closes connection automatically", func(t *testing.T) {
		mockTM, manager := setupTestManager(ctrl)
		currentManager = manager
		currentUserID = 999

		// Token expires in 400ms (enough time to establish connection, then expire quickly)
		claims := setupValidClaims(999, "expireuser", 500*time.Millisecond)
		mockTM.EXPECT().VerifyToken("expiring-token").Return(claims, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events?token=expiring-token", nil)

		connClosed := make(chan bool, 1)
		go func() {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				// Read until connection closes
				buf := make([]byte, 1024)
				for {
					_, err := resp.Body.Read(buf)
					if err != nil {
						break
					}
				}
			}
			connClosed <- true
		}()

		// Verify connection established
		time.Sleep(200 * time.Millisecond)
		assertClientCount(t, manager, 999, 1)

		// Wait for token to expire and connection to close
		select {
		case <-connClosed:
			// Connection closed as expected
			time.Sleep(20 * time.Millisecond)
			assertClientCount(t, manager, 999, 0)
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Connection did not close after token expiry")
		}
	})
}

func TestNotify(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, manager := setupTestManager(ctrl)

	userID := 1
	clientChan := manager.addClient(userID)

	// Send notification in goroutine
	go func() {
		manager.Notify(userID, "test", models.NotificationSeverityInfo,
			map[string]string{"message": "test message", "status": "ok"}, "test-id")
	}()

	// Receive notification
	select {
	case msg := <-clientChan:
		assert.Equal(t, "test", msg.Type)
		assert.Equal(t, string(models.NotificationSeverityInfo), msg.Severity)
		assert.Equal(t, "test message", msg.Data["message"])
		assert.Equal(t, "test-id", msg.ID)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for notification")
	}

	manager.Stop()
}

// Test helpers
func setupTestManager(ctrl *gomock.Controller) (*mocks.MockTokenManager, *SSEManager) {
	mockTM := mocks.NewMockTokenManager(ctrl)
	manager := NewSSEManager(mockTM)
	return mockTM, manager
}

func setupValidClaims(userID int, username string, expiresIn time.Duration) *internal.TokenClaims {
	return &internal.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
		UserID:   userID,
		Username: username,
	}
}

func assertClientCount(t *testing.T, manager *SSEManager, userID int, expected int) {
	manager.mu.RLock()
	actual := len(manager.clients[userID])
	manager.mu.RUnlock()
	assert.Equal(t, expected, actual)
}
