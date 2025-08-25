//go:build example

package tests

import (
	"testing"
)

func TestClient_GetAllNotifications(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	// Test getting all notifications without pagination
	notifications, err := client.GetAllNotifications(0, 0)
	if err != nil {
		t.Errorf("Failed to get all notifications: %v", err)
		return
	}

	t.Logf("Retrieved %d notifications (count: %d)", len(notifications.Notifications), notifications.Count)

	// Print details of each notification
	for i, notification := range notifications.Notifications {
		t.Logf("Notification %d:", i+1)
		t.Logf("  ID: %d", notification.ID)
		t.Logf("  Type: %s", notification.Type)
		t.Logf("  Title: %s", notification.Title)
		t.Logf("  Message: %s", notification.Message)
		t.Logf("  Status: %s", notification.Status)
		t.Logf("  Created At: %s", notification.CreatedAt)
		if notification.ReadAt != nil {
			t.Logf("  Read At: %s", *notification.ReadAt)
		}
		if notification.TaskID != "" {
			t.Logf("  Task ID: %s", notification.TaskID)
		}
		t.Log("  ---")
	}
}

func TestClient_GetUnreadNotifications(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	// Get unread notifications
	notifications, err := client.GetUnreadNotifications(10, 0)
	if err != nil {
		t.Errorf("Failed to get unread notifications: %v", err)
		return
	}

	t.Logf("Retrieved %d unread notifications (count: %d)", len(notifications.Notifications), notifications.Count)

	// Print details of each notification
	for i, notification := range notifications.Notifications {
		t.Logf("Notification %d:", i+1)
		t.Logf("  ID: %d", notification.ID)
		t.Logf("  Type: %s", notification.Type)
		t.Logf("  Title: %s", notification.Title)
		t.Logf("  Message: %s", notification.Message)
		t.Logf("  Status: %s", notification.Status)
		t.Logf("  Created At: %s", notification.CreatedAt)
		if notification.ReadAt != nil {
			t.Logf("  Read At: %s", *notification.ReadAt)
		}
		if notification.TaskID != "" {
			t.Logf("  Task ID: %s", notification.TaskID)
		}
		t.Log("  ---")
	}
}

func TestClient_MarkNotificationRead(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	err = client.MarkNotificationRead(1)
	if err != nil {
		t.Errorf("Failed to mark notification as read: %v", err)
		return
	}

	t.Logf("Successfully marked notification ID %d as read", 1)
}

func TestClient_MarkNotificationUnread(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	err = client.MarkNotificationUnread(1)
	if err != nil {
		t.Errorf("Failed to mark notification as unread: %v", err)
		return
	}

	t.Logf("Successfully marked notification ID %d as unread", 1)
}

func TestClient_MarkAllNotificationsRead(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	// Get count of unread notifications before marking all as read
	unreadBefore, err := client.GetUnreadNotifications(0, 0)
	if err != nil {
		t.Errorf("Failed to get unread notifications before test: %v", err)
		return
	}

	t.Logf("Found %d unread notifications before marking all as read", len(unreadBefore.Notifications))

	// Mark all notifications as read
	err = client.MarkAllNotificationsRead()
	if err != nil {
		t.Errorf("Failed to mark all notifications as read: %v", err)
		return
	}

	t.Log("Successfully marked all notifications as read")

	// Verify no unread notifications remain
	unreadAfter, err := client.GetUnreadNotifications(0, 0)
	if err != nil {
		t.Errorf("Failed to get unread notifications after test: %v", err)
		return
	}

	if len(unreadAfter.Notifications) != 0 {
		t.Errorf("Expected 0 unread notifications after marking all as read, got %d",
			len(unreadAfter.Notifications))
	}

	t.Logf("Verified: %d unread notifications remaining (should be 0)", len(unreadAfter.Notifications))
}

func TestClient_DeleteNotification(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	notificationID := uint(1)
	err = client.DeleteNotification(notificationID)
	if err != nil {
		t.Errorf("Failed to delete notification: %v", err)
		return
	}

	t.Logf("Successfully deleted notification ID %d", notificationID)
}

func TestClient_DeleteAllNotifications(t *testing.T) {
	client := NewClient()

	// Login first
	err := client.Login(testEmail, testPassword)
	if err != nil {
		t.Errorf("Login failed: %v", err)
		return
	}
	t.Log("Login successful")

	// Get count of notifications before deleting all
	allBefore, err := client.GetAllNotifications(0, 0)
	if err != nil {
		t.Errorf("Failed to get notifications before deletion: %v", err)
		return
	}

	t.Logf("Found %d notifications before deleting all", len(allBefore.Notifications))

	// Delete all notifications
	err = client.DeleteAllNotifications()
	if err != nil {
		t.Errorf("Failed to delete all notifications: %v", err)
		return
	}

	t.Log("Successfully deleted all notifications")

	// Verify no notifications remain
	allAfter, err := client.GetAllNotifications(0, 0)
	if err != nil {
		t.Errorf("Failed to get notifications after deletion: %v", err)
		return
	}

	if len(allAfter.Notifications) != 0 {
		t.Errorf("Expected 0 notifications after deleting all, got %d", len(allAfter.Notifications))
	}

	t.Logf("Verified: %d notifications remaining (should be 0)", len(allAfter.Notifications))
}
