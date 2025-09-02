package activities

import (
	"context"
	"kubecloud/internal/notification"

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
		return nil
	}
}
