package models

import "time"

type operation string
type State string

const (
	WithdrawOperation operation = "withdraw"
	DepositOperation  operation = "deposit"

	FailedState  State = "failed"
	SuccessState State = "success"
	PendingState State = "pending"
)

type TransferRecord struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id" gorm:"not null"`
	Username  string    `json:"username"`
	TFTAmount uint64    `json:"tft_amount" gorm:"not null"` // TFTs are multiplied by 1e7
	Operation operation `json:"operation" gorm:"not null"`
	State     State     `json:"state" gorm:"not null;default:pending"`
	Failure   string    `json:"failure" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}
