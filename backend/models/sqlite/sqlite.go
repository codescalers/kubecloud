package sqlite

import (
	"fmt"
	"kubecloud/models"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Sqlite struct implements db interface with sqlite
type Sqlite struct {
	db    *gorm.DB
	mutex sync.Mutex
}

// NewSqliteStorage connects to the database file
func NewSqliteStorage(file string) (*Sqlite, error) {
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate models
	err = db.AutoMigrate(&models.User{}, &models.Voucher{}, models.Transaction{})
	if err != nil {
		return nil, err
	}

	return &Sqlite{db: db}, nil
}

// Close closes the database connection
func (s *Sqlite) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RegisterUser registers a new user to the system
func (s *Sqlite) RegisterUser(user *models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Create(user).Error
}

// GetUserByEmail returns user by its email if found
func (s *Sqlite) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := s.db.First(&user, "email = ?", email)
	return user, query.Error
}

// GetUserByEmail returns user by its email if found
func (s *Sqlite) GetUserByID(userID int) (models.User, error) {
	var user models.User
	query := s.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (s *Sqlite) UpdateUserByID(user *models.User) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
}

func (s *Sqlite) UpdatePassword(email string, hashedPassword []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result := s.db.Model(&models.User{}).
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

func (s *Sqlite) UpdateUserVerification(userID int, verified bool) error {
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("verified", verified)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with ID %d", userID)
	}

	return nil
}

// ListAllUsers lists all users in system
func (s *Sqlite) ListAllUsers() ([]models.User, error) {
	var users []models.User

	err := s.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil

}

// DeleteUserByID deletes user by its ID
func (s *Sqlite) DeleteUserByID(userID int) error {
	return s.db.Where("id = ?", userID).Delete(&models.User{}).Error
}

// CreateVoucher creates new voucher in system
func (s *Sqlite) CreateVoucher(voucher *models.Voucher) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Create(voucher).Error
}

// ListAllVouchers gets all vouchers in system
func (s *Sqlite) ListAllVouchers() ([]models.Voucher, error) {
	var vouchers []models.Voucher

	err := s.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

// CreateTransaction creates a payment transaction
func (s *Sqlite) CreateTransaction(transaction *models.Transaction) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Create(transaction).Error
}

// CreditUserBalance add credited balance to user by its ID
func (s *Sqlite) CreditUserBalance(userID int, amount float64) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}
