package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db DB) *GormUserRepository {
	return &GormUserRepository{db: db.GetDB()}
}

// RegisterUser registers a new user to the system
func (r *GormUserRepository) RegisterUser(user *User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail returns user by its email if found
func (r *GormUserRepository) GetUserByEmail(email string) (User, error) {
	var user User
	query := r.db.First(&user, "email = ?", email)

	if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return User{}, ErrUserNotFound
	}

	return user, query.Error
}

// GetUserByEmail returns user by its email if found
func (r *GormUserRepository) GetUserByID(userID int) (User, error) {
	var user User
	query := r.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (r *GormUserRepository) UpdateUserByID(user *User) error {
	if user.ID == 0 {
		return ErrUserNotFound
	}

	user.UpdatedAt = time.Now()
	query := r.db.Model(&User{}).
		Where("id = ?", user.ID).
		Updates(user)

	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		return query.Error
	}

	if query.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListAllUsers lists all users in system
func (r *GormUserRepository) ListAllUsers() ([]User, error) {
	var users []User

	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListAdmins gets all admins
func (r *GormUserRepository) ListAdmins() ([]User, error) {
	var admins []User
	return admins, r.db.Where("admin = true and verified = true").Find(&admins).Error
}

// DeleteUserByID deletes user by its ID
func (r *GormUserRepository) DeleteUserByID(userID int) error {
	result := r.db.Where("id = ?", userID).Delete(&User{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// CreditUserBalance add credited balance to user by its ID
func (r *GormUserRepository) CreditUserBalance(userID int, amount uint64) error {
	return r.db.Model(&User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}

// CountAllUsers returns the total number of users in the system
func (r *GormUserRepository) CountAllUsers() (int64, error) {
	var count int64
	err := r.db.Model(&User{}).Count(&count).Error
	return count, err
}

// CreateNotification creates a new notification
// CreateSSHKey creates a new SSH key for a user
func (r *GormUserRepository) CreateSSHKey(sshKey *SSHKey) error {
	sshKey.CreatedAt = time.Now()
	sshKey.UpdatedAt = time.Now()
	query := r.db.Create(sshKey)

	if isUniqueViolation(query.Error) {
		return ErrSSHKeyAlreadyExists
	}
	return query.Error
}

// ListUserSSHKeys returns all SSH keys for a user
func (r *GormUserRepository) ListUserSSHKeys(userID int) ([]SSHKey, error) {
	var sshKeys []SSHKey
	err := r.db.Where("user_id = ?", userID).Find(&sshKeys).Error
	if err != nil {
		return nil, err
	}
	return sshKeys, nil
}

// DeleteSSHKey deletes an SSH key by ID for a specific user and returns the key name
func (r *GormUserRepository) DeleteSSHKey(sshKeyID int, userID int) (string, error) {
	// First get the SSH key to retrieve its name
	sshKey, err := r.GetSSHKeyByID(sshKeyID, userID)
	if err != nil {
		return "", err
	}

	// Then delete the SSH key
	result := r.db.Where("id = ? AND user_id = ?", sshKeyID, userID).Delete(&SSHKey{})
	if result.Error != nil {
		return "", result.Error
	}

	if result.RowsAffected == 0 {
		return "", ErrSSHKeyNotFound
	}

	return sshKey.Name, nil
}

// GetSSHKeyByID returns an SSH key by ID for a specific user
func (r *GormUserRepository) GetSSHKeyByID(sshKeyID int, userID int) (SSHKey, error) {
	var sshKey SSHKey
	query := r.db.Where("id = ? AND user_id = ?", sshKeyID, userID).First(&sshKey)

	if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return SSHKey{}, ErrSSHKeyNotFound
	}

	return sshKey, query.Error
}

type GormClusterRepository struct {
	db *gorm.DB
}

func NewGormClusterRepository(db DB) *GormClusterRepository {
	return &GormClusterRepository{db: db.GetDB()}
}

// CreateCluster creates a new cluster in the database
func (r *GormClusterRepository) CreateCluster(userID int, cluster *Cluster) error {
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	cluster.UserID = userID
	return r.db.Create(cluster).Error
}

// ListUserClusters returns all clusters for a specific user
func (r *GormClusterRepository) ListUserClusters(userID int) ([]Cluster, error) {
	var clusters []Cluster
	query := r.db.Where("user_id = ?", userID).Find(&clusters)
	return clusters, query.Error
}

// GetClusterByName returns a cluster by name for a specific user
func (r *GormClusterRepository) GetClusterByName(userID int, projectName string) (Cluster, error) {
	var cluster Cluster
	query := r.db.Where("user_id = ? AND project_name = ?", userID, projectName).First(&cluster)

	if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return Cluster{}, ErrClusterNotFound
	}
	return cluster, query.Error
}

// UpdateCluster updates an existing cluster
func (r *GormClusterRepository) UpdateCluster(cluster *Cluster) error {
	cluster.UpdatedAt = time.Now()
	return r.db.Model(&Cluster{}).
		Where("user_id = ? AND project_name = ?", cluster.UserID, cluster.ProjectName).
		Updates(cluster).Error
}

// DeleteCluster deletes a cluster by name for a specific user
func (r *GormClusterRepository) DeleteCluster(userID int, projectName string) error {
	query := r.db.Where("user_id = ? AND project_name = ?", userID, projectName).Delete(&Cluster{})

	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return ErrClusterNotFound
	}

	if query.RowsAffected == 0 {
		return ErrClusterNotFound
	}

	return query.Error
}

// DeleteAllUserClusters deletes all clusters for a specific user
func (r *GormClusterRepository) DeleteAllUserClusters(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&Cluster{}).Error
}

// CountAllClusters returns the total number of clusters in the system
func (r *GormClusterRepository) CountAllClusters() (int64, error) {
	var count int64
	err := r.db.Model(&Cluster{}).Count(&count).Error
	return count, err
}

func (r *GormClusterRepository) ListAllClusters() ([]Cluster, error) {
	var clusters []Cluster
	return clusters, r.db.Find(&clusters).Error
}

type GormVoucherRepository struct {
	db *gorm.DB
}

func NewGormVoucherRepository(db DB) *GormVoucherRepository {
	return &GormVoucherRepository{db: db.GetDB()}
}

// CreateVoucher creates new voucher in system
func (r *GormVoucherRepository) CreateVoucher(voucher *Voucher) error {
	return r.db.Create(voucher).Error
}

// ListAllVouchers gets all vouchers in system
func (r *GormVoucherRepository) ListAllVouchers() ([]Voucher, error) {
	var vouchers []Voucher

	err := r.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

// GetVoucherByCode returns voucher by its code
func (r *GormVoucherRepository) GetVoucherByCode(code string) (Voucher, error) {
	var voucher Voucher
	query := r.db.First(&voucher, "code = ?", code)

	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return Voucher{}, ErrVoucherNotFound
	}

	return voucher, query.Error
}

// RedeemVoucher updates status if voucher
func (r *GormVoucherRepository) RedeemVoucher(code string) error {
	result := r.db.Model(&Voucher{}).
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

type GormUserNodesRepository struct {
	db *gorm.DB
}

func NewGormUserNodesRepository(db DB) *GormUserNodesRepository {
	return &GormUserNodesRepository{db: db.GetDB()}
}

// CreateUserNode creates new node record for user
func (r *GormUserNodesRepository) CreateUserNode(userNode *UserNodes) error {
	return r.db.Create(&userNode).Error
}

// DeleteUserNode deletes a node record for user by its contract ID
func (r *GormUserNodesRepository) DeleteUserNode(contractID uint64) error {
	return r.db.Where("contract_id = ?", contractID).Delete(&UserNodes{}).Error
}

// ListUserNodes returns all nodes records for user by its ID
func (r *GormUserNodesRepository) ListUserNodes(userID int) ([]UserNodes, error) {
	var userNodes []UserNodes
	return userNodes, r.db.Where("user_id = ?", userID).Find(&userNodes).Error
}

// ListAllReservedNodes returns all reserved nodes from all users
func (r *GormUserNodesRepository) ListAllReservedNodes() ([]UserNodes, error) {
	var userNodes []UserNodes
	return userNodes, r.db.Find(&userNodes).Error
}

func (r *GormUserNodesRepository) GetUserNodeByNodeID(nodeID uint64) (UserNodes, error) {
	var userNode UserNodes
	result := r.db.Where("node_id = ?", nodeID).First(&userNode)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return UserNodes{}, ErrUserNodeNotFound
	}

	return userNode, result.Error
}

func (r *GormUserNodesRepository) GetUserNodeByContractID(contractID uint64) (UserNodes, error) {
	var userNode UserNodes
	result := r.db.Where("contract_id = ?", contractID).First(&userNode)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return UserNodes{}, ErrUserNodeNotFound
	}

	return userNode, result.Error
}

type GormTransactionRepository struct {
	db *gorm.DB
}

func NewGormTransactionRepository(db DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db.GetDB()}
}

