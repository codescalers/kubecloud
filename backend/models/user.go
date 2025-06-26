package models

import "time"

// User represents a user in the system
type User struct {
	ID                int       `gorm:"primaryKey;autoIncrement;column:id"`
	Username          string    `json:"username" binding:"required"`
	Email             string    `json:"email" gorm:"unique" binding:"required"`
	Password          []byte    `json:"password" binding:"required"`
	UpdatedAt         time.Time `json:"updated_at"`
	Verified          bool      `json:"verified"`
	Code              int       `json:"code"`
	Admin             bool      `json:"admin"`
	CreditCardBalance float64   `json:"credit_card_balance" gorm:"default:0"` // money from credit card
	CreditedBalance   float64   `json:"credited_balance" gorm:"default:0"`    // manually added by admin or from vouchers
	Mnemonic          string    `json:"-" gorm:"column:mnemonic"`
}
