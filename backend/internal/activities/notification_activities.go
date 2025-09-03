package activities

import (
	"context"
	"fmt"
	"kubecloud/internal/logger"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"slices"
	"strconv"

	"github.com/xmonader/ewf"
)

func SendNotification(db models.DB, notifier notification.Notifier) ewf.StepFn {
	return func(ctx context.Context, wf ewf.State) error {
		raw, ok := wf["notification"]
		if !ok {
			return fmt.Errorf("missing notification in workflow state")
		}
		notif, ok := raw.(*models.Notification)
		if !ok || notif == nil {
			return fmt.Errorf("invalid notification in workflow state")
		}
		if !slices.Contains(notif.Channels, notifier.GetType()) {
			logger.GetLogger().Info().Msgf("SendNotification: step skipped for channel %s (not in notification channels)", notifier.GetType())
			return nil
		}
		userID, err := strconv.Atoi(notif.UserID)
		if err != nil {
			return fmt.Errorf("invalid user ID: %v", notif.UserID)
		}
		user, err := db.GetUserByID(userID)
		if err != nil {
			return fmt.Errorf("failed to get user by ID (id: %v): %w", userID, err)
		}
		if err := notifier.Notify(*notif, user.Email); err != nil {
			return fmt.Errorf("failed to send notification (id: %v) to %s: %w", notif.ID, notifier.GetType(), err)
		}
		logger.GetLogger().Info().Msgf("Sent notification (id: %v) to %s", notif.ID, notifier.GetType())
		return nil
	}
}