// CreateTransaction creates a payment transaction
func (r *GormTransactionRepository) CreateTransaction(transaction *Transaction) error {
	return r.db.Create(transaction).Error
}

type GormSettingsRepository struct {
	db *gorm.DB
}

func NewGormSettingsRepository(db DB) *GormSettingsRepository {
	return &GormSettingsRepository{db: db.GetDB()}
}

// GetSetting retrieves a setting value by name
func (r *GormSettingsRepository) GetSetting(name string) (string, error) {
	var setting Settings
	err := r.db.Where("name = ?", name).First(&setting).Error
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

// SetSetting sets a setting value (creates or updates)
func (r *GormSettingsRepository) SetSetting(name, value string) error {
	setting := Settings{
		Name:  name,
		Value: value,
	}

	return r.db.Save(&setting).Error
}

// SetMaintenanceMode sets the maintenance mode
func (r *GormSettingsRepository) SetMaintenanceMode(enabled bool) error {
	value := maintenanceModeDisabled
	if enabled {
		value = maintenanceModeEnabled
	}
	return r.SetSetting("maintenance_mode", value)
}

// GetMaintenanceMode gets the current maintenance mode status
func (r *GormSettingsRepository) GetMaintenanceMode() (bool, error) {
	value, err := r.GetSetting("maintenance_mode")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if value == "" {
		return false, nil
	}

	return value == maintenanceModeEnabled, nil
}

type GormPendingRecordRepository struct {
	db *gorm.DB
}

func NewGormPendingRecordRepository(db DB) *GormPendingRecordRepository {
	return &GormPendingRecordRepository{db: db.GetDB()}
}

func (r *GormPendingRecordRepository) CreatePendingRecord(record *PendingRecord) error {
	record.CreatedAt = time.Now()
	return r.db.Create(record).Error
}

func (r *GormPendingRecordRepository) ListAllPendingRecords() ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, r.db.Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) ListOnlyPendingRecords() ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, r.db.Where("tft_amount > transferred_tft_amount").Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) ListUserPendingRecords(userID int) ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, r.db.Where("user_id = ?", userID).Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) UpdatePendingRecordTransferredAmount(id int, amount uint64) error {
	return r.db.Model(&PendingRecord{}).
		Where("id = ?", id).
		UpdateColumn("transferred_tft_amount", gorm.Expr("transferred_tft_amount + ?", amount)).
		UpdateColumn("updated_at", gorm.Expr("?", time.Now())).
		Error
}

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db.GetDB()}
}

