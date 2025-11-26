package models

import (
	"time"

	"gorm.io/gorm"
)

// UserNodes model holds info of reserved nodes of user
type UserNodes struct {
	ID         int            `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int            `gorm:"user_id" binding:"required"`
	ContractID uint64         `gorm:"contract_id" binding:"required"`
	NodeID     uint32         `gorm:"column:node_id;index:idx_user_node_id,unique,where:deleted_at IS NULL" binding:"required"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
