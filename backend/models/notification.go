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

// CreateNotification creates a new notification
func (s *GormDB) CreateNotification(notification *Notification) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Create(notification).Error
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *GormDB) GetUserNotifications(userID int, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkNotificationAsRead marks a specific notification as read
func (s *GormDB) MarkNotificationAsRead(notificationID string, userID int) error {
	now := time.Now()
	result := s.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  NotificationStatusRead,
			"read_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func (s *GormDB) MarkAllNotificationsAsRead(userID int) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  NotificationStatusRead,
			"read_at": &now,
		}).Error
}

// DeleteNotification deletes a notification for a user
func (s *GormDB) DeleteNotification(notificationID string, userID int) error {
	result := s.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&Notification{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetUnreadNotifications retrieves only unread notifications for a user with pagination
func (s *GormDB) GetUnreadNotifications(userID int, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := s.db.Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// DeleteAllNotifications deletes all notifications for a user
func (s *GormDB) DeleteAllNotifications(userID int) error {
	return s.db.Where("user_id = ?", userID).Delete(&Notification{}).Error
}

// MarkNotificationAsUnread marks a specific notification as unread
func (s *GormDB) MarkNotificationAsUnread(notificationID string, userID int) error {
	result := s.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  NotificationStatusUnread,
			"read_at": nil,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