// CreateNotification creates a new notification
func (r *GormNotificationRepository) CreateNotification(notification *Notification) error {
	return r.db.Create(notification).Error
}

// GetUserNotifications retrieves notifications for a user with pagination
func (r *GormNotificationRepository) GetUserNotifications(userID int, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkNotificationAsRead marks a specific notification as read
func (r *GormNotificationRepository) MarkNotificationAsRead(notificationID string, userID int) error {
	now := time.Now()
	result := r.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  NotificationStatusRead,
			"read_at": &now,
		})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func (r *GormNotificationRepository) MarkAllNotificationsAsRead(userID int) error {
	now := time.Now()
	return r.db.Model(&Notification{}).
		Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  NotificationStatusRead,
			"read_at": &now,
		}).Error
}

// DeleteNotification deletes a notification for a user
func (r *GormNotificationRepository) DeleteNotification(notificationID string, userID int) error {
	result := r.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&Notification{})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// GetUnreadNotifications retrieves only unread notifications for a user with pagination
func (r *GormNotificationRepository) GetUnreadNotifications(userID int, limit, offset int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("user_id = ? AND status = ?", userID, NotificationStatusUnread).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// DeleteAllNotifications deletes all notifications for a user
func (r *GormNotificationRepository) DeleteAllNotifications(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&Notification{}).Error
}

// MarkNotificationAsUnread marks a specific notification as unread
func (r *GormNotificationRepository) MarkNotificationAsUnread(notificationID string, userID int) error {
	result := r.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  NotificationStatusUnread,
			"read_at": nil,
		})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

type GormInvoiceRepository struct {
	db *gorm.DB
}

func NewGormInvoiceRepository(db DB) *GormInvoiceRepository {
	return &GormInvoiceRepository{db: db.GetDB()}
}

// CreateInvoice creates new invoice
func (r *GormInvoiceRepository) CreateInvoice(invoice *Invoice) error {
	return r.db.Create(&invoice).Error
}

// GetInvoice returns an invoice by ID
func (r *GormInvoiceRepository) GetInvoice(id int) (Invoice, error) {
	var invoice Invoice
	err := r.db.First(&invoice, id).Error
	if err != nil {
		return Invoice{}, err
	}

	var nodes []NodeItem
	if err = r.db.Model(&invoice).Association("Nodes").Find(&nodes); err != nil {
		return Invoice{}, err
	}

	invoice.Nodes = nodes
	return invoice, nil
}

// ListUserInvoices returns all invoices of user
func (r *GormInvoiceRepository) ListUserInvoices(userID int) ([]Invoice, error) {
	var invoices []Invoice
	err := r.db.Where("user_id = ?", userID).Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	for idx := range invoices {
		invoices[idx], err = r.GetInvoice(invoices[idx].ID)
		if err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

// ListInvoices returns all invoices (admin)
func (r *GormInvoiceRepository) ListInvoices() ([]Invoice, error) {
	var invoices []Invoice
	err := r.db.Find(&invoices).Error

	if err != nil {
		return nil, err
	}

	for idx := range invoices {
		invoices[idx], err = r.GetInvoice(invoices[idx].ID)
		if err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

func (r *GormInvoiceRepository) UpdateInvoicePDF(id int, data []byte) error {
	return r.db.Model(&Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"file_data": data}).Error
}
