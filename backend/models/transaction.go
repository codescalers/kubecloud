package models

import "time"

// Transaction model holds all data for any transaction
type Transaction struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id" validate:"required"`
	AdminID   int       `json:"admin_id" validate:"required"`
	Amount    float64   `json:"amount" validate:"required,gt=0"` // in USD
	Memo      string    `json:"memo" validate:"required,min=3,max=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}
