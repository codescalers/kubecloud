package internal

import (
	"context"
	"encoding/json"
	"io"
	"kubecloud/models"
	"net/http"
	"sync"
	"time"

	"kubecloud/internal/logger"

	"github.com/gin-gonic/gin"
)

// Notification types
const (
	Success = "success"
	Info    = "info"
	Error   = "error"
)

// SSEClient represents a single SSE client connection
type SSEClient struct {
	Channel      chan SSEMessage
	Token        string
	ConnectedAt  time.Time
	CancelFunc   context.CancelFunc
}

// SSEManager handles Server-Sent Events for real-time notifications
type SSEManager struct {
	clients       map[int][]*SSEClient // userID -> client connections
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	tokenManager  TokenManager
	checkInterval time.Duration
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
func NewSSEManager() *SSEManager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &SSEManager{
		clients:       make(map[int][]*SSEClient),
		ctx:           ctx,
		cancel:        cancel,
		checkInterval: 30 * time.Second, // Check tokens every 30 seconds
	}

	return manager
}

// SetTokenManager sets the token manager for token validation
func (s *SSEManager) SetTokenManager(tm TokenManager) {
	s.tokenManager = tm
	// Start the token validation goroutine
	go s.validateTokensPeriodically()
}

// Stop gracefully shuts down the SSE manager
func (s *SSEManager) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Close all client channels and cancel their contexts
	for userID, clients := range s.clients {
		for _, client := range clients {
			if client.CancelFunc != nil {
				client.CancelFunc()
			}
			close(client.Channel)
		}
		delete(s.clients, userID)
	}

}

// AddClient adds a new client connection for a user with token tracking
func (s *SSEManager) AddClient(userID int, token string) (*SSEClient, context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a cancellable context for this specific client
	ctx, cancel := context.WithCancel(s.ctx)

	client := &SSEClient{
		Channel:      make(chan SSEMessage, 10),
		Token:        token,
		ConnectedAt:  time.Now(),
		CancelFunc:   cancel,
	}
	s.clients[userID] = append(s.clients[userID], client)

	return client, ctx
}

// RemoveClient removes a client connection for a user
func (s *SSEManager) RemoveClient(userID int, client *SSEClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clients := s.clients[userID]
	for i, c := range clients {
		if c == client {
			s.clients[userID] = append(clients[:i], clients[i+1:]...)
			if client.CancelFunc != nil {
				client.CancelFunc()
			}
			close(client.Channel)
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
		Type:     msgType,
		Severity: string(severity),
		Data: map[string]string{
			"message": data["message"],
			"status":  data["status"],
		},
		Timestamp: time.Now(),
		ID:        id,
	}

	if len(taskID) > 0 {
		message.TaskID = taskID[0]
	}

	s.mu.RLock()
	clients := make([]*SSEClient, len(s.clients[userID]))
	copy(clients, s.clients[userID])
	s.mu.RUnlock()

	// Send to all user's clients
	for _, client := range clients {
		select {
		case client.Channel <- message:
			// Message sent successfully
		case <-time.After(2 * time.Second):
			// Client not responding, remove it
			go s.RemoveClient(userID, client)
		case <-s.ctx.Done():
			return
		}
	}

}

// validateTokensPeriodically checks all active SSE connections and closes those with expired tokens
func (s *SSEManager) validateTokensPeriodically() {
	if s.tokenManager == nil {
		logger.GetLogger().Warn().Msg("Token manager not set, skipping periodic token validation")
		return
	}

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.validateAllTokens()
		case <-s.ctx.Done():
			return
		}
	}
}

// validateAllTokens validates tokens for all active SSE connections
func (s *SSEManager) validateAllTokens() {
	s.mu.RLock()
	// Create a snapshot of all clients to validate
	clientsToValidate := make(map[int][]*SSEClient)
	for userID, clients := range s.clients {
		clientsToValidate[userID] = make([]*SSEClient, len(clients))
		copy(clientsToValidate[userID], clients)
	}
	s.mu.RUnlock()

	// Validate tokens outside the lock to avoid blocking
	for userID, clients := range clientsToValidate {
		for _, client := range clients {
			_, err := s.tokenManager.VerifyToken(client.Token)
			if err != nil {
				logger.GetLogger().Info().
					Int("user_id", userID).
					Err(err).
					Msg("SSE connection token expired, closing connection")
				// Remove the client with expired token
				s.RemoveClient(userID, client)
			}
		}
	}
}

// HandleSSE handles SSE HTTP connections
func (s *SSEManager) HandleSSE(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get the token from the context (set by the middleware)
	token := c.Query("token")
	if token == "" {
		// Fallback to Authorization header if available
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			token = authHeader[len("Bearer "):]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token not found"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Add client and get client object with its context
	client, clientCtx := s.AddClient(userID, token)
	defer s.RemoveClient(userID, client)

	// Send initial connection message
	s.Notify(userID, "connected", models.NotificationSeverityInfo, map[string]string{"status": "connected"}, "")

	// Stream messages to client
	c.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-client.Channel:
			if !ok {
				return false // Channel closed
			}

			data, err := json.Marshal(message)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to marshal SSE message")
				return false
			}

			c.SSEvent("message", string(data))
			return true

		case <-clientCtx.Done():
			logger.GetLogger().Debug().Int("user_id", userID).Msg("SSE client context cancelled (token expired or connection closed)")
			return false

		case <-c.Request.Context().Done():
			logger.GetLogger().Debug().Int("user_id", userID).Msg("Client disconnected")
			return false

		case <-s.ctx.Done():
			return false
		}
	})
}
