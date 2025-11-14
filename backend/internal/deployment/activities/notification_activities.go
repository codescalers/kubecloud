package activities

import (
	"context"
	"fmt"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/core/notification"
	"kubecloud/internal/core/models"
	"slices"

	"github.com/xmonader/ewf"
)

func SendNotification(userRepo models.UserRepository, notifier notification.Notifier) ewf.StepFn {
	return func(ctx context.Context, wf ewf.State) error {
		log := logger.ForOperation("notification_activities", "send_notification")
		raw, ok := wf["notification"]
		if !ok {
			return fmt.Errorf("missing notification in workflow state")
		}
		notif, ok := raw.(*models.Notification)
		if !ok || notif == nil {
			return fmt.Errorf("invalid notification in workflow state")
		}
		if !slices.Contains(notif.Channels, notifier.GetType()) {
			log.Debug().
				Str("channel", notifier.GetType()).
				Msg("Step skipped, channel not in notification channels")
			return nil
		}
		user, err := userRepo.GetUserByID(notif.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user by ID (id: %v): %w", notif.UserID, err)
		}
		if err := notifier.Notify(*notif, user.Email); err != nil {
			return fmt.Errorf("failed to send notification (id: %v) to %s: %w", notif.ID, notifier.GetType(), err)
		}
		return nil
	}
}
