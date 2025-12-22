package realtime

import (
	"context"
	"encoding/json"
	"io"
	"kubecloud/internal/auth"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/logger"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Notification types
const (
	Success = "success"
	Info    = "info"
	Error   = "error"
)

// SSEManager handles Server-Sent Events for real-time notifications
type SSEManager struct {
	clients      map[int][]chan SSEMessage // userID -> client channels
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	tokenManager auth.TokenManager
}

// SSEMessage represents a server-sent event message
type SSEMessage struct {
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
	Severity  string            `json:"severity"`
	TaskID    string            `json:"task_id,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	ID        string            `json:"id,omitempty"`
}

// NewSSEManager creates a new SSE manager
func NewSSEManager(tokenManager auth.TokenManager) *SSEManager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &SSEManager{
		clients:      make(map[int][]chan SSEMessage),
		ctx:          ctx,
		cancel:       cancel,
		tokenManager: tokenManager,
	}

	return manager
}

// Stop gracefully shuts down the SSE manager
func (s *SSEManager) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Close all client channels
	for userID, channels := range s.clients {
		for _, ch := range channels {
			close(ch)
		}
		delete(s.clients, userID)
	}

}

// AddClient adds a new client channel for a user
func (s *SSEManager) addClient(userID int) chan SSEMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan SSEMessage, 10)
	s.clients[userID] = append(s.clients[userID], ch)

	return ch
}

// RemoveClient removes a client channel for a user
func (s *SSEManager) removeClient(userID int, clientChan chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channels := s.clients[userID]
	for i, ch := range channels {
		if ch == clientChan {
			s.clients[userID] = append(channels[:i], channels[i+1:]...)
			close(ch)
			break
		}
	}

	if len(s.clients[userID]) == 0 {
		delete(s.clients, userID)
	}
}

// Notify sends a message to all clients of a specific user
func (s *SSEManager) Notify(userID int, msgType string, severity models.NotificationSeverity, data map[string]string, id string, taskID ...string) {
	message := SSEMessage{
		Type:      msgType,
		Severity:  string(severity),
		Data:      data, // Pass complete data, not filtered
		Timestamp: time.Now(),
		ID:        id,
	}

	if len(taskID) > 0 {
		message.TaskID = taskID[0]
	}

	s.mu.RLock()
	channels := make([]chan SSEMessage, len(s.clients[userID]))
	copy(channels, s.clients[userID])
	s.mu.RUnlock()

	// Send to all user's clients
	for _, ch := range channels {
		select {
		case ch <- message:
			// Message sent successfully
		case <-time.After(2 * time.Second):
			// Client not responding, remove it
			go s.removeClient(userID, ch)
		case <-s.ctx.Done():
			return
		}
	}

}

// setupExpiryTimer creates a timer that fires when the token expires
func (s *SSEManager) setupExpiryTimer(claims *auth.TokenClaims) *time.Timer {
	if claims == nil || claims.ExpiresAt == nil {
		return nil
	}

	return time.NewTimer(time.Until(claims.ExpiresAt.Time))
}

// HandleSSE handles SSE HTTP connections
func (s *SSEManager) HandleSSE(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Extract token from query or Authorization header
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	}

	claims, err := s.tokenManager.VerifyToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Setup token expiry enforcement timer using claims
	expiryTimer := s.setupExpiryTimer(claims)
	if expiryTimer != nil {
		defer expiryTimer.Stop()
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Add client and get channel
	clientChan := s.addClient(userID)
	defer s.removeClient(userID, clientChan)

	log := logger.ForOperation("sse", "handle_connection").With().Int("user_id", userID).Logger()

	// Send initial connection message
	s.Notify(userID, "connected", models.NotificationSeverityInfo, map[string]string{"status": "connected"}, "")

	var expiryC <-chan time.Time
	if expiryTimer != nil {
		expiryC = expiryTimer.C
	}

	// Stream messages to client
	c.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-clientChan:
			if !ok {
				return false // Channel closed
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal SSE message")
				return false
			}

			c.SSEvent("message", string(data))
			return true

		case <-expiryC:
			// Token expired (nil channel never fires)
			return false

		case <-c.Request.Context().Done():
			log.Debug().Msg("Client disconnected")
			return false

		case <-s.ctx.Done():
			return false
		}
	})
}
