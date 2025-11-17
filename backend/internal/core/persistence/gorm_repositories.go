package persistence

import (
	"errors"
	"fmt"
	"kubecloud/internal/core/models"
	"time"

	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

const gormUserID = "gorm_user_id"

// User Repository

var _ models.UserRepository = (*GormUserRepository)(nil)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db models.DB) models.UserRepository {
	return &GormUserRepository{db: db.GetDB()}
}

// RegisterUser registers a new user to the system
func (r *GormUserRepository) RegisterUser(user *models.User) error {
	return r.db.Create(user).Error
}

// SetStateUserID sets gorm user ID in workflow state
func SetStateUserID(wf *ewf.Workflow, userID int) error {
	if wf == nil {
		return fmt.Errorf("workflow is nil")
	}
	if wf.State == nil {
		wf.State = make(ewf.State)
	}
	wf.State[gormUserID] = userID
	return nil
}

// GetUserByEmail returns user by its email if found
func (r *GormUserRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := r.db.First(&user, "email = ?", email)

	if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return models.User{}, models.ErrUserNotFound
	}

	return user, query.Error
}

// GetUserByID returns user by its ID
func (r *GormUserRepository) GetUserByID(userID int) (models.User, error) {
	var user models.User
	query := r.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (r *GormUserRepository) UpdateUserByID(user *models.User) error {
	if user.ID == 0 {
		return models.ErrUserNotFound
	}

	user.UpdatedAt = time.Now()
	query := r.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user)

	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return models.ErrUserNotFound
		}

		return query.Error
	}

	if query.RowsAffected == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// ListAllUsers lists all users in system
func (r *GormUserRepository) ListAllUsers() ([]models.User, error) {
	var users []models.User

	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListAdmins gets all admins
func (r *GormUserRepository) ListAdmins() ([]models.User, error) {
	var admins []models.User
	return admins, r.db.Where("admin = true and verified = true").Find(&admins).Error
}

// DeleteUserByID deletes user by its ID
func (r *GormUserRepository) DeleteUserByID(userID int) error {
	result := r.db.Where("id = ?", userID).Delete(&models.User{})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.ErrUserNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrUserNotFound
	}
	return nil
}

// CreditUserBalance add credited balance to user by its ID
func (r *GormUserRepository) CreditUserBalance(userID int, amount uint64) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}

// CountAllUsers returns the total number of users in the system
func (r *GormUserRepository) CountAllUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CreateSSHKey creates a new SSH key for a user
func (r *GormUserRepository) CreateSSHKey(sshKey *models.SSHKey) error {
	sshKey.CreatedAt = time.Now()
	sshKey.UpdatedAt = time.Now()
	query := r.db.Create(sshKey)

	if models.IsUniqueViolation(query.Error) {
		return models.ErrSSHKeyAlreadyExists
	}
	return query.Error
}

// ListUserSSHKeys returns all SSH keys for a user
func (r *GormUserRepository) ListUserSSHKeys(userID int) ([]models.SSHKey, error) {
	var sshKeys []models.SSHKey
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
	result := r.db.Where("id = ? AND user_id = ?", sshKeyID, userID).Delete(&models.SSHKey{})
	if result.Error != nil {
		return "", result.Error
	}

	if result.RowsAffected == 0 {
		return "", models.ErrSSHKeyNotFound
	}

	return sshKey.Name, nil
}

// GetSSHKeyByID returns an SSH key by ID for a specific user
func (r *GormUserRepository) GetSSHKeyByID(sshKeyID int, userID int) (models.SSHKey, error) {
	var sshKey models.SSHKey
	query := r.db.Where("id = ? AND user_id = ?", sshKeyID, userID).First(&sshKey)

	if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return models.SSHKey{}, models.ErrSSHKeyNotFound
	}

	return sshKey, query.Error
}

