package notification

import (
	"fmt"
	"time"

	"kubecloud/internal/core/models"
)

// NotificationBuilder provides a fluent API for building notifications
type NotificationBuilder struct {
	userID    int
	notifType models.NotificationType
	status    string
	message   string
	subject   string
	severity  models.NotificationSeverity
	channels  []string
	persist   bool
	taskID    string
	extras    map[string]string
}

// NewNotification creates a new notification builder
func NewNotification(userID int, notificationType models.NotificationType) *NotificationBuilder {
	return &NotificationBuilder{
		userID:    userID,
		notifType: notificationType,
		severity:  models.NotificationSeverityInfo,
		channels:  []string{ChannelUI},
		persist:   true,
		extras:    make(map[string]string),
	}
}

// WithStatus sets the notification status (e.g., "succeeded", "failed", "started")
func (b *NotificationBuilder) WithStatus(status string) *NotificationBuilder {
	b.status = status
	return b
}

// WithMessage sets the user-facing message
func (b *NotificationBuilder) WithMessage(message string) *NotificationBuilder {
	b.message = message
	return b
}

// WithSubject sets the subject/title (mainly for email)
func (b *NotificationBuilder) WithSubject(subject string) *NotificationBuilder {
	b.subject = subject
	return b
}

// WithSeverity sets notification severity (info, success, warning, error)
func (b *NotificationBuilder) WithSeverity(severity models.NotificationSeverity) *NotificationBuilder {
	b.severity = severity
	return b
}

// WithChannels sets which channels to use (ui, email, sms, etc.)
// Replaces default channels
func (b *NotificationBuilder) WithChannels(channels ...string) *NotificationBuilder {
	b.channels = channels
	return b
}

// AddChannel adds a channel without replacing existing
func (b *NotificationBuilder) AddChannel(channel string) *NotificationBuilder {
	b.channels = append(b.channels, channel)
	return b
}

// WithPersist enables/disables database persistence
func (b *NotificationBuilder) WithPersist(persist bool) *NotificationBuilder {
	b.persist = persist
	return b
}

// NoPersist disables persistence (convenience method)
func (b *NotificationBuilder) NoPersist() *NotificationBuilder {
	b.persist = false
	return b
}

// WithTaskID associates notification with a workflow task
func (b *NotificationBuilder) WithTaskID(taskID string) *NotificationBuilder {
	b.taskID = taskID
	return b
}

// WithExtra adds a custom payload field
func (b *NotificationBuilder) WithExtra(key, value string) *NotificationBuilder {
	b.extras[key] = value
	return b
}

// WithExtras adds multiple custom payload fields
func (b *NotificationBuilder) WithExtras(extras map[string]string) *NotificationBuilder {
	for k, v := range extras {
		b.extras[k] = v
	}
	return b
}

// Build creates the final Notification object
func (b *NotificationBuilder) Build() *models.Notification {
	payload := make(map[string]string)

	// Add standard fields
	if b.subject != "" {
		payload["subject"] = b.subject
	}
	if b.status != "" {
		payload["status"] = b.status
	}
	if b.message != "" {
		payload["message"] = b.message
	}

	// Add extras
	for k, v := range b.extras {
		payload[k] = v
	}

	// Create options
	opts := []models.NotificationOption{
		models.WithChannels(b.channels...),
		models.WithSeverity(b.severity),
	}

	if !b.persist {
		opts = append(opts, models.WithNoPersist())
	}

	notification := models.NewNotification(
		b.userID,
		b.notifType,
		payload,
		opts...,
	)

	if b.taskID != "" {
		notification.TaskID = b.taskID
	}

	return notification
}

// Common notification builders for convenience

// ClusterNotification starts building a cluster deployment notification
func ClusterNotification(userID int, clusterName string) *NotificationBuilder {
	return NewNotification(userID, models.NotificationTypeDeployment).
		WithExtra("cluster_name", clusterName).
		WithExtra("timestamp", time.Now().Local().Format(TimestampFormat))
}

// BillingNotification starts building a billing/payment notification
func BillingNotification(userID int) *NotificationBuilder {
	return NewNotification(userID, models.NotificationTypeBilling).
		WithExtra("timestamp", time.Now().Local().Format(TimestampFormat))
}

// NodeNotification starts building a node-related notification
func NodeNotification(userID int, nodeID uint32) *NotificationBuilder {
	return NewNotification(userID, models.NotificationTypeNode).
		WithExtra("node_id", fmt.Sprintf("%d", nodeID)).
		WithExtra("timestamp", time.Now().Local().Format(TimestampFormat))
}

// UserNotification starts building a user account notification
func UserNotification(userID int) *NotificationBuilder {
	return NewNotification(userID, models.NotificationTypeUser).
		WithExtra("timestamp", time.Now().Local().Format(TimestampFormat))
}

// Success helper for successful operations
func (b *NotificationBuilder) Success(message string) *NotificationBuilder {
	return b.
		WithStatus("succeeded").
		WithSeverity(models.NotificationSeveritySuccess).
		WithMessage(message)
}

// Failure helper for failed operations
func (b *NotificationBuilder) Failure(message string, err error) *NotificationBuilder {
	payload := message
	if err != nil {
		payload = fmt.Sprintf("%s: %v", message, err)
	}
	return b.
		WithStatus("failed").
		WithSeverity(models.NotificationSeverityError).
		WithMessage(payload).
		WithExtra("error", err.Error())
}

// Info helper for informational messages
func (b *NotificationBuilder) Info(message string) *NotificationBuilder {
	return b.
		WithStatus("info").
		WithSeverity(models.NotificationSeverityInfo).
		WithMessage(message)
}

// Warning helper for warning messages
func (b *NotificationBuilder) Warning(message string) *NotificationBuilder {
	return b.
		WithStatus("warning").
		WithSeverity(models.NotificationSeverityWarning).
		WithMessage(message)
}
