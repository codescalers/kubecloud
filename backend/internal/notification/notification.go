package notification

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
	jobs                  chan *models.Notification
	quit                  chan struct{}
	notificationTemplates map[models.NotificationType]models.Notification
}

func NewNotificationService(db models.DB, notifiers ...Notifier) *NotificationService {
	notifiersMap := make(map[string]Notifier)
	for _, notifier := range notifiers {
		notifiersMap[notifier.GetType()] = notifier
	}

	s := &NotificationService{
		db:                    db,
		notifiers:             notifiersMap,
		jobs:                  make(chan *models.Notification, 25),
		quit:                  make(chan struct{}),
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

func (s *NotificationService) Start() {
	go func() {
		for {
			select {
			case job := <-s.jobs:
				if job == nil {
					continue
				}
				for _, channel := range job.Channels {
					n, ok := s.notifiers[channel]
					if !ok {
						continue
					}
					if err := n.Notify(job); err != nil {
						log.Error().Err(err).Str("channel", channel).Msg("Failed to notify")
					}
				}
			case <-s.quit:
				return
			}
		}
	}()
}

func (s *NotificationService) Stop() {
	close(s.quit)
}

func (s *NotificationService) Send(notificationType models.NotificationType, payload map[string]string, userID string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    notificationType,
		Payload: payload,
	}

	if err := s.db.CreateNotification(notification); err != nil {
		return err
	}

	if len(notification.Channels) == 0 {
		notification.Channels = []string{ChannelUI}
	}

	select {
	case s.jobs <- notification:
		log.Info().Str("notification_id", notification.ID.String()).Msg("Notification enqueued")
	default:
		log.Error().Msg("Notification queue is full")
	}

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
