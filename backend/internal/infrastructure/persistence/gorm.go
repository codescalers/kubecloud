package persistence

import (
	"context"

	"kubecloud/internal/core/models"

	"gorm.io/gorm"
)

// GormDB struct implements models.DB interface with gorm
type GormDB struct {
	db *gorm.DB
}

// NewGormStorage connects to the database using the given dialector
func NewGormStorage(dialector gorm.Dialector) (*GormDB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return newGormDB(db)
}

func newGormDB(db *gorm.DB) (*GormDB, error) {
	// Migrate models
	err := db.AutoMigrate(
		&models.User{},
		&models.Voucher{},
		models.Transaction{},
		models.Invoice{},
		models.NodeItem{},
		models.UserNodes{},
		&models.Notification{},
		&models.SSHKey{},
		&models.Cluster{},
		&models.PendingRecord{},
		&models.Settings{},
	)
	if err != nil {
		return nil, err
	}

	return &GormDB{db: db}, nil
}

func (s *GormDB) GetDB() *gorm.DB {
	return s.db
}

// Close closes the database connection
func (s *GormDB) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping implements the DB interface health check
func (s *GormDB) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
