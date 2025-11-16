package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSettingsRepo struct {
	mock.Mock
}

func (m *mockSettingsRepo) GetSetting(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

func (m *mockSettingsRepo) SetSetting(name, value string) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *mockSettingsRepo) SetMaintenanceMode(enabled bool) error {
	args := m.Called(enabled)
	return args.Error(0)
}

func (m *mockSettingsRepo) GetMaintenanceMode() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// Test 1: GetMaintenanceMode - ENABLED
func TestSettingsService_GetMaintenanceMode_Enabled(t *testing.T) {
	mockSettingsRepo := new(mockSettingsRepo)
	mockSettingsRepo.On("GetMaintenanceMode").Return(true, nil)

	service := NewSettingsService(mockSettingsRepo)

	enabled, err := service.GetMaintenanceMode()

	require.NoError(t, err)
	assert.True(t, enabled)
	mockSettingsRepo.AssertCalled(t, "GetMaintenanceMode")
}

// Test 2: GetMaintenanceMode - DISABLED
func TestSettingsService_GetMaintenanceMode_Disabled(t *testing.T) {
	mockSettingsRepo := new(mockSettingsRepo)
	mockSettingsRepo.On("GetMaintenanceMode").Return(false, nil)

	service := NewSettingsService(mockSettingsRepo)

	enabled, err := service.GetMaintenanceMode()

	require.NoError(t, err)
	assert.False(t, enabled)
}

// Test 3: GetMaintenanceMode - ERROR
func TestSettingsService_GetMaintenanceMode_Error(t *testing.T) {
	mockSettingsRepo := new(mockSettingsRepo)
	mockSettingsRepo.On("GetMaintenanceMode").Return(false, assert.AnError)

	service := NewSettingsService(mockSettingsRepo)

	_, err := service.GetMaintenanceMode()

	require.Error(t, err)
}

// Test 4: SetMaintenanceMode - ENABLE SUCCESS
func TestSettingsService_SetMaintenanceMode_Enable(t *testing.T) {
	mockSettingsRepo := new(mockSettingsRepo)
	mockSettingsRepo.On("SetMaintenanceMode", true).Return(nil)

	service := NewSettingsService(mockSettingsRepo)

	err := service.SetMaintenanceMode(true)

	require.NoError(t, err)
	mockSettingsRepo.AssertCalled(t, "SetMaintenanceMode", true)
}

// Test 5: SetMaintenanceMode - DISABLE SUCCESS
func TestSettingsService_SetMaintenanceMode_Disable(t *testing.T) {
	mockSettingsRepo := new(mockSettingsRepo)
	mockSettingsRepo.On("SetMaintenanceMode", false).Return(nil)

	service := NewSettingsService(mockSettingsRepo)

	err := service.SetMaintenanceMode(false)

	require.NoError(t, err)
	mockSettingsRepo.AssertCalled(t, "SetMaintenanceMode", false)
}

// Test 6: SetMaintenanceMode - ERROR
func TestSettingsService_SetMaintenanceMode_Error(t *testing.T) {
	mockSettingsRepo := new(mockSettingsRepo)
	mockSettingsRepo.On("SetMaintenanceMode", true).Return(assert.AnError)

	service := NewSettingsService(mockSettingsRepo)

	err := service.SetMaintenanceMode(true)

	require.Error(t, err)
}
