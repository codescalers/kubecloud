package models

import (
	"gorm.io/driver/sqlite"
)

// NewSqliteGormDB returns a models.DB using SQLite as the backend.
func NewSqliteGormDB(file string) (*GormDB, error) {
	dialector := sqlite.Open(file)
	return NewGormStorage(dialector)
}
