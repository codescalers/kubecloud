package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction model holds all data for any transaction
type Transaction struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"user_id" validate:"required"`
	AdminID   int       `json:"admin_id" validate:"required"`
	Amount    float64   `json:"amount" validate:"required,gt=0"` // in USD
	Memo      string    `json:"memo" validate:"required,min=3,max=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}

type GormTransactionRepository struct {
	db *gorm.DB
}

func NewGormTransactionRepository(db DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db.GetDB()}
}

// CreateTransaction creates a payment transaction
func (r *GormTransactionRepository) CreateTransaction(transaction *Transaction) error {
	return r.db.Create(transaction).Error
}
