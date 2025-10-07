package models

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// GormDB struct implements db interface with gorm
type GormDB struct {
	db    *gorm.DB
	mutex sync.Mutex
}

// NewGormStorage connects to the database using the given dialector
func NewGormStorage(dialector gorm.Dialector) (DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate models
	err = db.AutoMigrate(
		&User{},
		&Voucher{},
		Transaction{},
		Invoice{},
		NodeItem{},
		UserNodes{},
		&Notification{},
		&SSHKey{},
		&Cluster{},
		&PendingRecord{},
	)
	if err != nil {
		return nil, err
	}

	if err := migrateNotifications(db); err != nil {
		return nil, err
	}

	gormDB := &GormDB{db: db}
	return gormDB, gormDB.UpdatePendingRecordsWithUsername()
}

func NewGormStorageNoMigrate(dialector gorm.Dialector) (DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &GormDB{db: db}, nil
}

// TODO: TO BE REMOVED
func (s *GormDB) UpdatePendingRecordsWithUsername() error {
	var pendingRecords []PendingRecord
	if err := s.db.Find(&pendingRecords).Where("username IS NULL").Error; err != nil {
		return fmt.Errorf("failed to find pending records: %w", err)
	}

	for _, record := range pendingRecords {
		user, err := s.GetUserByID(record.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user by ID %d: %w", record.UserID, err)
		}

		// Update the record with the username
		if err := s.db.Model(&record).Update("username", user.Username).Error; err != nil {
			return fmt.Errorf("failed to update pending record with username: %w", err)
		}
	}

	return nil
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

// RegisterUser registers a new user to the system
func (s *GormDB) RegisterUser(user *User) error {
	return s.db.Create(user).Error
}

// GetUserByEmail returns user by its email if found
func (s *GormDB) GetUserByEmail(email string) (User, error) {
	var user User
	query := s.db.First(&user, "email = ?", email)
	return user, query.Error
}

// GetUserByEmail returns user by its email if found
func (s *GormDB) GetUserByID(userID int) (User, error) {
	var user User
	query := s.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (s *GormDB) UpdateUserByID(user *User) error {
	user.UpdatedAt = time.Now()
	return s.db.Model(&User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
}

// UpdatePassword updates password of user by its email
func (s *GormDB) UpdatePassword(email string, hashedPassword []byte) error {
	result := s.db.Model(&User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with email %s", email)
	}

	return nil
}

// ListAllUsers lists all users in system
func (s *GormDB) ListAllUsers() ([]User, error) {
	var users []User

	err := s.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListAdmins gets all admins
func (s *GormDB) ListAdmins() ([]User, error) {
	var admins []User
	return admins, s.db.Where("admin = true and verified = true").Find(&admins).Error
}

// DeleteUserByID deletes user by its ID
func (s *GormDB) DeleteUserByID(userID int) error {
	result := s.db.Where("id = ?", userID).Delete(&User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateVoucher creates new voucher in system
func (s *GormDB) CreateVoucher(voucher *Voucher) error {
	return s.db.Create(voucher).Error
}

// ListAllVouchers gets all vouchers in system
func (s *GormDB) ListAllVouchers() ([]Voucher, error) {
	var vouchers []Voucher

	err := s.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

// GetVoucherByCode returns voucher by its code
func (s *GormDB) GetVoucherByCode(code string) (Voucher, error) {
	var voucher Voucher
	query := s.db.First(&voucher, "code = ?", code)
	return voucher, query.Error
}

// RedeemVoucher updates status if voucher
func (s *GormDB) RedeemVoucher(code string) error {
	result := s.db.Model(&Voucher{}).
		Where("code = ?", code).
		Update("redeemed", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no voucher found with Code %s", code)
	}

	return nil
}

// CreateTransaction creates a payment transaction
func (s *GormDB) CreateTransaction(transaction *Transaction) error {
	return s.db.Create(transaction).Error
}

// CreditUserBalance add credited balance to user by its ID
func (s *GormDB) CreditUserBalance(userID int, amount uint64) error {
	return s.db.Model(&User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}

// CreateInvoice creates new invoice
func (s *GormDB) CreateInvoice(invoice *Invoice) error {
	return s.db.Create(&invoice).Error
}

// GetInvoice returns an invoice by ID
func (s *GormDB) GetInvoice(id int) (Invoice, error) {
	var invoice Invoice
	err := s.db.First(&invoice, id).Error
	if err != nil {
		return Invoice{}, err
	}

	var nodes []NodeItem
	if err = s.db.Model(&invoice).Association("Nodes").Find(&nodes); err != nil {
		return Invoice{}, err
	}

	invoice.Nodes = nodes
	return invoice, nil
}

// ListUserInvoices returns all invoices of user
func (s *GormDB) ListUserInvoices(userID int) ([]Invoice, error) {
	var invoices []Invoice
	err := s.db.Where("user_id = ?", userID).Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	for idx := range invoices {
		invoices[idx], err = s.GetInvoice(invoices[idx].ID)
		if err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

// ListInvoices returns all invoices (admin)
func (s *GormDB) ListInvoices() ([]Invoice, error) {
	var invoices []Invoice
	err := s.db.Find(&invoices).Error

	if err != nil {
		return nil, err
	}

	for idx := range invoices {
		invoices[idx], err = s.GetInvoice(invoices[idx].ID)
		if err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

func (s *GormDB) UpdateInvoicePDF(id int, data []byte) error {
	return s.db.Model(&Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"file_data": data}).Error
}

// CreateUserNode creates new node record for user
func (s *GormDB) CreateUserNode(userNode *UserNodes) error {
	return s.db.Create(&userNode).Error
}

// DeleteUserNode deletes a node record for user by its contract ID
func (s *GormDB) DeleteUserNode(contractID uint64) error {
	return s.db.Where("contract_id = ?", contractID).Delete(&UserNodes{}).Error
}

// ListUserNodes returns all nodes records for user by its ID
func (s *GormDB) ListUserNodes(userID int) ([]UserNodes, error) {
	var userNodes []UserNodes
	return userNodes, s.db.Where("user_id = ?", userID).Find(&userNodes).Error
}

// ListAllReservedNodes returns all reserved nodes from all users
func (s *GormDB) ListAllReservedNodes() ([]UserNodes, error) {
	var userNodes []UserNodes
	return userNodes, s.db.Find(&userNodes).Error
}

func (s *GormDB) GetUserNodeByNodeID(nodeID uint64) (UserNodes, error) {
	var userNode UserNodes
	return userNode, s.db.Where("node_id = ?", nodeID).First(&userNode).Error
}

// CreateNotification creates a new notification
// CreateSSHKey creates a new SSH key for a user
func (s *GormDB) CreateSSHKey(sshKey *SSHKey) error {
	sshKey.CreatedAt = time.Now()
	sshKey.UpdatedAt = time.Now()
	return s.db.Create(sshKey).Error
}

// ListUserSSHKeys returns all SSH keys for a user
func (s *GormDB) ListUserSSHKeys(userID int) ([]SSHKey, error) {
	var sshKeys []SSHKey
	err := s.db.Where("user_id = ?", userID).Find(&sshKeys).Error
	if err != nil {
		return nil, err
	}
	return sshKeys, nil
}

// DeleteSSHKey deletes an SSH key by ID for a specific user
func (s *GormDB) DeleteSSHKey(sshKeyID int, userID int) error {
	result := s.db.Where("id = ? AND user_id = ?", sshKeyID, userID).Delete(&SSHKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no SSH key found with ID %d for user %d", sshKeyID, userID)
	}
	return nil
}

// GetSSHKeyByID returns an SSH key by ID for a specific user
func (s *GormDB) GetSSHKeyByID(sshKeyID int, userID int) (SSHKey, error) {
	var sshKey SSHKey
	query := s.db.Where("id = ? AND user_id = ?", sshKeyID, userID).First(&sshKey)
	return sshKey, query.Error
}

// CreateCluster creates a new cluster in the database
func (s *GormDB) CreateCluster(userID int, cluster *Cluster) error {
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	cluster.UserID = userID
	return s.db.Create(cluster).Error
}

// ListUserClusters returns all clusters for a specific user
func (s *GormDB) ListUserClusters(userID int) ([]Cluster, error) {
	var clusters []Cluster
	query := s.db.Where("user_id = ?", userID).Find(&clusters)
	return clusters, query.Error
}

// GetClusterByName returns a cluster by name for a specific user
func (s *GormDB) GetClusterByName(userID int, projectName string) (Cluster, error) {
	var cluster Cluster
	query := s.db.Where("user_id = ? AND project_name = ?", userID, projectName).First(&cluster)
	return cluster, query.Error
}

// UpdateCluster updates an existing cluster
func (s *GormDB) UpdateCluster(cluster *Cluster) error {
	cluster.UpdatedAt = time.Now()
	return s.db.Model(&Cluster{}).
		Where("user_id = ? AND project_name = ?", cluster.UserID, cluster.ProjectName).
		Updates(cluster).Error
}

// DeleteCluster deletes a cluster by name for a specific user
func (s *GormDB) DeleteCluster(userID int, projectName string) error {
	return s.db.Where("user_id = ? AND project_name = ?", userID, projectName).Delete(&Cluster{}).Error
}

// DeleteAllUserClusters deletes all clusters for a specific user
func (s *GormDB) DeleteAllUserClusters(userID int) error {
	return s.db.Where("user_id = ?", userID).Delete(&Cluster{}).Error
}

func (s *GormDB) CreatePendingRecord(record *PendingRecord) error {
	record.CreatedAt = time.Now()
	return s.db.Create(record).Error
}

func (s *GormDB) ListAllPendingRecords() ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, s.db.Find(&pendingRecords).Error
}

func (s *GormDB) ListOnlyPendingRecords() ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, s.db.Where("tft_amount > transferred_tft_amount").Find(&pendingRecords).Error
}

func (s *GormDB) ListUserPendingRecords(userID int) ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, s.db.Where("user_id = ?", userID).Find(&pendingRecords).Error
}

func (s *GormDB) UpdatePendingRecordTransferredAmount(id int, amount uint64) error {
	return s.db.Model(&PendingRecord{}).
		Where("id = ?", id).
		UpdateColumn("transferred_tft_amount", gorm.Expr("transferred_tft_amount + ?", amount)).
		UpdateColumn("updated_at", gorm.Expr("?", time.Now())).
		Error
}

// CountAllUsers returns the total number of users in the system
func (s *GormDB) CountAllUsers() (int64, error) {
	var count int64
	err := s.db.Model(&User{}).Count(&count).Error
	return count, err
}

// CountAllClusters returns the total number of clusters in the system
func (s *GormDB) CountAllClusters() (int64, error) {
	var count int64
	err := s.db.Model(&Cluster{}).Count(&count).Error
	return count, err
}

func (s *GormDB) ListAllClusters() ([]Cluster, error) {
	var clusters []Cluster
	return clusters, s.db.Find(&clusters).Error
}

func (s *GormDB) GetUserNodeByContractID(contractID uint64) (UserNodes, error) {
	var userNode UserNodes
	return userNode, s.db.Where("contract_id = ?", contractID).First(&userNode).Error
}

func migrateNotifications(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&Notification{}) {
		return nil
	}

	if m.HasColumn(&Notification{}, "data") {
		if err := db.Exec("UPDATE notifications SET payload = data").Error; err != nil {
			return err
		}
		_ = m.DropColumn(&Notification{}, "data")
	}

	if m.HasColumn(&Notification{}, "message") {
		_ = m.DropColumn(&Notification{}, "message")
	}
	if m.HasColumn(&Notification{}, "title") {
		_ = m.DropColumn(&Notification{}, "title")
	}

	testNotification := &Notification{
		ID:        "test-migration-id",
		UserID:    1,
		Type:      NotificationTypeDeployment,
		Severity:  NotificationSeverityInfo,
		Channels:  []string{"ui"},
		Payload:   map[string]string{"test": "value"},
		Status:    NotificationStatusUnread,
		CreatedAt: time.Now(),
	}

	err := db.Create(testNotification).Error
	if err == nil {
		db.Delete(testNotification, "id = ?", "test-migration-id")
		return nil
	}

	if !strings.Contains(err.Error(), "datatype mismatch") {
		return fmt.Errorf("failed to create test notification during migration: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_task_id").Error; err != nil {
			return err
		}
		if err := tx.Exec("DROP INDEX IF EXISTS idx_notifications_user_id").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE notifications RENAME TO notifications_old").Error; err != nil {
			return err
		}

		if err := tx.AutoMigrate(&Notification{}); err != nil {
			return err
		}

		if err := tx.Exec("DROP TABLE notifications_old").Error; err != nil {
			return err
		}

		return nil
	})
}
