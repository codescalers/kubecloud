package persistence

import (
	"context"
	"fmt"

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
	if err := migrateUserNodes(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("user_nodes: %w", err)
	}
	if err := migrateClusters(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("clusters: %w", err)
	}
	if err := migratePendingRecords(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("pending_records: %w", err)
	}
	if err := migrateNotificationsToDst(ctx, srcGormDB, dstGormDB); err != nil {
		return fmt.Errorf("notifications: %w", err)
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

func migrateUserNodes(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.UserNodes
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

func migratePendingRecords(ctx context.Context, src *GormDB, dst *GormDB) error {
	var rows []models.PendingRecord
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
