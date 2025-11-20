package models

import (
	"time"
)

// UserContractData model holds info of contracts of user
type UserContractData struct {
	ID         int          `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int          `gorm:"user_id" binding:"required"`
	ContractID uint64       `gorm:"contract_id" binding:"required"`
	NodeID     uint32       `gorm:"node_id" binding:"required"`
	Type       ContractType `gorm:"type" binding:"required"`
	CreatedAt  time.Time    `json:"created_at"`
	DeletedAt  time.Time    `json:"deleted_at"`
}
