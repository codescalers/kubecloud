package activities

import (
	"context"
	"kubecloud/internal/notification"

	"github.com/xmonader/ewf"
)

func SendNotification(notifiers map[string]notification.Notifier) ewf.StepFn {
	return func(ctx context.Context, wf ewf.State) error {

		return nil
	}
}
