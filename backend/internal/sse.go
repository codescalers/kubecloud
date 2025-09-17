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

// SSEManager handles Server-Sent Events for real-time notifications
type SSEManager struct {
	clients map[int][]chan SSEMessage // userID -> client channels
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// SSEMessage represents a server-sent event message
type SSEMessage struct {
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
	Severity  string            `json:"severity"`
	TaskID    string            `json:"task_id,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewSSEManager creates a new SSE manager
func NewSSEManager() *SSEManager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &SSEManager{
		clients: make(map[int][]chan SSEMessage),
		ctx:     ctx,
		cancel:  cancel,
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
func (s *SSEManager) AddClient(userID int) chan SSEMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan SSEMessage, 10)
	s.clients[userID] = append(s.clients[userID], ch)

	return ch
}

// RemoveClient removes a client channel for a user
func (s *SSEManager) RemoveClient(userID int, clientChan chan SSEMessage) {
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
func (s *SSEManager) Notify(userID int, msgType string, severity models.NotificationSeverity, data map[string]string, taskID ...string) {
	message := SSEMessage{
		Type:     msgType,
		Severity: string(severity),
		Data: map[string]string{
			"message": data["message"],
			"status":  data["status"],
		},
		Timestamp: time.Now(),
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
			go s.RemoveClient(userID, ch)
		case <-s.ctx.Done():
			return
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

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Add client and get channel
	clientChan := s.AddClient(userID)
	defer s.RemoveClient(userID, clientChan)

	// Send initial connection message
	s.Notify(userID, "connected", models.NotificationSeverityInfo, map[string]string{"status": "connected"})

	// Stream messages to client
	c.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-clientChan:
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

		case <-c.Request.Context().Done():
			logger.GetLogger().Debug().Int("user_id", userID).Msg("Client disconnected")
			return false

		case <-s.ctx.Done():
			return false
		}
	})
}
