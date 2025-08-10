package models

import "time"

// Voucher struct holds all data for vouchers, voucher used only by one user.
// Voucher struct holds all data for vouchers, voucher used only by one user.
type Voucher struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string    `json:"code" gorm:"unique;not null" validate:"required"`
	Value     float64   `json:"value" gorm:"not null" validate:"required,gt=0"`
	Redeemed  bool      `json:"redeemed" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	ExpiresAt time.Time `json:"expires_at" validate:"required,gtfield=CreatedAt"`
}
