package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	maintenanceModeEnabled  = "true"
	maintenanceModeDisabled = "false"
)

// Settings represents a key-value store for system-wide configuration
type Settings struct {
	Name      string    `gorm:"primaryKey;type:text" json:"name"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type GormSettingsRepository struct {
	db *gorm.DB
}

func NewGormSettingsRepository(db DB) *GormSettingsRepository {
	return &GormSettingsRepository{db: db.GetDB()}
}

// GetSetting retrieves a setting value by name
func (r *GormSettingsRepository) GetSetting(name string) (string, error) {
	var setting Settings
	err := r.db.Where("name = ?", name).First(&setting).Error
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

// SetSetting sets a setting value (creates or updates)
func (r *GormSettingsRepository) SetSetting(name, value string) error {
	setting := Settings{
		Name:  name,
		Value: value,
	}

	return r.db.Save(&setting).Error
}

// SetMaintenanceMode sets the maintenance mode
func (r *GormSettingsRepository) SetMaintenanceMode(enabled bool) error {
	value := maintenanceModeDisabled
	if enabled {
		value = maintenanceModeEnabled
	}
	return r.SetSetting("maintenance_mode", value)
}

// GetMaintenanceMode gets the current maintenance mode status
func (r *GormSettingsRepository) GetMaintenanceMode() (bool, error) {
	value, err := r.GetSetting("maintenance_mode")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if value == "" {
		return false, nil
	}

	return value == maintenanceModeEnabled, nil
}
