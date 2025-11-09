package internal

import (
	"kubecloud/models"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTokenManager is a mock implementation of TokenManager for testing
type MockTokenManager struct {
	mu            sync.RWMutex
	validTokens   map[string]*TokenClaims
	shouldExpire  bool
	expireAfter   time.Duration
	creationTimes map[string]time.Time
}

func NewMockTokenManager() *MockTokenManager {
	return &MockTokenManager{
		validTokens:   make(map[string]*TokenClaims),
		creationTimes: make(map[string]time.Time),
	}
}

func (m *MockTokenManager) CreateTokenPair(userID int, username string, isAdmin bool) (*TokenPair, error) {
	token := "mock_token_" + username
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.validTokens[token] = &TokenClaims{
		UserID:   userID,
		Username: username,
		Admin:    isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	m.creationTimes[token] = time.Now()
	
	return &TokenPair{
		AccessToken:  token,
		RefreshToken: "mock_refresh_" + username,
	}, nil
}

func (m *MockTokenManager) VerifyToken(tokenString string) (*TokenClaims, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	claims, exists := m.validTokens[tokenString]
	if !exists {
		return nil, jwt.ErrSignatureInvalid
	}
	
	// Check if we should simulate expiration
	if m.shouldExpire && m.expireAfter > 0 {
		creationTime := m.creationTimes[tokenString]
		if time.Since(creationTime) > m.expireAfter {
			return nil, jwt.ErrTokenExpired
		}
	}
	
	return claims, nil
}

func (m *MockTokenManager) AccessTokenFromRefresh(refreshToken string) (string, error) {
	return "mock_new_access_token", nil
}

func (m *MockTokenManager) InvalidateToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.validTokens, token)
	delete(m.creationTimes, token)
}

func (m *MockTokenManager) SetExpireAfter(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldExpire = true
	m.expireAfter = duration
}

func TestSSEManager_AddClient(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()
	
	mockToken := NewMockTokenManager()
	manager.SetTokenManager(mockToken)
	
	// Create a token
	tokenPair, err := mockToken.CreateTokenPair(1, "testuser", false)
	require.NoError(t, err)
	
	// Add a client
	client, ctx := manager.AddClient(1, tokenPair.AccessToken)
	
	assert.NotNil(t, client)
	assert.NotNil(t, client.Channel)
	assert.Equal(t, tokenPair.AccessToken, client.Token)
	assert.NotNil(t, ctx)
	
	// Verify client is tracked
	manager.mu.RLock()
	clients := manager.clients[1]
	manager.mu.RUnlock()
	
	assert.Len(t, clients, 1)
	assert.Equal(t, client, clients[0])
}

func TestSSEManager_RemoveClient(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()
	
	mockToken := NewMockTokenManager()
	manager.SetTokenManager(mockToken)
	
	// Create a token and add client
	tokenPair, err := mockToken.CreateTokenPair(1, "testuser", false)
	require.NoError(t, err)
	
	client, _ := manager.AddClient(1, tokenPair.AccessToken)
	
	// Verify client exists
	manager.mu.RLock()
	assert.Len(t, manager.clients[1], 1)
	manager.mu.RUnlock()
	
	// Remove client
	manager.RemoveClient(1, client)
	
	// Verify client is removed
	manager.mu.RLock()
	_, exists := manager.clients[1]
	manager.mu.RUnlock()
	
	assert.False(t, exists, "User should have no clients after removal")
}

func TestSSEManager_Notify(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()
	
	mockToken := NewMockTokenManager()
	manager.SetTokenManager(mockToken)
	
	// Create a token and add client
	tokenPair, err := mockToken.CreateTokenPair(1, "testuser", false)
	require.NoError(t, err)
	
	client, _ := manager.AddClient(1, tokenPair.AccessToken)
	
	// Send a notification
	go manager.Notify(1, "test", models.NotificationSeverityInfo, map[string]string{
		"message": "test message",
		"status":  "success",
	}, "test-id")
	
	// Receive the notification
	select {
	case msg := <-client.Channel:
		assert.Equal(t, "test", msg.Type)
		assert.Equal(t, "info", msg.Severity)
		assert.Equal(t, "test message", msg.Data["message"])
		assert.Equal(t, "test-id", msg.ID)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for notification")
	}
}

func TestSSEManager_TokenExpiration(t *testing.T) {
	// Create manager with short check interval for testing
	manager := NewSSEManager()
	manager.checkInterval = 100 * time.Millisecond // Check every 100ms
	defer manager.Stop()
	
	mockToken := NewMockTokenManager()
	// Set tokens to expire after 200ms
	mockToken.SetExpireAfter(200 * time.Millisecond)
	manager.SetTokenManager(mockToken)
	
	// Create a token and add client
	tokenPair, err := mockToken.CreateTokenPair(1, "testuser", false)
	require.NoError(t, err)
	
	client, ctx := manager.AddClient(1, tokenPair.AccessToken)
	
	// Verify client exists
	manager.mu.RLock()
	assert.Len(t, manager.clients[1], 1)
	manager.mu.RUnlock()
	
	// Wait for token to expire and be validated
	// Token expires after 200ms, check happens every 100ms
	// So we should detect expiration within 300ms
	select {
	case <-ctx.Done():
		// Context should be cancelled when token expires
		t.Log("Client context cancelled due to token expiration")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for token expiration")
	}
	
	// Verify the channel is closed
	select {
	case _, ok := <-client.Channel:
		assert.False(t, ok, "Client channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel was not closed")
	}
	
	// Give some time for cleanup
	time.Sleep(100 * time.Millisecond)
	
	// Verify client is removed from manager
	manager.mu.RLock()
	clients, exists := manager.clients[1]
	manager.mu.RUnlock()
	
	if exists {
		assert.Empty(t, clients, "No clients should remain after token expiration")
	}
}

