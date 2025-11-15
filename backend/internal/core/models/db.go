package models

import (
	"context"

	"gorm.io/gorm"
)

// DB interface defines the contract for database connections
type DB interface {
	Ping(ctx context.Context) error
	Close() error
	GetDB() *gorm.DB
}

// DBPoolConfig holds database connection pool configuration
type DBPoolConfig struct {
	MaxOpenConns           int `json:"max_open_conns"`
	MaxIdleConns           int `json:"max_idle_conns"`
	ConnMaxLifetimeMinutes int `json:"conn_max_lifetime_minutes"`
	ConnMaxIdleTimeMinutes int `json:"conn_max_idle_time_minutes"`
}
