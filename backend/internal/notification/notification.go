package notification

import (
	"context"
	"fmt"
	"kubecloud/models"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
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
	GetStepName() string
}

type NotificationServiceInterface interface {
	Send(ctx context.Context, notificationType models.NotificationType, payload map[string]string, userID string, taskID ...string) error
	SendWithOptions(ctx context.Context, notificationType models.NotificationType, payload map[string]string, userID string, opts *SendOptions) error
	GetNotifiers() map[string]Notifier
	RegisterNotifier(notifier Notifier)
}

type NotificationService struct {
	db                    models.DB
	notifiers             map[string]Notifier
	engine                *ewf.Engine
	notificationTemplates map[models.NotificationType]models.Notification
}

type SendOptions struct {
	WithPersist  bool
	WithChannels []string
	WithSeverity models.NotificationSeverity
	TaskID       string
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

// SendOptions holds optional behaviors for sending notifications

func (s *NotificationService) Send(ctx context.Context, notificationType models.NotificationType, payload map[string]string, userID string, taskID ...string) error {
	var opts *SendOptions
	if len(taskID) > 0 {
		opts = &SendOptions{TaskID: taskID[0], WithPersist: true}
	}
	return s.SendWithOptions(ctx, notificationType, payload, userID, opts)
}

func (s *NotificationService) SendWithOptions(ctx context.Context, notificationType models.NotificationType, payload map[string]string, userID string, opts *SendOptions) error {

	if opts == nil {
		opts = &SendOptions{WithPersist: true}
	}

	tpl, hasTpl := s.templates[notificationType]
	var rule ChannelRule
	if hasTpl {
		rule = tpl.Default
		if status, ok := payload["status"]; ok && tpl.ByStatus != nil {
			if r, exists := tpl.ByStatus[status]; exists {
				rule = r
			}
		}
	}

	finalChannels := opts.WithChannels
	if len(finalChannels) == 0 && hasTpl {
		finalChannels = rule.Channels
	}
	if len(finalChannels) == 0 {
		finalChannels = []string{ChannelUI}
	}

	finalSeverity := opts.WithSeverity
	if (finalSeverity == models.NotificationSeverity("")) && hasTpl {
		finalSeverity = rule.Severity
	}

	notification := &models.Notification{
		ID:       uuid.NewString(),
		UserID:   userID,
		Type:     notificationType,
		Channels: append([]string{}, finalChannels...),
		Severity: finalSeverity,
		Payload:  payload,
		TaskID:   opts.TaskID,
	}

	if opts.WithPersist {
		if err := s.db.CreateNotification(notification); err != nil {
			return err
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

// HandleNotificationSSE streams notifications via Server-Sent Events (SSE)
// @Summary Stream notifications
// @Description Streams user-specific notifications via Server-Sent Events (SSE). Keeps the HTTP connection open and pushes events as they occur.
// @Tags notifications
// @Produce text/event-stream
// @Success 200 {string} string "event stream"
// @Failure 500 {object} gin.H "SSE notifier not available"
// @Security UserMiddleware
// @Router /notifications/stream [get]
func (s *NotificationService) HandleNotificationSSE(c *gin.Context) {
	uiNotifier, ok := s.notifiers[ChannelUI]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE notifier not available"})
		return
	}

	sseNotifier, ok := uiNotifier.(*SSENotifier)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE notifier not available"})
		return
	}
	sseNotifier.GetSSEManager().HandleSSE(c)

}
