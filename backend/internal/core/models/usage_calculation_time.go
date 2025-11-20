package models

import "time"

// UserUsageCalculationTime represents the last time a user's usage was calculated
type UserUsageCalculationTime struct {
	ID           int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID       int       `gorm:"user_id;index:idx_user_id,unique" binding:"required"`
	LastCalcTime time.Time `json:"last_calc_time"`
	UpdatedAt    time.Time `json:"updated_at"`
}
