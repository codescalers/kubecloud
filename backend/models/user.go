package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                int       `gorm:"primaryKey;autoIncrement;column:id"`
	StripeCustomerID  string    `json:"stripe_customer_id"`
	Username          string    `json:"username" binding:"required"`
	Email             string    `json:"email" gorm:"unique" binding:"required"`
	Password          []byte    `json:"password" binding:"required"`
	UpdatedAt         time.Time `json:"updated_at"`
	Verified          bool      `json:"verified"`
	Code              int       `json:"code"`
	Admin             bool      `json:"admin"`
	CreditCardBalance uint64    `json:"credit_card_balance" gorm:"default:0"` // millicent, money from credit card
	CreditedBalance   uint64    `json:"credited_balance" gorm:"default:0"`    // millicent, manually added by admin or from vouchers
	Mnemonic          string    `json:"-" gorm:"column:mnemonic"`
	SSHKey            string    `json:"ssh_key"`
	Debt              uint64    `json:"debt"` // millicent
	Sponsored         bool      `json:"sponsored"`
	AccountAddress    string    `json:"account_address" gorm:"column:account_address"`
}

// SSHKey represents an SSH key for a user
type SSHKey struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id"`                                                 // Primary key
	UserID    int       `gorm:"user_id;index:idx_user_name,unique;index:idx_user_pubkey,unique" binding:"required"` // User owner
	Name      string    `json:"name" binding:"required" gorm:"index:idx_user_name,unique"`                          // Unique name per user
	PublicKey string    `json:"public_key" binding:"required" gorm:"index:idx_user_pubkey,unique"`                  // Unique public key per user
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db DB) *GormUserRepository {
	return &GormUserRepository{db: db.GetDB()}
}

// RegisterUser registers a new user to the system
func (r *GormUserRepository) RegisterUser(user *User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail returns user by its email if found
func (r *GormUserRepository) GetUserByEmail(email string) (User, error) {
	var user User
	query := r.db.First(&user, "email = ?", email)
	return user, query.Error
}

// GetUserByEmail returns user by its email if found
func (r *GormUserRepository) GetUserByID(userID int) (User, error) {
	var user User
	query := r.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (r *GormUserRepository) UpdateUserByID(user *User) error {
	user.UpdatedAt = time.Now()
	return r.db.Model(&User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
}

// UpdatePassword updates password of user by its email
func (r *GormUserRepository) UpdatePassword(email string, hashedPassword []byte) error {
	result := r.db.Model(&User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with email %s", email)
	}

	return nil
}

// ListAllUsers lists all users in system
func (r *GormUserRepository) ListAllUsers() ([]User, error) {
	var users []User

	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListAdmins gets all admins
func (r *GormUserRepository) ListAdmins() ([]User, error) {
	var admins []User
	return admins, r.db.Where("admin = true and verified = true").Find(&admins).Error
}

// DeleteUserByID deletes user by its ID
func (r *GormUserRepository) DeleteUserByID(userID int) error {
	result := r.db.Where("id = ?", userID).Delete(&User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreditUserBalance add credited balance to user by its ID
func (r *GormUserRepository) CreditUserBalance(userID int, amount uint64) error {
	return r.db.Model(&User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}

// CountAllUsers returns the total number of users in the system
func (r *GormUserRepository) CountAllUsers() (int64, error) {
	var count int64
	err := r.db.Model(&User{}).Count(&count).Error
	return count, err
}

// CreateNotification creates a new notification
// CreateSSHKey creates a new SSH key for a user
func (r *GormUserRepository) CreateSSHKey(sshKey *SSHKey) error {
	sshKey.CreatedAt = time.Now()
	sshKey.UpdatedAt = time.Now()
	return r.db.Create(sshKey).Error
}

// ListUserSSHKeys returns all SSH keys for a user
func (r *GormUserRepository) ListUserSSHKeys(userID int) ([]SSHKey, error) {
	var sshKeys []SSHKey
	err := r.db.Where("user_id = ?", userID).Find(&sshKeys).Error
	if err != nil {
		return nil, err
	}
	return sshKeys, nil
}

// DeleteSSHKey deletes an SSH key by ID for a specific user
func (r *GormUserRepository) DeleteSSHKey(sshKeyID int, userID int) error {
	result := r.db.Where("id = ? AND user_id = ?", sshKeyID, userID).Delete(&SSHKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no SSH key found with ID %d for user %d", sshKeyID, userID)
	}
	return nil
}

// GetSSHKeyByID returns an SSH key by ID for a specific user
func (r *GormUserRepository) GetSSHKeyByID(sshKeyID int, userID int) (SSHKey, error) {
	var sshKey SSHKey
	query := r.db.Where("id = ? AND user_id = ?", sshKeyID, userID).First(&sshKey)
	return sshKey, query.Error
}
