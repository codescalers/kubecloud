package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeDeployment NotificationType = "deployment"
	NotificationTypeBilling    NotificationType = "billing"
	NotificationTypeUser       NotificationType = "user"
	NotificationTypeConnected  NotificationType = "connected"
	NotificationTypeNode       NotificationType = "node"
	NotificationTypeAdmin      NotificationType = "admin"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)

type NotificationSeverity string

const (
	NotificationSeverityInfo    NotificationSeverity = "info"
	NotificationSeverityError   NotificationSeverity = "error"
	NotificationSeverityWarning NotificationSeverity = "warning"
	NotificationSeveritySuccess NotificationSeverity = "success"
)

// Notification represents a persistent notification
type Notification struct {
	ID        string               `json:"id" gorm:"primaryKey"`
	UserID    int                  `json:"user_id" gorm:"not null;index"`
	TaskID    string               `json:"task_id,omitempty" gorm:"index"`
	Type      NotificationType     `json:"type" gorm:"not null"`
	Severity  NotificationSeverity `json:"severity" gorm:"not null;default:'info'"`
	Channels  []string             `json:"channels" gorm:"serializer:json;default:'[\"ui\"]'"`
	Payload   map[string]string    `json:"payload" gorm:"serializer:json"`
	Status    NotificationStatus   `json:"status" gorm:"default:'unread'"`
	CreatedAt time.Time            `json:"created_at" gorm:"autoCreateTime"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
	DeletedAt gorm.DeletedAt       `json:"-" gorm:"index"`

	// Non-persisted fields
	Persist bool `json:"-" gorm:"-"`
}

// NotificationOption is a functional option for configuring notifications
type NotificationOption func(*Notification)

// NewNotification creates a new notification with the given options
func NewNotification(userID int, notifType NotificationType, payload map[string]string, options ...NotificationOption) *Notification {
	n := &Notification{
		ID:       uuid.NewString(),
		UserID:   userID,
		Type:     notifType,
		Severity: "",
		Channels: []string{},
		Payload:  payload,
		Status:   NotificationStatusUnread,
		Persist:  true,
	}

	for _, option := range options {
		option(n)
	}

	return n
}

// WithTaskID associates the notification with a task
func WithTaskID(taskID string) NotificationOption {
	return func(n *Notification) {
		n.TaskID = taskID
	}
}

// WithSeverity sets the notification severity
func WithSeverity(severity NotificationSeverity) NotificationOption {
	return func(n *Notification) {
		n.Severity = severity
	}
}

// WithChannels sets the notification channels
func WithChannels(channels ...string) NotificationOption {
	return func(n *Notification) {
		n.Channels = channels
	}
}

// WithNoPersist controls whether to save the notification to database
func WithNoPersist() NotificationOption {
	return func(n *Notification) {
		n.Persist = false
	}
}

// WithPayload sets the notification payload
func WithPayload(payload map[string]string) NotificationOption {
	return func(n *Notification) {
		n.Payload = payload
	}
}
