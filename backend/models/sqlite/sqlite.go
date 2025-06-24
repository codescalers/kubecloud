package sqlite

import (
	"kubecloud/models"
	"sync"

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
	err = db.AutoMigrate(&models.User{})
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
