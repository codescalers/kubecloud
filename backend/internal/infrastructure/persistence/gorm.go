package persistence

import (
	"context"
	"strings"

	"kubecloud/internal/core/models"

	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// GormDB struct implements models.DB interface with gorm
type GormDB struct {
	db *gorm.DB
}

// NewGormStorage connects to the database using the given dialector
func NewGormStorage(dialector gorm.Dialector) (*GormDB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.Use(tracing.NewPlugin()); err != nil {
		return nil, err
	}

	return newGormDB(db)
}

func newGormDB(db *gorm.DB) (*GormDB, error) {
	// Migrate models
	err := db.AutoMigrate(
		&models.User{},
		&models.Voucher{},
		models.Transaction{},
		models.Invoice{},
		models.NodeItem{},
		models.UserNodes{},
		&models.Notification{},
		&models.SSHKey{},
		&models.Cluster{},
		&models.PendingRecord{},
		&models.Settings{},
	)
	if err != nil {
		return nil, err
	}

	if err := ensureSoftDeleteIndexes(db); err != nil {
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

func ensureSoftDeleteIndexes(db *gorm.DB) error {
	statements := []string{
		`DROP INDEX IF EXISTS idx_user_project`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_project ON clusters (user_id, project_name) WHERE deleted_at IS NULL`,
		`DROP INDEX IF EXISTS idx_user_node_id`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_node_id ON user_nodes (node_id) WHERE deleted_at IS NULL`,
		`DROP INDEX IF EXISTS idx_users_email`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL`,
		`DROP INDEX IF EXISTS idx_user_name`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_name ON ssh_keys (user_id, name) WHERE deleted_at IS NULL`,
		`DROP INDEX IF EXISTS idx_user_pubkey`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_pubkey ON ssh_keys (user_id, public_key) WHERE deleted_at IS NULL`,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		combined := strings.Join(statements, "; ")
		if !strings.HasSuffix(combined, ";") {
			combined += ";"
		}
		return tx.Exec(combined).Error
	})
}
