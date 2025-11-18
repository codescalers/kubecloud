package services

import (
	"kubecloud/internal/core/models"
)

type NotificationService struct {
	notificationRepo models.NotificationRepository
}

func NewNotificationService(
	notificationRepo models.NotificationRepository,
) NotificationService {
	return NotificationService{
		notificationRepo: notificationRepo,
	}
}

func (svc *NotificationService) GetUserNotifications(userID, limit, offset int) ([]models.Notification, error) {
	return svc.notificationRepo.GetUserNotifications(userID, limit, offset)
}
func (svc *NotificationService) MarkNotificationAsRead(notificationIDStr string, userID int) error {
	return svc.notificationRepo.MarkNotificationAsRead(notificationIDStr, userID)
}

func (svc *NotificationService) MarkNotificationAsUnread(notificationIDStr string, userID int) error {
	return svc.notificationRepo.MarkNotificationAsUnread(notificationIDStr, userID)
}

func (svc *NotificationService) MarkAllNotificationsAsRead(userID int) error {
	return svc.notificationRepo.MarkAllNotificationsAsRead(userID)
}

func (svc *NotificationService) DeleteNotification(notificationIDStr string, userID int) error {
	return svc.notificationRepo.DeleteNotification(notificationIDStr, userID)
}

func (svc *NotificationService) DeleteAllNotifications(userID int) error {
	return svc.notificationRepo.DeleteAllNotifications(userID)
}

func (svc *NotificationService) GetUnreadNotifications(userID, limit, offset int) ([]models.Notification, error) {
	return svc.notificationRepo.GetUnreadNotifications(userID, limit, offset)
}