// ListRemainingWorkflowsByUserID returns remaining workflows for a specific user
func (r *GormUserRepository) ListRemainingWorkflowsByUserID(userID int) ([]models.GormWorkflowRecord, error) {
	var records []models.GormWorkflowRecord

	if err := r.db.Where("user_id = ?", userID).
		Where("status IN ?", []string{string(ewf.StatusPending), string(ewf.StatusRunning)}).
		Order("uuid DESC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

// Cluster Repository

var _ models.ClusterRepository = (*GormClusterRepository)(nil)

type GormClusterRepository struct {
	db *gorm.DB
}

func NewGormClusterRepository(db models.DB) models.ClusterRepository {
	return &GormClusterRepository{db: db.GetDB()}
}

func (r *GormClusterRepository) CreateCluster(userID int, cluster *models.Cluster) error {
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	cluster.UserID = userID
	return r.db.Create(cluster).Error
}

func (r *GormClusterRepository) ListUserClusters(userID int) ([]models.Cluster, error) {
	var clusters []models.Cluster
	query := r.db.Where("user_id = ?", userID).Find(&clusters)
	return clusters, query.Error
}

func (r *GormClusterRepository) GetClusterByName(userID int, projectName string) (models.Cluster, error) {
	var cluster models.Cluster
	query := r.db.Where("user_id = ? AND project_name = ?", userID, projectName).First(&cluster)

	if query.Error != nil && errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return models.Cluster{}, models.ErrClusterNotFound
	}
	return cluster, query.Error
}

func (r *GormClusterRepository) UpdateCluster(cluster *models.Cluster) error {
	cluster.UpdatedAt = time.Now()
	return r.db.Model(&models.Cluster{}).
		Where("user_id = ? AND project_name = ?", cluster.UserID, cluster.ProjectName).
		Updates(cluster).Error
}

func (r *GormClusterRepository) DeleteCluster(userID int, projectName string) error {
	query := r.db.Where("user_id = ? AND project_name = ?", userID, projectName).Delete(&models.Cluster{})

	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return models.ErrClusterNotFound
	}

	if query.RowsAffected == 0 {
		return models.ErrClusterNotFound
	}

	return query.Error
}

func (r *GormClusterRepository) DeleteAllUserClusters(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Cluster{}).Error
}

func (r *GormClusterRepository) CountAllClusters() (int64, error) {
	var count int64
	err := r.db.Model(&models.Cluster{}).Count(&count).Error
	return count, err
}

func (r *GormClusterRepository) ListAllClusters() ([]models.Cluster, error) {
	var clusters []models.Cluster
	return clusters, r.db.Find(&clusters).Error
}

// Voucher Repository

var _ models.VoucherRepository = (*GormVoucherRepository)(nil)

type GormVoucherRepository struct {
	db *gorm.DB
}

func NewGormVoucherRepository(db models.DB) models.VoucherRepository {
	return &GormVoucherRepository{db: db.GetDB()}
}

func (r *GormVoucherRepository) CreateVoucher(voucher *models.Voucher) error {
	return r.db.Create(voucher).Error
}

func (r *GormVoucherRepository) ListAllVouchers() ([]models.Voucher, error) {
	var vouchers []models.Voucher

	err := r.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

func (r *GormVoucherRepository) GetVoucherByCode(code string) (models.Voucher, error) {
	var voucher models.Voucher
	query := r.db.First(&voucher, "code = ?", code)

	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		return models.Voucher{}, models.ErrVoucherNotFound
	}

	return voucher, query.Error
}

func (r *GormVoucherRepository) RedeemVoucher(code string) error {
	result := r.db.Model(&models.Voucher{}).
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

// UserNodes Repository

var _ models.UserNodesRepository = (*GormUserNodesRepository)(nil)

type GormUserNodesRepository struct {
	db *gorm.DB
}

func NewGormUserNodesRepository(db models.DB) models.UserNodesRepository {
	return &GormUserNodesRepository{db: db.GetDB()}
}

func (r *GormUserNodesRepository) CreateUserNode(userNode *models.UserNodes) error {
	return r.db.Create(&userNode).Error
}

func (r *GormUserNodesRepository) DeleteUserNode(contractID uint64) error {
	return r.db.Where("contract_id = ?", contractID).Delete(&models.UserNodes{}).Error
}

func (r *GormUserNodesRepository) ListUserNodes(userID int) ([]models.UserNodes, error) {
	var userNodes []models.UserNodes
	return userNodes, r.db.Where("user_id = ?", userID).Find(&userNodes).Error
}

func (r *GormUserNodesRepository) ListAllReservedNodes() ([]models.UserNodes, error) {
	var userNodes []models.UserNodes
	return userNodes, r.db.Find(&userNodes).Error
}

func (r *GormUserNodesRepository) GetUserNodeByNodeID(nodeID uint64) (models.UserNodes, error) {
	var userNode models.UserNodes
	result := r.db.Where("node_id = ?", nodeID).First(&userNode)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.UserNodes{}, models.ErrUserNodeNotFound
	}

	return userNode, result.Error
}

func (r *GormUserNodesRepository) GetUserNodeByContractID(contractID uint64) (models.UserNodes, error) {
	var userNode models.UserNodes
	result := r.db.Where("contract_id = ?", contractID).First(&userNode)
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.UserNodes{}, models.ErrUserNodeNotFound
	}

	return userNode, result.Error
}

// Transaction Repository

var _ models.TransactionRepository = (*GormTransactionRepository)(nil)

type GormTransactionRepository struct {
	db *gorm.DB
}

func NewGormTransactionRepository(db models.DB) models.TransactionRepository {
	return &GormTransactionRepository{db: db.GetDB()}
}

