package models

import "time"

// Voucher struct holds all data for vouchers
type Voucher struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Voucher   string    `json:"voucher" gorm:"unique; not null"`
	Value     float64   `gorm:"not null"`
	Redeemed  bool      `json:"redeemed" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}
