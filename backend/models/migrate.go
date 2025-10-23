package models

import (
	"context"
	"fmt"
	"kubecloud/internal/logger"
	"strings"

	"gorm.io/gorm"
)

type FileStorage interface {
	WriteInvoiceFile(userID, invoiceID int, data []byte) (string, error)
	WriteKubeconfigFile(userID, clusterID int, projectName string, data []byte) (string, error)
}

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

func MigrateFileDataToStorage(db DB, fileStorage FileStorage) error {
	gormDB := db.GetDB()
	m := gormDB.Migrator()

	if err := migrateInvoiceFileData(gormDB, m, fileStorage); err != nil {
		return fmt.Errorf("failed to migrate invoice file data: %w", err)
	}

	// Migrate cluster kubeconfigs from kubeconfig column to file storage
	if err := migrateClusterKubeconfigs(gormDB, m, fileStorage); err != nil {
		return fmt.Errorf("failed to migrate cluster kubeconfigs: %w", err)
	}

	return nil
}

func migrateInvoiceFileData(db *gorm.DB, m gorm.Migrator, fileStorage FileStorage) error {

	if !m.HasTable(&Invoice{}) || !m.HasColumn(&Invoice{}, "file_data") {
		return nil
	}

	type InvoiceWithFileData struct {
		ID       int    `gorm:"primaryKey"`
		UserID   int    `gorm:"user_id"`
		FileData []byte `gorm:"type:bytea;column:file_data"`
	}

	var invoices []InvoiceWithFileData
	if err := db.Table("invoices").Where("file_data IS NOT NULL AND file_data != ''").Find(&invoices).Error; err != nil {
		return fmt.Errorf("failed to query invoices with file data: %w", err)
	}

	migratedCount := 0
	for _, invoice := range invoices {
		if len(invoice.FileData) == 0 {
			continue
		}

		// Use the file storage service to write the file
		if _, err := fileStorage.WriteInvoiceFile(invoice.UserID, invoice.ID, invoice.FileData); err != nil {
			return fmt.Errorf("failed to write invoice file for invoice %d: %w", invoice.ID, err)
		}

		migratedCount++
	}

	if err := m.DropColumn(&Invoice{}, "file_data"); err != nil {
		return fmt.Errorf("failed to drop file_data column: %w", err)
	}

	logger.GetLogger().Info().Msgf("Successfully migrated %d invoice PDFs to file storage", migratedCount)
	return nil
}

func migrateClusterKubeconfigs(db *gorm.DB, m gorm.Migrator, fileStorage FileStorage) error {
	if !m.HasTable(&Cluster{}) || !m.HasColumn(&Cluster{}, "kubeconfig") {
		return nil
	}

	type ClusterWithKubeconfig struct {
		ID          int    `gorm:"primaryKey"`
		UserID      int    `gorm:"user_id"`
		ProjectName string `gorm:"project_name"`
		Kubeconfig  string `gorm:"type:text"`
	}

	var clusters []ClusterWithKubeconfig
	if err := db.Table("clusters").Where("kubeconfig IS NOT NULL AND kubeconfig != ''").Find(&clusters).Error; err != nil {
		return fmt.Errorf("failed to query clusters with kubeconfig: %w", err)
	}

	migratedCount := 0
	for _, cluster := range clusters {
		if strings.TrimSpace(cluster.Kubeconfig) == "" {
			continue
		}

		if _, err := fileStorage.WriteKubeconfigFile(cluster.UserID, cluster.ID, cluster.ProjectName, []byte(cluster.Kubeconfig)); err != nil {
			return fmt.Errorf("failed to write kubeconfig file for cluster %d: %w", cluster.ID, err)
		}

		migratedCount++
	}

	if err := m.DropColumn(&Cluster{}, "kubeconfig"); err != nil {
		return fmt.Errorf("failed to drop kubeconfig column: %w", err)
	}

	logger.GetLogger().Info().Msgf("Successfully migrated %d cluster kubeconfigs to file storage", migratedCount)
	return nil
}
