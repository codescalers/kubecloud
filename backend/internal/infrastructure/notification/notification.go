package notification

import (
	"context"
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/logger"
	"sync"

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

type ChannelRule struct {
	Channels []string
	Severity models.NotificationSeverity
}

type NotificationTemplate struct {
	Default  ChannelRule
	ByStatus map[string]ChannelRule
}

// CommonPayload represents commonly used payload fields across channels
type CommonPayload struct {
	Subject string
	Status  string
	Message string
	Error   string
}

// MergePayload combines CommonPayload with additional fields
func MergePayload(cp CommonPayload, extras map[string]string) map[string]string {
	out := make(map[string]string)
	if cp.Subject != "" {
		out["subject"] = cp.Subject
	}
	if cp.Status != "" {
		out["status"] = cp.Status
	}
	if cp.Message != "" {
		out["message"] = cp.Message
	}
	if cp.Error != "" {
		out["error"] = cp.Error
	}
	for k, v := range extras {
		out[k] = v
	}
	return out
}

type NotificationDispatcherInterface interface {
	Send(ctx context.Context, notification *models.Notification) error
	GetNotifiers() map[string]Notifier
	RegisterNotifier(notifier Notifier)
}

type NotificationDispatcher struct {
	models.NotificationRepository
	userRepo  models.UserRepository
	notifiers map[string]Notifier
	templates map[models.NotificationType]NotificationTemplate
	mu        sync.RWMutex
}

func NewNotificationDispatcher(
	db models.NotificationRepository,
	userRepo models.UserRepository,
	_ *ewf.Engine,
) (*NotificationDispatcher, error) {
	s := &NotificationDispatcher{
		NotificationRepository: db,
		userRepo:               userRepo,
		notifiers:              make(map[string]Notifier),
		templates:              make(map[models.NotificationType]NotificationTemplate),
	}

	return s, nil
}

func (s *NotificationDispatcher) RegisterNotifier(notifier Notifier) {
	s.notifiers[notifier.GetType()] = notifier
}

func (s *NotificationDispatcher) GetNotifiers() map[string]Notifier {
	return s.notifiers
}

func (s *NotificationDispatcher) Send(ctx context.Context, notification *models.Notification) error {
	s.applyTemplateFallbacks(notification)

	// Persist only Email notifications (not UI-only ephemeral messages)
	if notification.Persist && s.hasChannel(notification, ChannelEmail) {
		if err := s.CreateNotification(notification); err != nil {
			logger.GetLogger().Error().Err(err).
				Int("user_id", notification.UserID).
				Str("notification_id", notification.ID).
				Msg("failed to persist notification")
			// Continue anyway - notification will still be sent
		}
	}

	// Dispatch to all registered notifiers that are in the notification's channels
	for _, channel := range notification.Channels {
		notifier, ok := s.notifiers[channel]
		if !ok {
			logger.GetLogger().Warn().Str("channel", channel).Msg("notifier not registered for channel")
			continue
		}

		// Each notifier is responsible for its own delivery strategy
		if err := notifier.Notify(notification); err != nil {
			logger.GetLogger().Error().Err(err).Str("channel", channel).Msg("failed to send notification")
		}
	}

	return nil
}

// hasChannel checks if a notification includes a specific channel
func (s *NotificationDispatcher) hasChannel(notification *models.Notification, channel string) bool {
	for _, ch := range notification.Channels {
		if ch == channel {
			return true
		}
	}
	return false
}

func (s *NotificationDispatcher) applyTemplateFallbacks(notification *models.Notification) {
	s.mu.RLock()
	template, hasTemplate := s.templates[notification.Type]
	s.mu.RUnlock()
	if !hasTemplate {
		if len(notification.Channels) == 0 {
			notification.Channels = []string{ChannelUI}
		}
		if notification.Severity == "" {
			notification.Severity = models.NotificationSeverityInfo
		}
		return
	}

	rule := template.Default
	if status, ok := notification.Payload["status"]; ok && template.ByStatus != nil {
		if r, exists := template.ByStatus[status]; exists {
			rule = r
		}
	}

	if len(notification.Channels) == 0 {
		notification.Channels = rule.Channels
	}
	if notification.Severity == "" {
		notification.Severity = rule.Severity
	}

	if len(notification.Channels) == 0 {
		notification.Channels = []string{ChannelUI}
	}
	if notification.Severity == "" {
		notification.Severity = models.NotificationSeverityInfo
	}
}

func (s *NotificationDispatcher) ValidateConfigsChannelsAgainstRegistered(templates ...map[models.NotificationType]NotificationTemplate) error {
	if len(s.notifiers) == 0 {
		return fmt.Errorf("no notifiers registered")
	}

	// Use provided templates or fall back to current templates
	var templatesMap map[models.NotificationType]NotificationTemplate
	if len(templates) > 0 && templates[0] != nil {
		templatesMap = templates[0]
	} else {
		s.mu.RLock()
		templatesMap = s.templates
		s.mu.RUnlock()
	}

	for tName, tpl := range templatesMap {
		for _, ch := range tpl.Default.Channels {
			if _, ok := s.notifiers[ch]; !ok {
				return fmt.Errorf("channel %s in template %s is not registered", ch, tName)
			}
		}
		// Validate by_status rule channels
		if tpl.ByStatus != nil {
			for status, rule := range tpl.ByStatus {
				for _, ch := range rule.Channels {
					if _, ok := s.notifiers[ch]; !ok {
						return fmt.Errorf("channel %s in template %s (status %v) is not registered", ch, tName, status)
					}
				}
			}
		}
	}
	return nil
}
