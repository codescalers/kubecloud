package persistence

import (
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/shared/path"
	"net/url"
	"strings"
)

// NewGormDB creates a DB instance from a DSN string
func NewGormDB(dsn string, cfg models.DBPoolConfig) (models.DB, error) {
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
