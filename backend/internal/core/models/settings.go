package models

import (
	"time"
)

// Settings represents a key-value store for system-wide configuration
type Settings struct {
	Name      string    `gorm:"primaryKey;type:text" json:"name"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
