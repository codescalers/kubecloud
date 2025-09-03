package notification

import (
	"context"
	"fmt"
	"kubecloud/models"
	"time"

	"github.com/xmonader/ewf"
)

const (
	ChannelUI    = "ui"
	ChannelEmail = "email"
)

type Notifier interface {
	Notify(notification *models.Notification) error
	GetType() string
	GetStepName() string
}

type NotificationServiceInterface interface {
	Send(ctx context.Context, notification *models.Notification) error
	GetNotifiers() map[string]Notifier
	RegisterNotifier(notifier Notifier)
}

type NotificationService struct {
	db                    models.DB
	notifiers             map[string]Notifier
	engine                *ewf.Engine
	notificationTemplates map[models.NotificationType]models.Notification
}

func NewNotificationService(db models.DB, engine *ewf.Engine, notificationConfig internal.NotificationConfig) (*NotificationService, error) {

	s := &NotificationService{
		db:        db,
		notifiers: make(map[string]Notifier),
		engine:    engine,
		templates: make(map[models.NotificationType]NotificationTemplate),
	}

	if err := s.loadTemplatesFromConfigFile(notificationConfig); err != nil {
		return nil, fmt.Errorf("failed to load notification templates: %w", err)
	}

	return s, nil
}

func (s *NotificationService) registerTemplate(notificationType models.NotificationType, template NotificationTemplate) {
	s.templates[notificationType] = template
}

func (s *NotificationService) loadTemplatesFromConfigFile(notificationConfig internal.NotificationConfig) error {
	for templateName, templateConfig := range notificationConfig.TemplateTypes {
		notificationType := models.NotificationType(templateName)

		defaultRule := ChannelRule{
			Channels: templateConfig.Default.Channels,
			Severity: models.NotificationSeverity(templateConfig.Default.Severity),
		}

		byStatusRules := make(map[string]ChannelRule)
		for status, ruleConfig := range templateConfig.ByStatus {
			byStatusRules[status] = ChannelRule{
				Channels: ruleConfig.Channels,
				Severity: models.NotificationSeverity(ruleConfig.Severity),
			}
		}

		s.registerTemplate(notificationType, NotificationTemplate{
			Default:  defaultRule,
			ByStatus: byStatusRules,
		})
	}

	return nil
}

func (s *NotificationService) RegisterNotifier(notifier Notifier) {
	s.notifiers[notifier.GetType()] = notifier
}

func (s *NotificationService) GetNotifiers() map[string]Notifier {
	return s.notifiers
}

func (s *NotificationService) Send(ctx context.Context, notification *models.Notification) error {
	s.applyTemplateFallbacks(notification)

	// Persist to database if enabled
	if notification.Persist {
		if err := s.db.CreateNotification(notification); err != nil {
			return fmt.Errorf("failed to persist notification: %w", err)
		}
	}

	workflow, err := s.engine.NewWorkflow("send-notification")
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	steps := []ewf.Step{}
	for _, channel := range notification.Channels {
		notifier, exists := s.notifiers[channel]
		if !exists {
			continue
		}

		retryPolicy := &ewf.RetryPolicy{MaxAttempts: 3, BackOff: ewf.ConstantBackoff(2 * time.Second)}
		steps = append(steps, ewf.Step{
			Name:        notifier.GetStepName(),
			RetryPolicy: retryPolicy,
		})
	}
	workflow.Steps = steps
	workflow.State["notification"] = notification
	s.engine.RunAsync(ctx, workflow)
	return nil
}

func (s *NotificationService) applyTemplateFallbacks(notification *models.Notification) {
	template, hasTemplate := s.templates[notification.Type]
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
