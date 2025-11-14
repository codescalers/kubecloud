package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"kubecloud/internal/core/models"
)

type mockNotificationRepo struct {
	mock.Mock
}

func (m *mockNotificationRepo) CreateNotification(notification *models.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *mockNotificationRepo) GetUserNotifications(userID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *mockNotificationRepo) GetUnreadNotifications(userID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *mockNotificationRepo) MarkNotificationAsRead(notificationID string, userID int) error {
	args := m.Called(notificationID, userID)
	return args.Error(0)
}

func (m *mockNotificationRepo) MarkNotificationAsUnread(notificationID string, userID int) error {
	args := m.Called(notificationID, userID)
	return args.Error(0)
}

func (m *mockNotificationRepo) MarkAllNotificationsAsRead(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockNotificationRepo) DeleteNotification(notificationID string, userID int) error {
	args := m.Called(notificationID, userID)
	return args.Error(0)
}

func (m *mockNotificationRepo) DeleteAllNotifications(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

// Test 1: GetUserNotifications - SUCCESS
func TestNotificationService_GetUserNotifications_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	notifications := []models.Notification{
		{
			ID:     "notif1",
			UserID: 1,
		},
		{
			ID:     "notif2",
			UserID: 1,
		},
	}

	mockNotificationRepo.On("GetUserNotifications", 1, 10, 0).Return(notifications, nil)

	service := NewNotificationService(mockNotificationRepo)

	result, err := service.GetUserNotifications(1, 10, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "notif1", result[0].ID)
	mockNotificationRepo.AssertCalled(t, "GetUserNotifications", 1, 10, 0)
}

// Test 2: GetUserNotifications - NO RESULTS
func TestNotificationService_GetUserNotifications_Empty(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("GetUserNotifications", 999, 10, 0).Return([]models.Notification{}, nil)

	service := NewNotificationService(mockNotificationRepo)

	result, err := service.GetUserNotifications(999, 10, 0)

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

// Test 3: GetUnreadNotifications - SUCCESS
func TestNotificationService_GetUnreadNotifications_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	unreadNotifications := []models.Notification{
		{
			ID:     "unread1",
			UserID: 1,
		},
	}

	mockNotificationRepo.On("GetUnreadNotifications", 1, 5, 0).Return(unreadNotifications, nil)

	service := NewNotificationService(mockNotificationRepo)

	result, err := service.GetUnreadNotifications(1, 5, 0)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "unread1", result[0].ID)
}

// Test 4: MarkNotificationAsRead - SUCCESS
func TestNotificationService_MarkNotificationAsRead_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("MarkNotificationAsRead", "notif123", 1).Return(nil)

	service := NewNotificationService(mockNotificationRepo)

	err := service.MarkNotificationAsRead("notif123", 1)

	require.NoError(t, err)
	mockNotificationRepo.AssertCalled(t, "MarkNotificationAsRead", "notif123", 1)
}

// Test 5: MarkNotificationAsRead - ERROR
func TestNotificationService_MarkNotificationAsRead_Error(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("MarkNotificationAsRead", "notif123", 1).
		Return(fmt.Errorf("notification not found"))

	service := NewNotificationService(mockNotificationRepo)

	err := service.MarkNotificationAsRead("notif123", 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "notification not found")
}

// Test 6: MarkNotificationAsUnread - SUCCESS
func TestNotificationService_MarkNotificationAsUnread_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("MarkNotificationAsUnread", "notif123", 1).Return(nil)

	service := NewNotificationService(mockNotificationRepo)

	err := service.MarkNotificationAsUnread("notif123", 1)

	require.NoError(t, err)
	mockNotificationRepo.AssertCalled(t, "MarkNotificationAsUnread", "notif123", 1)
}

// Test 7: MarkAllNotificationsAsRead - SUCCESS
func TestNotificationService_MarkAllNotificationsAsRead_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("MarkAllNotificationsAsRead", 1).Return(nil)

	service := NewNotificationService(mockNotificationRepo)

	err := service.MarkAllNotificationsAsRead(1)

	require.NoError(t, err)
	mockNotificationRepo.AssertCalled(t, "MarkAllNotificationsAsRead", 1)
}

// Test 8: DeleteNotification - SUCCESS
func TestNotificationService_DeleteNotification_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("DeleteNotification", "notif123", 1).Return(nil)

	service := NewNotificationService(mockNotificationRepo)

	err := service.DeleteNotification("notif123", 1)

	require.NoError(t, err)
	mockNotificationRepo.AssertCalled(t, "DeleteNotification", "notif123", 1)
}

// Test 9: DeleteNotification - ERROR
func TestNotificationService_DeleteNotification_Error(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("DeleteNotification", "notif123", 1).
		Return(fmt.Errorf("not authorized"))

	service := NewNotificationService(mockNotificationRepo)

	err := service.DeleteNotification("notif123", 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authorized")
}

// Test 10: DeleteAllNotifications - SUCCESS
func TestNotificationService_DeleteAllNotifications_Success(t *testing.T) {
	mockNotificationRepo := new(mockNotificationRepo)

	mockNotificationRepo.On("DeleteAllNotifications", 1).Return(nil)

	service := NewNotificationService(mockNotificationRepo)

	err := service.DeleteAllNotifications(1)

	require.NoError(t, err)
	mockNotificationRepo.AssertCalled(t, "DeleteAllNotifications", 1)
}
