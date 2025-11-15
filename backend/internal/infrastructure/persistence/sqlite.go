package persistence

import (
	"kubecloud/internal/core/models"

	"gorm.io/driver/sqlite"
)

// NewSqliteGormDB returns a models.DB using SQLite as the backend.
func NewSqliteGormDB(file string) (models.DB, error) {
	dialector := sqlite.Open(file)
	return NewGormStorage(dialector)
}
