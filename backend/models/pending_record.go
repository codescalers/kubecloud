package models

import "time"

const (
	ChargeBalanceMode = "charge_balance"
	RedeemVoucherMode = "redeem_voucher"
	AdminCreditMode   = "admin_credit"
)

type PendingRecord struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID   int    `json:"user_id" gorm:"not null"`
	Username string `json:"username"`
	// TFTs are multiplied by 1e7
	TFTAmount            uint64    `json:"tft_amount" gorm:"not null"`
	TransferredTFTAmount uint64    `json:"transferred_tft_amount" gorm:"not null"`
	TransferMode         string    `json:"transfer_mode"`
	CreatedAt            time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"not null"`
}
