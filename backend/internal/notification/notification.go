package notification

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/constants"
	"kubecloud/models"
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
	GetStepName() string
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

type NotificationServiceInterface interface {
	Send(ctx context.Context, notification *models.Notification) error
	GetNotifiers() map[string]Notifier
	RegisterNotifier(notifier Notifier)
}

type NotificationService struct {
	db        models.DB
	notifiers map[string]Notifier
	engine    *ewf.Engine
	templates map[models.NotificationType]NotificationTemplate
	mu        sync.RWMutex
}

func NewNotificationService(db models.DB, engine *ewf.Engine, notificationConfig internal.NotificationConfig) (*NotificationService, error) {

	s := &NotificationService{
		db:        db,
		notifiers: make(map[string]Notifier),
		engine:    engine,
		templates: make(map[models.NotificationType]NotificationTemplate),
	}

	if err := s.LoadTemplatesFromConfigFile(notificationConfig); err != nil {
		return nil, fmt.Errorf("failed to load notification templates: %w", err)
	}

	return s, nil
}

// buildTemplatesFromConfig creates a templates map from notification config
func (s *NotificationService) buildTemplatesFromConfig(notificationConfig internal.NotificationConfig) map[models.NotificationType]NotificationTemplate {
	templates := make(map[models.NotificationType]NotificationTemplate)

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

		templates[notificationType] = NotificationTemplate{
			Default:  defaultRule,
			ByStatus: byStatusRules,
		}
	}

	return templates
}

func (s *NotificationService) LoadTemplatesFromConfigFile(notificationConfig internal.NotificationConfig) error {
	templates := s.buildTemplatesFromConfig(notificationConfig)

	s.mu.Lock()
	s.templates = templates
	s.mu.Unlock()
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

	workflow, err := s.engine.NewWorkflow(constants.WorkflowSendNotification)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	workflow.State["notification"] = notification
	s.engine.RunAsync(ctx, workflow)
	return nil
}

func (s *NotificationService) ReloadNotificationConfig(cfg internal.NotificationConfig) error {
	candidate := s.buildTemplatesFromConfig(cfg)

	err := s.ValidateConfigsChannelsAgainstRegistered(candidate)
	if err != nil {
		return fmt.Errorf("failed to validate notification config: %w", err)
	}

	s.mu.Lock()
	s.templates = candidate
	s.mu.Unlock()
	return nil
}

func (s *NotificationService) applyTemplateFallbacks(notification *models.Notification) {
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

func (s *NotificationService) ValidateConfigsChannelsAgainstRegistered(templates ...map[models.NotificationType]NotificationTemplate) error {
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
