package notification

import (
	"fmt"

	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/realtime"
)

type SSENotifier struct {
	sse *realtime.SSEManager
}

func NewSSENotifier(sse *realtime.SSEManager) *SSENotifier {
	return &SSENotifier{sse: sse}
}

func (n *SSENotifier) GetType() string {
	return ChannelUI
}

func (n *SSENotifier) GetStepName() string {
	return "send-ui-notification"
}

func (n *SSENotifier) Notify(notification models.Notification, receiver ...string) error {
	if n.sse == nil {
		return fmt.Errorf("sse manager is nil")
	}

	msgType := string(notification.Type)
	data := notification.Payload
	if data == nil {
		data = map[string]string{}
	}

	var notificationID string
	if notification.Persist {
		notificationID = notification.ID
	}

	n.sse.Notify(notification.UserID, msgType, notification.Severity, data, notificationID, notification.TaskID)
	return nil
}

func (n *SSENotifier) GetSSEManager() *realtime.SSEManager {
	return n.sse
}
