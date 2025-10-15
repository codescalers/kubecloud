package models

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func MigrateAll(ctx context.Context, src DB, dst DB) error {
	if err := migrateUsers(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("users: %w", err)
	}
	if err := migrateSSHKeys(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("ssh_keys: %w", err)
	}
	if err := migrateVouchers(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("vouchers: %w", err)
	}
	if err := migrateTransactions(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("transactions: %w", err)
	}
	if err := migrateInvoices(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("invoices: %w", err)
	}
	if err := migrateNodeItems(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("node_items: %w", err)
	}
	if err := migrateUserNodes(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("user_nodes: %w", err)
	}
	if err := migrateClusters(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("clusters: %w", err)
	}
	if err := migratePendingRecords(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("pending_records: %w", err)
	}
	if err := migrateNotificationsToDst(ctx, src.GetDB(), dst.GetDB()); err != nil {
		return fmt.Errorf("notifications: %w", err)
	}
	return nil
}

func insertOnConflictReturnError[T any](ctx context.Context, dst *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return dst.WithContext(ctx).Create(&rows).Error
}

func migrateUsers(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []User
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateSSHKeys(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []SSHKey
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateVouchers(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []Voucher
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateTransactions(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []Transaction
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateInvoices(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []Invoice
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateNodeItems(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []NodeItem
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateUserNodes(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []UserNodes
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateClusters(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []Cluster
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migratePendingRecords(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []PendingRecord
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}

func migrateNotificationsToDst(ctx context.Context, src *gorm.DB, dst *gorm.DB) error {
	var rows []Notification
	if err := src.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}
	return insertOnConflictReturnError(ctx, dst, rows)
}
