package workflows

import (
	"context"
	"fmt"

	"github.com/xmonader/ewf"

	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/mailservice"
)

// SendEmailNotificationStep sends an email notification from workflow state
func SendEmailNotificationStep(userRepo models.UserRepository, mailService mailservice.MailService) ewf.StepFn {
	return func(ctx context.Context, wf ewf.State) error {
		log := logger.ForOperation("notification_activities", "send_email_notification")

		// Get notification from state
		raw, ok := wf["notification"]
		if !ok {
			return fmt.Errorf("missing notification in workflow state")
		}
		notif, ok := raw.(*models.Notification)
		if !ok || notif == nil {
			return fmt.Errorf("invalid notification in workflow state")
		}

		// Fetch user to get receiver email
		user, err := userRepo.GetUserByID(notif.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user by ID (id: %v): %w", notif.UserID, err)
		}

		if user.Email == "" {
			return fmt.Errorf("user %d has no email address", notif.UserID)
		}

		receiver := user.Email

		// Send via mail service
		if err := mailService.SendEmailNotification(receiver, *notif); err != nil {
			log.Error().Err(err).Str("receiver", receiver).Msg("failed to send email")
			return fmt.Errorf("failed to send email notification to %s: %w", receiver, err)
		}
		log.Debug().Str("receiver", receiver).Str("notification_type", string(notif.Type)).Msg("email notification sent")
		return nil
	}
}
