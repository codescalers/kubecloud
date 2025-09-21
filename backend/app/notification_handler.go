package app

import (
	"errors"
	"fmt"
	"kubecloud/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// Default pagination values
	DefaultNotificationLimit = 20
	MaxNotificationLimit     = 100
	DefaultOffset            = 0
)

// @Model NotificationResponse
// @Description A notification response
// @Property ID string
// @Property TaskID string
// @Property Type models.NotificationType
// @Property Severity models.NotificationSeverity
// @Property Payload map[string]string
// @Property Status models.NotificationStatus
// @Property CreatedAt string
// @Property ReadAt *string
// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID        string                      `json:"id"`
	TaskID    string                      `json:"task_id,omitempty"`
	Type      models.NotificationType     `json:"type"`
	Severity  models.NotificationSeverity `json:"severity"`
	Payload   map[string]string           `json:"payload"`
	Status    models.NotificationStatus   `json:"status"`
	CreatedAt string                      `json:"created_at"`
	ReadAt    *string                     `json:"read_at,omitempty"`
}

func (n NotificationResponse) String() string {
	base := fmt.Sprintf(`Notification{
	ID: %s,
	TaskID: %s,
	Type: %s,
	Severity: %s,
	Payload: %v,
	Status: %s,
	CreatedAt: %s`, n.ID, n.TaskID, n.Type, n.Severity, n.Payload, n.Status, n.CreatedAt)

	if n.ReadAt != nil {
		base += fmt.Sprintf(`,
	ReadAt: %s`, *n.ReadAt)
	}

	return base + "\n}"
}

// convertToNotificationResponse converts a models.Notification to NotificationResponse
func convertToNotificationResponse(notification models.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        notification.ID,
		TaskID:    notification.TaskID,
		Type:      notification.Type,
		Severity:  notification.Severity,
		Payload:   notification.Payload,
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
func getUserIDFromContext(c *gin.Context) (int, error) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		return 0, fmt.Errorf("unauthorized")
	}

	return userID, nil
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

// @Summary Get all notifications
// @Description Retrieves all user notifications with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of notifications to return (default: 20, max: 100)"
// @Param offset query int false "Number of notifications to skip (default: 0)"
// @Success 200 {object} APIResponse{data=object{notifications=[]NotificationResponse,limit=int,offset=int,count=int}} "Notifications retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid pagination parameters"
// @Failure 500 {object} APIResponse "Failed to retrieve notifications"
// @Router /notifications [get]
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

// @Summary Mark a specific notification as read
// @Description Marks a specific notification as read for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification_id path string true "Notification ID"
// @Success 200 {object} APIResponse{data=object{}} "Notification marked as read successfully"
// @Failure 400 {object} APIResponse "Invalid notification ID"
// @Failure 401 {object} APIResponse "Authentication required"
// @Failure 404 {object} APIResponse "Notification not found"
// @Failure 500 {object} APIResponse "Failed to mark notification as read"
// @Router /notifications/{notification_id}/read [patch]
// MarkNotificationReadHandler marks a specific notification as read
func (h *Handler) MarkNotificationReadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	notificationIDStr := c.Param("notification_id")
	if _, parseErr := uuid.Parse(notificationIDStr); parseErr != nil {
		Error(c, http.StatusBadRequest, "Invalid notification ID", "notification_id must be a valid UUID")
		return
	}

	err = h.db.MarkNotificationAsRead(notificationIDStr, userID)
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

// @Summary Mark all notifications as read
// @Description Marks all notifications as read for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=object{}} "All notifications marked as read successfully"
// @Failure 400 {object} APIResponse "Invalid notification ID"
// @Failure 401 {object} APIResponse "Authentication required"
// @Failure 500 {object} APIResponse "Failed to mark notifications as read"
// @Router /notifications/read-all [patch]
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

// @Summary Delete a specific notification
// @Description Deletes a specific notification for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification_id path string true "Notification ID"
// @Success 200 {object} APIResponse{data=object{}} "Notification deleted successfully"
// @Failure 400 {object} APIResponse "Invalid notification ID"
// @Failure 401 {object} APIResponse "Authentication required"
// @Failure 404 {object} APIResponse "Notification not found"
// @Failure 500 {object} APIResponse "Failed to delete notification"
// @Router /notifications/{notification_id} [delete]
// DeleteNotificationHandler deletes a specific notification
func (h *Handler) DeleteNotificationHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	notificationIDStr := c.Param("notification_id")
	if _, parseErr := uuid.Parse(notificationIDStr); parseErr != nil {
		Error(c, http.StatusBadRequest, "Invalid notification ID", "notification_id must be a valid UUID")
		return
	}

	err = h.db.DeleteNotification(notificationIDStr, userID)
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

// @Summary Get unread notifications
// @Description Retrieves only unread notifications for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} APIResponse{data=object{notifications=[]NotificationResponse,limit=int,offset=int,count=int}} "Unread notifications retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid pagination parameters"
// @Failure 500 {object} APIResponse "Failed to retrieve unread notifications"
// @Router /notifications/unread [get]
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

// @Summary Delete all notifications
// @Description Deletes all notifications for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=object{}} "All notifications deleted successfully"
// @Failure 401 {object} APIResponse "Authentication required"
// @Failure 500 {object} APIResponse "Failed to delete notifications"
// @Router /notifications [delete]
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

// @Summary Mark a specific notification as unread
// @Description Marks a specific notification as unread for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification_id path string true "Notification ID"
// @Success 200 {object} APIResponse{data=object{}} "Notification marked as unread successfully"
// @Failure 400 {object} APIResponse "Invalid notification ID"
// @Failure 401 {object} APIResponse "Authentication required"
// @Failure 404 {object} APIResponse "Notification not found"
// @Failure 500 {object} APIResponse "Failed to mark notification as unread"
// @Router /notifications/{notification_id}/unread [patch]
// MarkNotificationUnreadHandler marks a specific notification as unread
func (h *Handler) MarkNotificationUnreadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Authentication required", err.Error())
		return
	}

	notificationIDStr := c.Param("notification_id")
	if _, parseErr := uuid.Parse(notificationIDStr); parseErr != nil {
		Error(c, http.StatusBadRequest, "Invalid notification ID", "notification_id must be a valid UUID")
		return
	}

	err = h.db.MarkNotificationAsUnread(notificationIDStr, userID)
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
