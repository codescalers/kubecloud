package models

import "time"

// Transaction model holds all data for any transaction
type Transaction struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id"`
	AdminID   int       `json:"admin_id"`
	Amount    float64   `json:"amount"`
	Memo      string    `json:"memo"`
	CreatedAt time.Time `json:"created_at"`
}
