package models

import (
	"time"

	"gorm.io/gorm"
)

type ContractType string

const (
	ContractTypeRented   ContractType = "rented"
	ContractTypeDeployed ContractType = "deployed"
)

// User represents a user in the system
type User struct {
	ID                int            `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	StripeCustomerID  string         `json:"stripe_customer_id"`
	Username          string         `json:"username" binding:"required"`
	Email             string         `json:"email" gorm:"column:email;uniqueIndex:idx_users_email,where:deleted_at IS NULL" binding:"required"`
	Password          []byte         `json:"-" binding:"required"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	Verified          bool           `json:"verified"`
	Code              int            `json:"code"`
	Admin             bool           `json:"admin"`
	CreditCardBalance uint64         `json:"credit_card_balance" gorm:"default:0"` // millicent, money from credit card
	CreditedBalance   uint64         `json:"credited_balance" gorm:"default:0"`    // millicent, manually added by admin or from vouchers
	Mnemonic          string         `json:"-" gorm:"column:mnemonic"`
	SSHKey            string         `json:"ssh_key"`
	Debt              uint64         `json:"debt"` // millicent
	Sponsored         bool           `json:"sponsored"`
	AccountAddress    string         `json:"account_address" gorm:"column:account_address"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// SSHKey represents an SSH key for a user
type SSHKey struct {
	ID        int            `gorm:"primaryKey;autoIncrement;column:id"`                                                                                                   // Primary key
	UserID    int            `gorm:"user_id;index:idx_user_name,unique,where:deleted_at IS NULL;index:idx_user_pubkey,unique,where:deleted_at IS NULL" binding:"required"` // User owner
	Name      string         `json:"name" binding:"required" gorm:"index:idx_user_name,unique,where:deleted_at IS NULL"`                                                   // Unique name per user
	PublicKey string         `json:"public_key" binding:"required" gorm:"index:idx_user_pubkey,unique,where:deleted_at IS NULL"`                                           // Unique public key per user
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