func TestSSEManager_MultipleClientsTokenExpiration(t *testing.T) {
	// Create manager with short check interval for testing
	manager := NewSSEManager()
	manager.checkInterval = 100 * time.Millisecond
	defer manager.Stop()
	
	mockToken := NewMockTokenManager()
	mockToken.SetExpireAfter(200 * time.Millisecond)
	manager.SetTokenManager(mockToken)
	
	// Add multiple clients for the same user
	tokenPair1, err := mockToken.CreateTokenPair(1, "user1", false)
	require.NoError(t, err)
	tokenPair2, err := mockToken.CreateTokenPair(1, "user1_device2", false)
	require.NoError(t, err)
	
	client1, ctx1 := manager.AddClient(1, tokenPair1.AccessToken)
	client2, ctx2 := manager.AddClient(1, tokenPair2.AccessToken)
	
	// Verify both clients exist
	manager.mu.RLock()
	assert.Len(t, manager.clients[1], 2)
	manager.mu.RUnlock()
	
	// Wait for tokens to expire
	ctxDone := 0
	done := make(chan struct{})
	
	go func() {
		<-ctx1.Done()
		ctxDone++
		if ctxDone == 2 {
			close(done)
		}
	}()
	
	go func() {
		<-ctx2.Done()
		ctxDone++
		if ctxDone == 2 {
			close(done)
		}
	}()
	
	select {
	case <-done:
		t.Log("All client contexts cancelled due to token expiration")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for all tokens to expire")
	}
	
	// Verify both channels are closed
	_, ok1 := <-client1.Channel
	_, ok2 := <-client2.Channel
	assert.False(t, ok1, "Client1 channel should be closed")
	assert.False(t, ok2, "Client2 channel should be closed")
	
	// Give some time for cleanup
	time.Sleep(100 * time.Millisecond)
	
	// Verify all clients are removed
	manager.mu.RLock()
	clients, exists := manager.clients[1]
	manager.mu.RUnlock()
	
	if exists {
		assert.Empty(t, clients, "No clients should remain after token expiration")
	}
}

func TestSSEManager_ValidTokensNotExpired(t *testing.T) {
	// Create manager with short check interval
	manager := NewSSEManager()
	manager.checkInterval = 50 * time.Millisecond
	defer manager.Stop()
	
	mockToken := NewMockTokenManager()
	// Don't set expiration - tokens remain valid
	manager.SetTokenManager(mockToken)
	
	// Add a client
	tokenPair, err := mockToken.CreateTokenPair(1, "testuser", false)
	require.NoError(t, err)
	
	client, ctx := manager.AddClient(1, tokenPair.AccessToken)
	
	// Wait for multiple check intervals
	time.Sleep(200 * time.Millisecond)
	
	// Verify client still exists and context is not cancelled
	select {
	case <-ctx.Done():
		t.Fatal("Valid token should not cause context cancellation")
	default:
		// Context is still active, which is correct
	}
	
	// Verify client is still tracked
	manager.mu.RLock()
	clients := manager.clients[1]
	manager.mu.RUnlock()
	
	assert.Len(t, clients, 1)
	assert.Equal(t, client, clients[0])
	
	// Verify channel is still open by sending a message
	go manager.Notify(1, "test", models.NotificationSeverityInfo, map[string]string{
		"message": "still connected",
		"status":  "ok",
	}, "")
	
	select {
	case msg := <-client.Channel:
		assert.Equal(t, "still connected", msg.Data["message"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should receive message on valid connection")
	}
}

func TestSSEManager_StopClosesAllConnections(t *testing.T) {
	manager := NewSSEManager()
	
	mockToken := NewMockTokenManager()
	manager.SetTokenManager(mockToken)
	
	// Add multiple clients
	tokenPair1, err := mockToken.CreateTokenPair(1, "user1", false)
	require.NoError(t, err)
	tokenPair2, err := mockToken.CreateTokenPair(2, "user2", false)
	require.NoError(t, err)
	
	client1, ctx1 := manager.AddClient(1, tokenPair1.AccessToken)
	client2, ctx2 := manager.AddClient(2, tokenPair2.AccessToken)
	
	// Stop the manager
	manager.Stop()
	
	// Wait a bit for cleanup
	time.Sleep(50 * time.Millisecond)
	
	// Verify all contexts are cancelled
	select {
	case <-ctx1.Done():
		// Expected
	default:
		t.Fatal("Context 1 should be cancelled after Stop()")
	}
	
	select {
	case <-ctx2.Done():
		// Expected
	default:
		t.Fatal("Context 2 should be cancelled after Stop()")
	}
	
	// Verify channels are closed
	_, ok1 := <-client1.Channel
	_, ok2 := <-client2.Channel
	assert.False(t, ok1)
	assert.False(t, ok2)
	
	// Verify all clients are removed
	manager.mu.RLock()
	assert.Empty(t, manager.clients)
	manager.mu.RUnlock()
}
