package notification

import (
	"context"
	"fmt"
	"kubecloud/models"
	"sync"

	"github.com/google/uuid"
	"github.com/xmonader/ewf"
)

const (
	ChannelUI    = "ui"
	ChannelEmail = "email"
)

type Notifier interface {
	Notify(notification *models.Notification) error
	GetType() string
}

type NotificationServiceInterface interface {
	Send(notificationType string, payload any, userID string) error
	GetNotifiers()map[string]Notifier
	GetUserNotifications(userID string, limit, offset int) ([]models.Notification, error)
	MarkAsRead(notificationID string) error
	DeleteNotification(notificationID uuid.UUID, userID string) error
	DeleteAllNotifications(userID string) error
	MarkAllNotificationsAsRead(userID string) error
	MarkNotificationAsUnread(notificationID uuid.UUID, userID string) error
	GetUnreadNotifications(userID string, limit, offset int) ([]models.Notification, error)
	RegisterTemplate(notificationType models.NotificationType, severity models.NotificationSeverity, notifiers []string)
}

type NotificationService struct {
	db                    models.DB
	notifiers             map[string]Notifier
	engine                *ewf.Engine
	notificationTemplates map[models.NotificationType]models.Notification
}

var (
	notificationServiceOnce     sync.Once
	notificationServiceInstance *NotificationService
)


func InitNotificationService(db models.DB, engine *ewf.Engine, notifiers ...Notifier) *NotificationService {
	notificationServiceOnce.Do(func() {
		notificationServiceInstance = NewNotificationService(db, engine, notifiers...)
	})
	return notificationServiceInstance
}

func GetNotificationService() (*NotificationService, error) {
	if notificationServiceInstance == nil {
		return nil, fmt.Errorf("notification service is not initialized; call InitNotificationService first")
	}
	return notificationServiceInstance, nil
}

func(s *NotificationService) GetNotifiers()map[string]Notifier{
	return s.notifiers
}

func NewNotificationService(db models.DB, engine *ewf.Engine, notifiers ...Notifier) *NotificationService {
	notifiersMap := make(map[string]Notifier)
	for _, notifier := range notifiers {
		notifiersMap[notifier.GetType()] = notifier
	}

	s := &NotificationService{
		db:                    db,
		notifiers:             notifiersMap,
		engine:                engine,
		notificationTemplates: make(map[models.NotificationType]models.Notification),
	}
	return s
}

func (s *NotificationService) RegisterTemplate(notificationType models.NotificationType, severity models.NotificationSeverity, notifiers []string) {
	s.notificationTemplates[notificationType] = models.Notification{
		Type:     notificationType,
		Severity: severity,
		Channels: notifiers,
	}

	return s, nil
}

func (s *NotificationService) Send(notificationType models.NotificationType, payload map[string]string, userID string) error {
	notificationTemplate, ok := s.notificationTemplates[notificationType]
	if !ok {
		return fmt.Errorf("notification template not found for type: %s", notificationType)
	}

	notification := &models.Notification{
		ID:       uuid.NewString(),
		UserID:   userID,
		Type:     notificationType,
		Channels: notificationTemplate.Channels,
		Severity: notificationTemplate.Severity,
		Payload:  payload,
	}

	if err := s.db.CreateNotification(notification); err != nil {
		return err
	}

	workflow, err := s.engine.NewWorkflow("send-notification")
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}
	workflow.State["notification"] = notification
	s.engine.RunAsync(context.Background(), workflow)

	return nil
}

func (s *NotificationService) MarkNotificationAsRead(userID string, notificationID uuid.UUID) error {
	return s.db.MarkNotificationAsRead(notificationID, userID)
}

func (s *NotificationService) GetUserNotifications(userID string, limit, offset int) ([]models.Notification, error) {
	return s.db.GetUserNotifications(userID, limit, offset)
}

func (s *NotificationService) MarkAllNotificationsAsRead(userID string) error {
	return s.db.MarkAllNotificationsAsRead(userID)
}

func (s *NotificationService) MarkNotificationAsUnread(userID string, notificationID uuid.UUID) error {
	return s.db.MarkNotificationAsUnread(notificationID, userID)
}

func (s *NotificationService) GetUnreadNotifications(userID string, limit, offset int) ([]models.Notification, error) {
	return s.db.GetUnreadNotifications(userID, limit, offset)
}

func (s *NotificationService) DeleteNotification(userID string, notificationID uuid.UUID) error {
	return s.db.DeleteNotification(notificationID, userID)
}

func (s *NotificationService) DeleteAllNotifications(userID string) error {
	return s.db.DeleteAllNotifications(userID)
}
