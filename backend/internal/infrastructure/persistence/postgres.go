package persistence

import (
	"time"

	"kubecloud/internal/core/models"

	"gorm.io/driver/postgres"
)

// NewPostgresGormDB returns a models.DB using Postgres as the backend (with AutoMigrate)
func NewPostgresGormDB(dsn string, cfg models.DBPoolConfig) (models.DB, error) {
	dialector := postgres.Open(dsn)
	storage, err := NewGormStorage(dialector)
	if err != nil {
		return nil, err
	}
	ConfigureSQLPool(storage, cfg)
	return storage, nil
}

func ConfigureSQLPool(storage models.DB, cfg models.DBPoolConfig) {
	if storage == nil {
		return
	}
	gdb := storage.GetDB()
	if gdb == nil {
		return
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns >= 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeMinutes > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)
	}
	if cfg.ConnMaxIdleTimeMinutes > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTimeMinutes) * time.Minute)
	}
}
