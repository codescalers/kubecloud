package models

import (
	"time"

	"gorm.io/driver/postgres"
)

type DBPoolConfig struct {
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime string `json:"conn_max_lifetime"`
	ConnMaxIdleTime string `json:"conn_max_idle_time"`
}

// NewPostgresDB returns a models.DB using Postgres as the backend (with AutoMigrate)
func NewPostgresDB(dsn string, cfg DBPoolConfig) (DB, error) {
	dialector := postgres.Open(dsn)
	storage, err := NewGormStorage(dialector)
	if err != nil {
		return nil, err
	}
	ConfigureSQLPool(storage, cfg)
	return storage, nil
}

// NewPostgresDBNoMigrate opens the Postgres database without running any schema
func NewPostgresDBNoMigrate(dsn string, cfg DBPoolConfig) (DB, error) {
	dialector := postgres.Open(dsn)
	storage, err := NewGormStorageNoMigrate(dialector)
	if err != nil {
		return nil, err
	}
	ConfigureSQLPool(storage, cfg)
	return storage, nil
}

func ConfigureSQLPool(storage DB, cfg DBPoolConfig) {
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
	if cfg.ConnMaxLifetime != "" {
		if d, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
			sqlDB.SetConnMaxLifetime(d)
		}
	}
	if cfg.ConnMaxIdleTime != "" {
		if d, err := time.ParseDuration(cfg.ConnMaxIdleTime); err == nil {
			sqlDB.SetConnMaxIdleTime(d)
		}
	}
}
