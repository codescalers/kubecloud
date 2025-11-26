package notification

import (
	"bytes"
	"context"
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/logger"
	mailservice "kubecloud/internal/infrastructure/mailservice"
	mailcontentformatter "kubecloud/internal/infrastructure/mailservice/mail_content_formatter"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/xmonader/ewf"
)

type EmailNotifier struct {
	mailService mailservice.MailService
	userRepo    models.UserRepository
	engine      *ewf.Engine
}

func NewEmailNotifier(mailService mailservice.MailService, userRepo models.UserRepository) *EmailNotifier {
	return &EmailNotifier{
		mailService: mailService,
		userRepo:    userRepo,
	}
}

// SetEWFEngine injects the EWF engine for async delivery with retries
func (n *EmailNotifier) SetEWFEngine(engine *ewf.Engine) {
	n.engine = engine
}

func (n *EmailNotifier) GetType() string {
	return ChannelEmail
}

// Notify sends email notification using async delivery via EWF if available, or direct send as fallback
func (n *EmailNotifier) Notify(notification *models.Notification) error {
	// Fetch user email from repository
	if n.userRepo == nil {
		return fmt.Errorf("user repository not available")
	}

	user, err := n.userRepo.GetUserByID(notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user email for notification: %w", err)
	}
	if user.Email == "" {
		return fmt.Errorf("user %d has no email address", notification.UserID)
	}

	receiverEmail := user.Email

	// If EWF engine available, queue for guaranteed delivery with retries
	if n.engine != nil {
		return n.queueEmailViaEWF(context.Background(), *notification, receiverEmail)
	}

	// Fallback: send directly (used during testing or if EWF unavailable)
	return n.sendEmailDirect(*notification, receiverEmail)
}

// queueEmailViaEWF uses the EWF workflow for guaranteed email delivery with retries
func (n *EmailNotifier) queueEmailViaEWF(ctx context.Context, notification models.Notification, receiverEmail string) error {
	wf, err := n.engine.NewWorkflow("send-email-notification")
	if err != nil {
		logger.GetLogger().Error().
			Err(err).
			Int("user_id", notification.UserID).
			Str("email", receiverEmail).
			Msg("failed to create email notification workflow")
		return fmt.Errorf("failed to create email workflow: %w", err)
	}

	wf.State = ewf.State{
		"notification": notification,
	}

	if err := n.engine.Run(ctx, wf, ewf.WithAsync()); err != nil {
		logger.GetLogger().Error().
			Err(err).
			Str("workflow_id", wf.UUID).
			Int("user_id", notification.UserID).
			Str("email", receiverEmail).
			Msg("failed to queue email notification workflow")
		return fmt.Errorf("failed to queue email workflow: %w", err)
	}

	logger.GetLogger().Debug().
		Str("workflow_id", wf.UUID).
		Int("user_id", notification.UserID).
		Str("email", receiverEmail).
		Msg("email notification queued for delivery via EWF")

	return nil
}

// sendEmailDirect sends the email immediately (fallback when EWF unavailable)
func (n *EmailNotifier) sendEmailDirect(notification models.Notification, receiverEmail string) error {
	from := mail.NewEmail("MyceliumCloud", n.mailService.GetMailConfig().Email)
	receiver := mail.NewEmail("MyceliumCloud User", receiverEmail)

	tplName := string(notification.Type)

	var buf bytes.Buffer
	emailTpls := mailcontentformatter.GetNotificationEmailTemplates()
	if err := emailTpls.ExecuteTemplate(&buf, tplName, notification); err != nil {
		return fmt.Errorf("failed to execute notification template '%s': %w", tplName, err)
	}

	subject := notification.Payload["subject"]
	if subject == "" {
		subject = string(notification.Type) + " Notification"
	}

	message := mail.NewSingleEmail(from, subject, receiver, "", buf.String())
	message.Content = []*mail.Content{
		mail.NewContent("text/html", buf.String()),
	}

	err := n.mailService.SendEmailNotification(receiverEmail, subject, buf.String())
	return err
}
