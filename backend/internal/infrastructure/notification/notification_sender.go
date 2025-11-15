package notification

import (
	"context"

	"kubecloud/internal/core/models"
)

const TimestampFormat = "Mon, 02 Jan 2006 15:04"

// NotificationSender routes notifications to registered channel providers.
//
// Usage:
//
//  1. Build a notification using the convenience builders:
//     notif := notification.ClusterNotification(userID, clusterName).
//     Success("Deployment completed").
//     WithSubject("Deployment Success").
//     Build()
//
//  2. Send it through registered channels (UI, Email, etc):
//     sender.Send(ctx, notif)
//
// The notification service automatically routes to all registered and configured channels.
type NotificationSender interface {
	// Send dispatches a notification to registered channel providers
	Send(ctx context.Context, notif *models.Notification) error
}

// NewNotificationSender creates a notification sender that routes to registered channel providers
func NewNotificationSender(ctx context.Context, dispatcher NotificationDispatcherInterface) NotificationSender {
	return &notificationSenderImpl{
		ctx:     ctx,
		service: dispatcher,
	}
}

type notificationSenderImpl struct {
	ctx     context.Context
	service NotificationDispatcherInterface
}

func (ns *notificationSenderImpl) Send(ctx context.Context, notif *models.Notification) error {
	return ns.service.Send(ctx, notif)
}
