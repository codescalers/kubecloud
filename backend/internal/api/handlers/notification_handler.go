package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"kubecloud/internal/core/models"
	"kubecloud/internal/core/services"
	"kubecloud/internal/infrastructure/logger"
)

const (
	// Default pagination values
	DefaultNotificationLimit = 20
	MaxNotificationLimit     = 100
	DefaultOffset            = 0
)

type NotificationHandler struct {
	svc services.NotificationService
}

func NewNotificationHandler(svc services.NotificationService) NotificationHandler {
	return NotificationHandler{
		svc: svc,
	}
}

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
func (h *NotificationHandler) GetAllNotificationsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}
	reqLog := requestLogger(c, "GetAllNotificationsHandler")

	// Parse and validate pagination parameters
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultNotificationLimit))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(DefaultOffset))

	limit, offset, err := validatePaginationParams(limitStr, offsetStr)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_pagination",
		})
		BadRequest(c, "Invalid pagination parameters")
		return
	}

	notifications, err := h.svc.GetUserNotifications(userID, limit, offset)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve notifications")
		auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
			"limit":  limit,
			"offset": offset,
		})
		InternalServerError(c)
		return
	}

	// Convert to response format
	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, convertToNotificationResponse(notification))
	}

	auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityInfo, map[string]any{
		"notifications_length": len(response),
		"limit":                limit,
		"offset":               offset,
	})

	OK(c, "Notifications retrieved successfully", gin.H{
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
func (h *NotificationHandler) MarkNotificationReadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}
	notificationIDStr := c.Param("notification_id")
	if _, parseErr := uuid.Parse(notificationIDStr); parseErr != nil {
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_notification_id",
		})
		BadRequest(c, "Invalid notification ID")
		return
	}

	reqLog := requestLogger(c, "MarkNotificationReadHandler")
	logWithNotification := reqLog.With().Str("notification_id", notificationIDStr).Logger()
	reqLog = &logWithNotification

	err = h.svc.MarkNotificationAsRead(notificationIDStr, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotificationNotFound) {
			auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
				"reason": "notification_not_found",
				"id":     notificationIDStr,
			})
			NotFound(c, "Notification not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to mark notification as read")
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
			"id":     notificationIDStr,
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityInfo, map[string]any{
		"id":     notificationIDStr,
		"result": "read",
	})

	OK(c, "Notification marked as read successfully", nil)
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
func (h *NotificationHandler) MarkAllNotificationsReadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}
	reqLog := requestLogger(c, "MarkAllNotificationsReadHandler")

	err = h.svc.MarkAllNotificationsAsRead(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to mark notifications as read")
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityInfo, map[string]any{
		"result": "all_marked_read",
	})

	OK(c, "All notifications marked as read successfully", nil)
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
func (h *NotificationHandler) DeleteNotificationHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}

	notificationIDStr := c.Param("notification_id")
	if _, parseErr := uuid.Parse(notificationIDStr); parseErr != nil {
		auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_notification_id",
		})
		BadRequest(c, "Invalid notification ID")
		return
	}

	reqLog := requestLogger(c, "DeleteNotificationHandler")
	logWithNotification := reqLog.With().Str("notification_id", notificationIDStr).Logger()
	reqLog = &logWithNotification

	err = h.svc.DeleteNotification(notificationIDStr, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotificationNotFound) {
			auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityWarning, map[string]any{
				"reason": "notification_not_found",
				"id":     notificationIDStr,
			})
			NotFound(c, "Notification not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to delete notification")
		auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
			"id":     notificationIDStr,
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityInfo, map[string]any{
		"id":     notificationIDStr,
		"result": "deleted",
	})

	OK(c, "Notification deleted successfully", nil)
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
func (h *NotificationHandler) GetUnreadNotificationsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}
	reqLog := requestLogger(c, "GetUnreadNotificationsHandler")

	// Parse and validate pagination parameters
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultNotificationLimit))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(DefaultOffset))

	limit, offset, err := validatePaginationParams(limitStr, offsetStr)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_pagination",
		})
		BadRequest(c, "Invalid pagination parameters")
		return
	}

	notifications, err := h.svc.GetUnreadNotifications(userID, limit, offset)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve unread notifications")
		auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	// Convert to response format
	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, convertToNotificationResponse(notification))
	}

	auditLogFromContext(c, logger.AuditActionNotificationList, logger.AuditSeverityInfo, map[string]any{
		"notifications_length": len(response),
		"scope":                "unread",
		"limit":                limit,
		"offset":               offset,
	})

	OK(c, "Unread notifications retrieved successfully", gin.H{
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
func (h *NotificationHandler) DeleteAllNotificationsHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}
	reqLog := requestLogger(c, "DeleteAllNotificationsHandler")

	err = h.svc.DeleteAllNotifications(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to delete notifications")
		auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionNotificationDelete, logger.AuditSeverityInfo, map[string]any{
		"result": "all_deleted",
	})

	OK(c, "All notifications deleted successfully", nil)
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
func (h *NotificationHandler) MarkNotificationUnreadHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
			"reason": "unauthorized_user",
		})
		Unauthorized(c, "Authentication required")
		return
	}
	notificationIDStr := c.Param("notification_id")
	if _, parseErr := uuid.Parse(notificationIDStr); parseErr != nil {
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_notification_id",
		})
		BadRequest(c, "Invalid notification ID")
		return
	}

	reqLog := requestLogger(c, "MarkNotificationUnreadHandler")
	logWithNotification := reqLog.With().Str("notification_id", notificationIDStr).Logger()
	reqLog = &logWithNotification

	err = h.svc.MarkNotificationAsUnread(notificationIDStr, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotificationNotFound) {
			auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityWarning, map[string]any{
				"reason": "notification_not_found",
				"id":     notificationIDStr,
			})
			NotFound(c, "Notification not found")
			return
		}
		reqLog.Error().Err(err).Msg("failed to mark notification as unread")
		auditLogFromContext(c, logger.AuditActionNotificationRead, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
			"id":     notificationIDStr,
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(
		c,
		logger.AuditActionNotificationMarkUnread,
		logger.AuditSeverityInfo,
		map[string]any{
			"id":     notificationIDStr,
			"result": "marked_unread",
		},
	)
	OK(c, "Notification marked as unread successfully", nil)
}
