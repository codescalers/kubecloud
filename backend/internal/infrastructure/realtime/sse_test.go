package realtime

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"kubecloud/internal/core/models"
)

func TestNewSSEManager(t *testing.T) {
	manager := NewSSEManager()
	if manager.clients == nil {
		t.Error("clients map not initialized")
	}
	if manager.ctx == nil {
		t.Error("context not initialized")
	}
	if manager.cancel == nil {
		t.Error("cancel function not set")
	}
}

func TestSSEManager_AddClient(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()

	userID := 1
	clientChan := manager.AddClient(userID)

	if clientChan == nil {
		t.Error("AddClient returned nil channel")
	}

	// Check that client was added
	manager.mu.RLock()
	channels, exists := manager.clients[userID]
	manager.mu.RUnlock()

	if !exists {
		t.Error("Client was not added to clients map")
	}
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}
	if channels[0] != clientChan {
		t.Error("Channel in map doesn't match returned channel")
	}
}

func TestSSEManager_RemoveClient(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()

	userID := 1
	clientChan := manager.AddClient(userID)

	// Verify client exists
	manager.mu.RLock()
	channels, exists := manager.clients[userID]
	manager.mu.RUnlock()

	if !exists || len(channels) != 1 {
		t.Error("Client was not properly added")
	}

	// Remove client
	manager.RemoveClient(userID, clientChan)

	// Verify client was removed
	manager.mu.RLock()
	_, exists = manager.clients[userID]
	manager.mu.RUnlock()

	if exists {
		t.Error("Client map entry should have been removed when last client was removed")
	}
}

func TestSSEManager_RemoveClient_MultipleClients(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()

	userID := 1
	clientChan1 := manager.AddClient(userID)
	clientChan2 := manager.AddClient(userID)

	// Verify both clients exist
	manager.mu.RLock()
	channels, exists := manager.clients[userID]
	manager.mu.RUnlock()

	if !exists || len(channels) != 2 {
		t.Error("Both clients were not properly added")
	}

	// Remove first client
	manager.RemoveClient(userID, clientChan1)

	// Verify only second client remains
	manager.mu.RLock()
	channels, exists = manager.clients[userID]
	manager.mu.RUnlock()

	if !exists || len(channels) != 1 {
		t.Errorf("Expected 1 client remaining, got %d", len(channels))
	}
	if channels[0] != clientChan2 {
		t.Error("Wrong client remaining")
	}
}

func TestSSEManager_Notify(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()

	userID := 1
	clientChan := manager.AddClient(userID)

	data := map[string]string{"message": "test message", "status": "success"}

	// Notify in a goroutine since it might block
	go manager.Notify(userID, "test", models.NotificationSeverityInfo, data, "test-id")

	// Wait for message with timeout
	select {
	case received := <-clientChan:
		if received.Type != "test" {
			t.Errorf("Expected type 'test', got %s", received.Type)
		}
		if received.Data["message"] != "test message" {
			t.Errorf("Expected message 'test message', got %s", received.Data["message"])
		}
		if received.Data["status"] != "success" {
			t.Errorf("Expected status 'success', got %s", received.Data["status"])
		}
		if received.Severity != string(models.NotificationSeverityInfo) {
			t.Errorf("Expected severity 'info', got %s", received.Severity)
		}
		if received.ID != "test-id" {
			t.Errorf("Expected ID 'test-id', got %s", received.ID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestSSEManager_Notify_MultipleClients(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()

	userID := 1
	clientChan1 := manager.AddClient(userID)
	clientChan2 := manager.AddClient(userID)

	message := SSEMessage{
		Type:     "test",
		Data:     map[string]string{"key": "value"},
		Severity: "info",
		ID:       "test-id",
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Start goroutines to receive messages
	go func() {
		defer wg.Done()
		select {
		case <-clientChan1:
			// Message received
		case <-time.After(1 * time.Second):
			t.Error("Client 1 didn't receive message")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case <-clientChan2:
			// Message received
		case <-time.After(1 * time.Second):
			t.Error("Client 2 didn't receive message")
		}
	}()

	// Notify
	manager.Notify(userID, message.Type, models.NotificationSeverityInfo, message.Data, message.ID)

	wg.Wait()
}

func TestSSEManager_Stop(t *testing.T) {
	manager := NewSSEManager()

	userID := 1
	manager.AddClient(userID)

	// Stop the manager
	manager.Stop()

	// Verify context is cancelled
	select {
	case <-manager.ctx.Done():
		// Context properly cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not cancelled after Stop()")
	}

	// Verify clients map is cleared
	manager.mu.RLock()
	_, exists := manager.clients[userID]
	manager.mu.RUnlock()

	if exists {
		t.Error("Clients map should be cleared after Stop()")
	}
}

func TestSSEMessage_JSONMarshal(t *testing.T) {
	message := SSEMessage{
		Type:      "test",
		Data:      map[string]string{"key": "value"},
		Severity:  "info",
		Timestamp: time.Now(),
		ID:        "test-id",
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Errorf("Failed to marshal SSEMessage: %v", err)
	}

	var unmarshaled SSEMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal SSEMessage: %v", err)
	}

	if unmarshaled.Type != message.Type {
		t.Errorf("Type mismatch: expected %s, got %s", message.Type, unmarshaled.Type)
	}
	if unmarshaled.Severity != message.Severity {
		t.Errorf("Severity mismatch: expected %s, got %s", message.Severity, unmarshaled.Severity)
	}
	if unmarshaled.ID != message.ID {
		t.Errorf("ID mismatch: expected %s, got %s", message.ID, unmarshaled.ID)
	}
}

func TestSSEManager_Notify_WithTaskID(t *testing.T) {
	manager := NewSSEManager()
	defer manager.Stop()

	userID := 1
	clientChan := manager.AddClient(userID)

	taskID := "task-123"

	go manager.Notify(userID, "test", models.NotificationSeverityInfo,
		map[string]string{"message": "test"}, "msg-id", taskID)

	select {
	case received := <-clientChan:
		if received.TaskID != taskID {
			t.Errorf("Expected TaskID %s, got %s", taskID, received.TaskID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}
