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

// GetSetting retrieves a setting value by name
func (g *GormDB) GetSetting(name string) (string, error) {
	var setting Settings
	err := g.db.Where("name = ?", name).First(&setting).Error
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

// SetSetting sets a setting value (creates or updates)
func (g *GormDB) SetSetting(name, value string) error {
	setting := Settings{
		Name:  name,
		Value: value,
	}

	return g.db.Save(&setting).Error
}

// SetMaintenanceMode sets the maintenance mode
func (g *GormDB) SetMaintenanceMode(enabled bool) error {
	value := maintenanceModeDisabled
	if enabled {
		value = maintenanceModeEnabled
	}
	return g.SetSetting("maintenance_mode", value)
}

// GetMaintenanceMode gets the current maintenance mode status
func (g *GormDB) GetMaintenanceMode() (bool, error) {
	value, err := g.GetSetting("maintenance_mode")
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
