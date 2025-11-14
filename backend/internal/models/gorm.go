package models

import (
	"context"

	"gorm.io/gorm"
)

// GormDB struct implements db interface with gorm
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
		&User{},
		&Voucher{},
		Transaction{},
		Invoice{},
		NodeItem{},
		UserNodes{},
		&Notification{},
		&SSHKey{},
		&Cluster{},
		&PendingRecord{},
		&Settings{},
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
