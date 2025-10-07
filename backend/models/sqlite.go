package models

import (
	"gorm.io/driver/sqlite"
)

// NewSqliteDB returns a models.DB using SQLite as the backend.
func NewSqliteDB(file string) (DB, error) {
	dialector := sqlite.Open(file)
	return NewGormStorage(dialector)
}

// NewSqliteDBNoMigrate opens the SQLite database without running any schema
func NewSqliteDBNoMigrate(file string) (DB, error) {
	dialector := sqlite.Open(file)
	return NewGormStorageNoMigrate(dialector)
}
