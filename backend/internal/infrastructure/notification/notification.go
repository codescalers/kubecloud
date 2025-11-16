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
	Notify(notification models.Notification, receiver ...string) error
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
	engine    *ewf.Engine
	templates map[models.NotificationType]NotificationTemplate
	mu        sync.RWMutex
}

func NewNotificationDispatcher(
	db models.NotificationRepository,
	userRepo models.UserRepository,
	engine *ewf.Engine,
) (*NotificationDispatcher, error) {
	s := &NotificationDispatcher{
		NotificationRepository: db,
		userRepo:               userRepo,
		notifiers:              make(map[string]Notifier),
		engine:                 engine,
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

	// Get user email for notifiers
	userEmail := ""
	if s.userRepo != nil {
		user, err := s.userRepo.GetUserByID(notification.UserID)
		if err != nil {
			logger.GetLogger().Warn().Err(err).Int("user_id", notification.UserID).Msg("failed to get user, continuing without email")
		} else {
			userEmail = user.Email
		}
	}

	// Sync UI notifications for immediate feedback
	if s.hasChannel(notification.Channels, ChannelUI) {
		if notifier, ok := s.notifiers[ChannelUI]; ok {
			if err := notifier.Notify(*notification, userEmail); err != nil {
				logger.GetLogger().Error().Err(err).Str("channel", ChannelUI).Msg("Failed to send UI notification")
			}
		}
	}

	// Use EWF for guaranteed email delivery with retry logic
	if s.hasChannel(notification.Channels, ChannelEmail) && s.engine != nil {
		// Create workflow for reliable email delivery with retries
		wf, err := s.engine.NewWorkflow("send-notification")
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to create notification workflow")
			return nil // Don't fail - just log
		}

		wf.State = ewf.State{
			"notification": notification,
		}

		if err := s.engine.RunAsync(ctx, wf); err != nil {
			logger.GetLogger().Error().Err(err).Str("workflow_id", wf.UUID).Msg("Failed to queue notification workflow")
			// Don't return error - workflow will be retried
		}
	}

	return nil
}

// hasChannel checks if a channel exists in the notification's channels list
func (s *NotificationDispatcher) hasChannel(channels []string, channelType string) bool {
	for _, ch := range channels {
		if ch == channelType {
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
