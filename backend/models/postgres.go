package models

import (
	"gorm.io/driver/postgres"
)

// NewPostgresDB returns a models.DB using Postgres as the backend (with AutoMigrate)
func NewPostgresDB(dsn string) (DB, error) {
	dialector := postgres.Open(dsn)
	return NewGormStorage(dialector)
}

// NewPostgresDBNoMigrate opens the Postgres database without running any schema
func NewPostgresDBNoMigrate(dsn string) (DB, error) {
	dialector := postgres.Open(dsn)
	return NewGormStorageNoMigrate(dialector)
}
