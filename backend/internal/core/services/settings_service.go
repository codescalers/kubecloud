package services

import (
	"kubecloud/internal/core/models"
)

type SettingsService struct {
	settingsRepo models.SettingsRepository
}

func NewSettingsService(settingsRepo models.SettingsRepository) SettingsService {
	return SettingsService{
		settingsRepo: settingsRepo,
	}
}

func (s *SettingsService) GetMaintenanceMode() (bool, error) {
	return s.settingsRepo.GetMaintenanceMode()
}

func (s *SettingsService) SetMaintenanceMode(enabled bool) error {
	return s.settingsRepo.SetMaintenanceMode(enabled)
}
