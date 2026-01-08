package persistence

import (
	"context"
	"fmt"
	"strings"

	"kubecloud/internal/core/models"
)

func MigrateAll(ctx context.Context, src models.DB, dst models.DB) error {
	srcGormDB := &GormDB{db: src.GetDB()}
	dstGormDB := &GormDB{db: dst.GetDB()}

	if err := migrateUsers(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("users: %w", err)
	}
	if err := migrateSSHKeys(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("ssh_keys: %w", err)
	}
	if err := migrateVouchers(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("vouchers: %w", err)
	}
	if err := migrateTransactions(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("transactions: %w", err)
	}
	if err := migrateInvoices(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("invoices: %w", err)
	}
	if err := migrateNodeItems(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("node_items: %w", err)
	}
	if err := migrateUserContractData(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("user_contract_data: %w", err)
	}
	if err := migrateClusters(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("clusters: %w", err)
	}
	if err := migrateTransferRecords(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("pending_records: %w", err)
	}
	if err := migrateNotificationsToDst(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("notifications: %w", err)
	}
	// Fix Postgres sequences for tables that were migrated with explicit IDs.
	// When rows are inserted with explicit ID values the table sequence may be
	// left behind (still pointing at a lower value) which later causes
	// duplicate key errors on normal INSERTs. Reset sequences to max(id)+1.
	if err := resetSequences(ctx, dstGormDB); err != nil {
		return fmt.Errorf("reset sequences: %w", err)
	}
	return nil
}

func insertOnConflictReturnError[T any](ctx context.Context, dst *GormDB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return dst.db.WithContext(ctx).Create(&rows).Error
}

func migrateUsers(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.User
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateSSHKeys(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.SSHKey
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateVouchers(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.Voucher
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateTransactions(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.Transaction
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateInvoices(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.Invoice
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateNodeItems(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.NodeItem
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateUserContractData(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.UserContractData
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateClusters(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.Cluster
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateTransferRecords(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.TransferRecord
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateNotificationsToDst(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.Notification
	if err := src.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

// resetSequences ensures postgres serial sequences are at least max(id)+1 for
// tables that use integer auto-increment IDs. This avoids duplicate-key errors
// after copying rows with explicit IDs into an empty DB.
func resetSequences(ctx context.Context, dst *GormDB) error {
	// Only Postgres (or Postgres-like) dialects have sequences we need to reset.
	// If the DB is sqlite or another type, skip the sequence fixes.
	if dst == nil || dst.db == nil || dst.db.Dialector == nil {
		return nil
	}
	// Dialector.Name() returns names like "postgres" or "sqlite" depending on driver.
	if strings.ToLower(dst.db.Name()) != "postgres" {
		// Not Postgres - nothing to do.
		return nil
	}
	tables := []string{
		"users",
		"vouchers",
		"transactions",
		"invoices",
		"node_items",
		"user_nodes",
		"clusters",
		"pending_records",
	}

	var errorsList []string
	for _, table := range tables {
		// set sequence safely to (MAX(id) + 1) or 1 if table empty
		query := fmt.Sprintf("SELECT setval('%s_id_seq', COALESCE((SELECT MAX(id) FROM %s), 0) + 1, false)", table, table)
		if err := dst.db.WithContext(ctx).Exec(query).Error; err != nil {
			// collect and continue
			errorsList = append(errorsList, fmt.Sprintf("%s: %v", table, err))
		}
	}

	if len(errorsList) > 0 {
		return fmt.Errorf("sequence reset errors: %s", strings.Join(errorsList, "; "))
	}
	return nil
}