func (r *GormTransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

// Settings Repository

var _ models.SettingsRepository = (*GormSettingsRepository)(nil)

type GormSettingsRepository struct {
	db *gorm.DB
}

func NewGormSettingsRepository(db models.DB) models.SettingsRepository {
	return &GormSettingsRepository{db: db.GetDB()}
}

const (
	maintenanceModeEnabled  = "enabled"
	maintenanceModeDisabled = "disabled"
)

func (r *GormSettingsRepository) GetSetting(name string) (string, error) {
	var setting models.Settings
	err := r.db.Where("name = ?", name).First(&setting).Error
	if err != nil {
		return "", err
	}

	return setting.Value, nil
}

func (r *GormSettingsRepository) SetSetting(name, value string) error {
	setting := models.Settings{
		Name:  name,
		Value: value,
	}

	return r.db.Save(&setting).Error
}

func (r *GormSettingsRepository) SetMaintenanceMode(enabled bool) error {
	value := maintenanceModeDisabled
	if enabled {
		value = maintenanceModeEnabled
	}
	return r.SetSetting("maintenance_mode", value)
}

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

// PendingRecord Repository

var _ models.PendingRecordRepository = (*GormPendingRecordRepository)(nil)

type GormPendingRecordRepository struct {
	db *gorm.DB
}

func NewGormPendingRecordRepository(db models.DB) models.PendingRecordRepository {
	return &GormPendingRecordRepository{db: db.GetDB()}
}

func (r *GormPendingRecordRepository) CreatePendingRecord(record *models.PendingRecord) error {
	record.CreatedAt = time.Now()
	return r.db.Create(record).Error
}

func (r *GormPendingRecordRepository) ListAllPendingRecords() ([]models.PendingRecord, error) {
	var pendingRecords []models.PendingRecord
	return pendingRecords, r.db.Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) ListOnlyPendingRecords() ([]models.PendingRecord, error) {
	var pendingRecords []models.PendingRecord
	return pendingRecords, r.db.Where("tft_amount > transferred_tft_amount").Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) ListUserPendingRecords(userID int) ([]models.PendingRecord, error) {
	var pendingRecords []models.PendingRecord
	return pendingRecords, r.db.Where("user_id = ?", userID).Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) UpdatePendingRecordTransferredAmount(id int, amount uint64) error {
	return r.db.Model(&models.PendingRecord{}).
		Where("id = ?", id).
		UpdateColumn("transferred_tft_amount", gorm.Expr("transferred_tft_amount + ?", amount)).
		UpdateColumn("updated_at", gorm.Expr("?", time.Now())).
		Error
}

// Notification Repository

var _ models.NotificationRepository = (*GormNotificationRepository)(nil)

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db models.DB) models.NotificationRepository {
	return &GormNotificationRepository{db: db.GetDB()}
}

func (r *GormNotificationRepository) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *GormNotificationRepository) GetUserNotifications(userID int, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *GormNotificationRepository) MarkNotificationAsRead(notificationID string, userID int) error {
	now := time.Now()
	result := r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": &now,
		})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.ErrNotificationNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrNotificationNotFound
	}

	return nil
}

func (r *GormNotificationRepository) MarkAllNotificationsAsRead(userID int) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, models.NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": &now,
		}).Error
}

func (r *GormNotificationRepository) DeleteNotification(notificationID string, userID int) error {
	result := r.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.ErrNotificationNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrNotificationNotFound
	}

	return nil
}

func (r *GormNotificationRepository) GetUnreadNotifications(userID int, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ? AND status = ?", userID, models.NotificationStatusUnread).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *GormNotificationRepository) DeleteAllNotifications(userID int) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Notification{}).Error
}

func (r *GormNotificationRepository) MarkNotificationAsUnread(notificationID string, userID int) error {
	result := r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusUnread,
			"read_at": nil,
		})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.ErrNotificationNotFound
		}

		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrNotificationNotFound
	}

	return nil
}

// Invoice Repository

var _ models.InvoiceRepository = (*GormInvoiceRepository)(nil)

type GormInvoiceRepository struct {
	db *gorm.DB
}

func NewGormInvoiceRepository(db models.DB) models.InvoiceRepository {
	return &GormInvoiceRepository{db: db.GetDB()}
}

func (r *GormInvoiceRepository) CreateInvoice(invoice *models.Invoice) error {
	return r.db.Create(&invoice).Error
}

func (r *GormInvoiceRepository) GetInvoice(id int) (models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.First(&invoice, id).Error
	if err != nil {
		return models.Invoice{}, err
	}

	var nodes []models.NodeItem
	if err = r.db.Model(&invoice).Association("Nodes").Find(&nodes); err != nil {
		return models.Invoice{}, err
	}

	invoice.Nodes = nodes
	return invoice, nil
}

func (r *GormInvoiceRepository) ListUserInvoices(userID int) ([]models.Invoice, error) {
	var invoices []models.Invoice
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

func (r *GormInvoiceRepository) ListInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice
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
	return r.db.Model(&models.Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"file_data": data}).Error
}
