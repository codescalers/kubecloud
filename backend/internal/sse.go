package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SSEManager manages Server-Sent Events connections
type SSEManager struct {
	clients    map[string]map[chan SSEMessage]bool // userID -> channels
	clientsMux sync.RWMutex
	redis      *RedisClient
}

// SSEMessage represents a message sent via SSE
type SSEMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	TaskID    string      `json:"task_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewSSEManager creates a new SSE manager
func NewSSEManager(redis *RedisClient) *SSEManager {
	manager := &SSEManager{
		clients: make(map[string]map[chan SSEMessage]bool),
		redis:   redis,
	}

	// Start listening for results from Redis
	go manager.listenForResults()

	return manager
}

// AddClient adds a new SSE client for a user
func (s *SSEManager) AddClient(userID string, clientChan chan SSEMessage) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	if s.clients[userID] == nil {
		s.clients[userID] = make(map[chan SSEMessage]bool)
	}
	s.clients[userID][clientChan] = true

	// TODO: how client-chan will connect to this
	log.Info().Str("user_id", userID).Msg("SSE client connected")
}

// RemoveClient removes an SSE client
func (s *SSEManager) RemoveClient(userID string, clientChan chan SSEMessage) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	if clients, exists := s.clients[userID]; exists {
		delete(clients, clientChan)
		if len(clients) == 0 {
			delete(s.clients, userID)
		}
	}
	close(clientChan)

	log.Info().Str("user_id", userID).Msg("SSE client disconnected")
}

// BroadcastToUser sends a message to all connections for a specific user
func (s *SSEManager) BroadcastToUser(userID string, message SSEMessage) {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	if clients, exists := s.clients[userID]; exists {
		message.Timestamp = time.Now()
		for clientChan := range clients {
			select {
			case clientChan <- message:
			case <-time.After(time.Second * 5):
				// Client is not responding, remove it
				go s.RemoveClient(userID, clientChan)
			}
		}
	}
}

// listenForResults listens for deployment results from Redis and broadcasts them
func (s *SSEManager) listenForResults() {
	ctx := context.Background()

	err := s.redis.SubscribeToResults(ctx, func(result *DeploymentResult) {
		// Get the task to find the user ID
		// In a real implementation, you'd store task metadata or query it
		message := SSEMessage{
			Type:   "deployment_update",
			Data:   result,
			TaskID: result.TaskID,
		}

		// For now, broadcast to all users (you'd want to filter by user)
		s.broadcastToAll(message)
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to listen for results")
	}
}

// broadcastToAll broadcasts a message to all connected clients
func (s *SSEManager) broadcastToAll(message SSEMessage) {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	for userID := range s.clients {
		s.BroadcastToUser(userID, message)
	}
}

// HandleSSE handles SSE connections
func (s *SSEManager) HandleSSE(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Create client channel
	clientChan := make(chan SSEMessage, 10)
	s.AddClient(userIDStr, clientChan)

	// Send initial connection message
	initialMessage := SSEMessage{
		Type: "connected",
		Data: map[string]string{"status": "connected"},
	}
	s.BroadcastToUser(userIDStr, initialMessage)

	// Handle client disconnection
	defer s.RemoveClient(userIDStr, clientChan)

	// Keep connection alive and send messages
	c.Stream(func(w io.Writer) bool {
		select {
		case message := <-clientChan:
			data, err := json.Marshal(message)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal SSE message")
				return false
			}
			c.SSEvent("message", string(data))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// SendTaskUpdate sends a task status update to the user
func (s *SSEManager) SendTaskUpdate(userID string, taskID string, status TaskStatus, message string, data interface{}) {
	sseMessage := SSEMessage{
		Type:   "task_update",
		TaskID: taskID,
		Data: map[string]interface{}{
			"status":  status,
			"message": message,
			"data":    data,
		},
	}
	s.BroadcastToUser(userID, sseMessage)
}
