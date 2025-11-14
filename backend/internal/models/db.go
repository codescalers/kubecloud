package models

import (
	"context"
	"fmt"
	"kubecloud/internal/path"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

type DB interface {
	Ping(ctx context.Context) error
	Close() error
	GetDB() *gorm.DB
}

func NewGormDB(dsn string, cfg DBPoolConfig) (*GormDB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	switch u.Scheme {
	case "postgres":
		return NewPostgresGormDB(dsn, cfg)
	case "sqlite", "sqlite3":
		path, err := path.ExpandPath(u.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path: %w", err)
		}
		return NewSqliteGormDB(path)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}
