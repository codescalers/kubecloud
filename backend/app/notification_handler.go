package app

import (
	"errors"
	"fmt"
	"kubecloud/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// Default pagination values
	DefaultNotificationLimit = 20
	MaxNotificationLimit     = 100
	DefaultOffset            = 0
)

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID        uint                      `json:"id"`
	Type      models.NotificationType   `json:"type"`
	Title     string                    `json:"title"`
	Message   string                    `json:"message"`
	Data      string                    `json:"data,omitempty"`
	TaskID    string                    `json:"task_id,omitempty"`
	Status    models.NotificationStatus `json:"status"`
	CreatedAt string                    `json:"created_at"`
	ReadAt    *string                   `json:"read_at,omitempty"`
}

// convertToNotificationResponse converts a models.Notification to NotificationResponse
func convertToNotificationResponse(notification models.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        notification.ID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Data:      notification.Data,
		TaskID:    notification.TaskID,
		Status:    notification.Status,
		CreatedAt: notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if notification.ReadAt != nil {
		readAtStr := notification.ReadAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ReadAt = &readAtStr
	}

	return resp
}

// getUserIDFromContext extracts and validates user ID from context
func getUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", fmt.Errorf("unauthorized")
	}
	return fmt.Sprint(userID), nil
}

// validatePaginationParams validates and normalizes pagination parameters
func validatePaginationParams(limitStr, offsetStr string) (int, int, error) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = DefaultNotificationLimit
	}
	if limit > MaxNotificationLimit {
		limit = MaxNotificationLimit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = DefaultOffset
	}

	return limit, offset, nil
}

// GetAllNotificationsHandler retrieves all user notifications with pagination
func (h *Handler) GetAllNotificationsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	// Parse and validate pagination parameters
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultNotificationLimit))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(DefaultOffset))

	limit, offset, err := validatePaginationParams(limitStr, offsetStr)
	if err != nil {
		Error(c, http.StatusBadRequest, "Invalid pagination parameters", err.Error())
		return
	}

	notifications, err := h.db.GetUserNotifications(userID, limit, offset)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to retrieve notifications", err.Error())
		return
	}

	// Convert to response format
	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, convertToNotificationResponse(notification))
	}

	Success(c, http.StatusOK, "Notifications retrieved successfully", gin.H{
		"notifications": response,
		"limit":         limit,
		"offset":        offset,
		"count":         len(response),
	})
}

// MarkNotificationReadHandler marks a specific notification as read
func (h *Handler) MarkNotificationReadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil || notificationID == 0 {
		Error(c, http.StatusBadRequest, "Invalid notification ID", "Notification ID must be a positive integer")
		return
	}

	err = h.db.MarkNotificationAsRead(uint(notificationID), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "Notification not found", "The notification does not exist or you don't have access to it")
			return
		}
		Error(c, http.StatusInternalServerError, "Database error", "Failed to mark notification as read")
		return
	}

	Success(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllNotificationsReadHandler marks all notifications as read for a user
func (h *Handler) MarkAllNotificationsReadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	err = h.db.MarkAllNotificationsAsRead(userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to mark notifications as read", err.Error())
		return
	}

	Success(c, http.StatusOK, "All notifications marked as read", nil)
}

// DeleteNotificationHandler deletes a specific notification
func (h *Handler) DeleteNotificationHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil || notificationID == 0 {
		Error(c, http.StatusBadRequest, "Invalid notification ID", "Notification ID must be a positive integer")
		return
	}

	err = h.db.DeleteNotification(uint(notificationID), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "Notification not found", "The notification does not exist or you don't have access to it")
			return
		}
		Error(c, http.StatusInternalServerError, "Database error", "Failed to delete notification")
		return
	}

	Success(c, http.StatusOK, "Notification deleted successfully", nil)
}

// GetUnreadNotificationsHandler retrieves only unread notifications for a user
func (h *Handler) GetUnreadNotificationsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	// Parse and validate pagination parameters
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultNotificationLimit))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(DefaultOffset))

	limit, offset, err := validatePaginationParams(limitStr, offsetStr)
	if err != nil {
		Error(c, http.StatusBadRequest, "Invalid pagination parameters", err.Error())
		return
	}

	notifications, err := h.db.GetUnreadNotifications(userID, limit, offset)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to retrieve unread notifications", err.Error())
		return
	}

	// Convert to response format
	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, convertToNotificationResponse(notification))
	}

	Success(c, http.StatusOK, "Unread notifications retrieved successfully", gin.H{
		"notifications": response,
		"limit":         limit,
		"offset":        offset,
		"count":         len(response),
	})
}

// DeleteAllNotificationsHandler deletes all notifications for a user
func (h *Handler) DeleteAllNotificationsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	err = h.db.DeleteAllNotifications(userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to delete notifications", err.Error())
		return
	}

	Success(c, http.StatusOK, "All notifications deleted successfully", nil)
}

// MarkNotificationUnreadHandler marks a specific notification as unread
func (h *Handler) MarkNotificationUnreadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	notificationIDStr := c.Param("notification_id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil || notificationID == 0 {
		Error(c, http.StatusBadRequest, "Invalid notification ID", "Notification ID must be a positive integer")
		return
	}

	err = h.db.MarkNotificationAsUnread(uint(notificationID), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "Notification not found", "The notification does not exist or you don't have access to it")
			return
		}
		Error(c, http.StatusInternalServerError, "Database error", "Failed to mark notification as unread")
		return
	}

	Success(c, http.StatusOK, "Notification marked as unread", nil)
}
