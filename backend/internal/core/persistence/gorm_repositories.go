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
	if err := r.db.Create(user).Error; err != nil {
		return err
	}

	return r.UpdateUserLastCalcTime(user.ID, time.Now())
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

func (r *GormUserRepository) DeductUserBalance(user *models.User, amount uint64) error {
	if user.CreditedBalance >= amount {
		return r.db.Model(&models.User{}).
			Where("id = ?", user.ID).
			UpdateColumn("credited_balance", gorm.Expr("credited_balance - ?", amount)).
			Error
	}

	if user.CreditedBalance > 0 && user.CreditCardBalance >= amount-user.CreditedBalance {
		return r.db.Model(&models.User{}).
			Where("id = ?", user.ID).
			UpdateColumn("credited_balance", gorm.Expr("credited_balance - ?", user.CreditedBalance)).
			UpdateColumn("credit_card_balance", gorm.Expr("credit_card_balance - ?", amount-user.CreditedBalance)).
			Error
	}

	if user.CreditCardBalance >= amount {
		return r.db.Model(&models.User{}).
			Where("id = ?", user.ID).
			UpdateColumn("credit_card_balance", gorm.Expr("credit_card_balance - ?", amount)).
			Error
	}

	// if credit card balance is not enough, add debt
	return r.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		UpdateColumn("debt", gorm.Expr("debt + ?", amount)).
		Error
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

// GetUserLastCalcTime returns the last calculation time for a user
func (r *GormUserRepository) GetUserLastCalcTime(userID int) (time.Time, error) {
	var calcTime models.UserUsageCalculationTime
	err := r.db.Where("user_id = ?", userID).First(&calcTime).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// If no record exists, return zero time
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return calcTime.LastCalcTime, nil
}

// UpdateUserLastCalcTime updates the last calculation time for a user
func (r *GormUserRepository) UpdateUserLastCalcTime(userID int, lastCalcTime time.Time) error {
	var calcTime models.UserUsageCalculationTime
	err := r.db.Where("user_id = ?", userID).First(&calcTime).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new record if one doesn't exist
			calcTime = models.UserUsageCalculationTime{
				UserID:       userID,
				LastCalcTime: lastCalcTime,
				UpdatedAt:    time.Now(),
			}
			return r.db.Create(&calcTime).Error
		}
		return err
	}

	// Update existing record
	calcTime.LastCalcTime = lastCalcTime
	calcTime.UpdatedAt = time.Now()
	return r.db.Save(&calcTime).Error
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

	clusterData, err := cluster.GetClusterResult()
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, node := range clusterData.Nodes {
			contractData := &models.UserContractData{
				UserID:     userID,
				ContractID: node.ContractID,
				NodeID:     node.NodeID,
				Type:       models.ContractTypeDeployed,
				CreatedAt:  time.Now(),
			}

			if err := tx.Create(contractData).Error; err != nil {
				return fmt.Errorf("failed to create contract data for node %d: %w", node.NodeID, err)
			}
		}

		return tx.Create(cluster).Error
	})
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

func (r *GormClusterRepository) UpdateCluster(contractsRepo models.ContractDataRepository, cluster *models.Cluster) error {
	existingCluster, err := r.GetClusterByName(cluster.UserID, cluster.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to get existing cluster: %w", err)
	}

	existingClusterData, err := existingCluster.GetClusterResult()
	if err != nil {
		return fmt.Errorf("failed to parse existing cluster data: %w", err)
	}

	newClusterData, err := cluster.GetClusterResult()
	if err != nil {
		return fmt.Errorf("failed to parse new cluster data: %w", err)
	}

	existingNodes := make(map[uint64]struct{})
	for _, node := range existingClusterData.Nodes {
		if node.ContractID != 0 {
			existingNodes[node.ContractID] = struct{}{}
		}
	}

	for _, node := range newClusterData.Nodes {
		if node.ContractID != 0 {
			if _, exists := existingNodes[node.ContractID]; !exists {
				// This is a new node, create a contract for it
				if err := contractsRepo.CreateUserContractData(
					&models.UserContractData{
						UserID:     cluster.UserID,
						ContractID: node.ContractID,
						NodeID:     node.NodeID,
						Type:       models.ContractTypeDeployed,
						CreatedAt:  time.Now(),
					},
				); err != nil {
					return fmt.Errorf("failed to create contract for new node: %w", err)
				}
			}
			// Remove from existing nodes map to track what is processed
			delete(existingNodes, node.ContractID)
		}
	}

	// Handle removed nodes - delete contracts for nodes that exist in old but not in new
	for contractID := range existingNodes {
		if err := contractsRepo.DeleteUserContract(contractID); err != nil {
			return fmt.Errorf("failed to delete contract for removed node: %w", err)
		}
	}

	// Update the cluster record
	cluster.UpdatedAt = time.Now()
	return r.db.Model(&models.Cluster{}).
		Where("user_id = ? AND project_name = ?", cluster.UserID, cluster.ProjectName).
		Updates(cluster).Error
}

func (r *GormClusterRepository) DeleteCluster(userID int, projectName string) error {
	cluster, err := r.GetClusterByName(userID, projectName)
	if err != nil {
		return err
	}

	clusterData, err := cluster.GetClusterResult()
	if err != nil {
		return err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		for _, node := range clusterData.Nodes {
			if err := tx.Model(&models.UserContractData{}).
				Where("contract_id = ?", node.ContractID).
				Update("deleted_at", time.Now()).Error; err != nil {
				return fmt.Errorf("failed to delete contract for node %d: %w", node.NodeID, err)
			}
		}

		return tx.Where("user_id = ? AND project_name = ?", userID, projectName).
			Delete(&models.Cluster{}).Error
	})

	return err
}

func (r *GormClusterRepository) DeleteAllUserClusters(contractsRepo models.ContractDataRepository, userID int) error {
	clusters, err := r.ListUserClusters(userID)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		clusterData, err := cluster.GetClusterResult()
		if err != nil {
			return err
		}

		for _, node := range clusterData.Nodes {
			if err := contractsRepo.DeleteUserContract(node.ContractID); err != nil {
				return err
			}
		}
	}

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

// UserContractData Repository

var _ models.ContractDataRepository = (*GormUserContractDataRepository)(nil)

type GormUserContractDataRepository struct {
	db *gorm.DB
}

func NewGormUserContractDataRepository(db models.DB) models.ContractDataRepository {
	return &GormUserContractDataRepository{db: db.GetDB()}
}

// CreateUserContractData creates new contract record for user
func (r *GormUserContractDataRepository) CreateUserContractData(contractData *models.UserContractData) error {
	return r.db.Create(&contractData).Error
}

// DeleteUserContract updates deleted time of a contract record for user by its contract ID
func (r *GormUserContractDataRepository) DeleteUserContract(contractID uint64) error {
	return r.db.Where("contract_id = ?", contractID).Update("deleted_at", time.Now()).Error
}

// ListUserRentedNodes returns all nodes records for user by its ID
func (r *GormUserContractDataRepository) ListUserRentedNodes(userID int) ([]models.UserContractData, error) {
	var userNodes []models.UserContractData
	return userNodes, r.db.Where("user_id = ? and type = ? and deleted_at = ?", userID, models.ContractTypeRented, time.Time{}).Find(&userNodes).Error
}

// ListAllContractsInPeriod returns all contracts that existed during the specified time period.
// This includes:
// 1. Contracts created before or during the period end date
// 2. AND either not deleted (deleted_at is zero time) OR deleted after the period start date
// If userID is provided (non-zero), it will only return contracts for that specific user.
// If userID is 0, it will return contracts for all users.
func (r *GormUserContractDataRepository) ListAllContractsInPeriod(userID int, start, end time.Time) ([]models.UserContractData, error) {
	var userNodes []models.UserContractData

	// Query for contracts that:
	// - Were created on or before the end date of the period
	// - AND are either not deleted (deleted_at is zero) OR were deleted after the start of the period
	query := r.db.Where("created_at <= ?", end).
		Where("(deleted_at = ? OR deleted_at >= ?)", time.Time{}, start)

	// If userID is provided (non-zero), filter by that user
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	return userNodes, query.Find(&userNodes).Error
}

// ListAllReservedNodes returns all reserved nodes from all users
func (r *GormUserContractDataRepository) ListAllReservedNodes() ([]models.UserContractData, error) {
	var userNodes []models.UserContractData
	return userNodes, r.db.Where("type = ? and deleted_at = ?", models.ContractTypeRented, time.Time{}).Find(&userNodes).Error
}

func (r *GormUserContractDataRepository) GetUserNodeByNodeID(nodeID uint64) (models.UserContractData, error) {
	var userNode models.UserContractData
	return userNode, r.db.Where("node_id = ? and deleted_at = ?", nodeID, time.Time{}).First(&userNode).Error
}

func (r *GormUserContractDataRepository) GetUserNodeByContractID(contractID uint64) (models.UserContractData, error) {
	var userNode models.UserContractData
	return userNode, r.db.Where("contract_id = ? and deleted_at = ?", contractID, time.Time{}).First(&userNode).Error
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

// TransferRecord Repository

var _ models.TransferRecordRepository = (*GormTransferRecordRepository)(nil)

type GormTransferRecordRepository struct {
	db *gorm.DB
}

func NewGormTransferRecordRepository(db models.DB) models.TransferRecordRepository {
	return &GormTransferRecordRepository{db: db.GetDB()}
}

func (r *GormTransferRecordRepository) CreateTransferRecord(record *models.TransferRecord) error {
	record.CreatedAt = time.Now()
	return r.db.Create(record).Error
}
func (r *GormTransferRecordRepository) ListTransferRecords() ([]models.TransferRecord, error) {
	var TransferRecords []models.TransferRecord
	return TransferRecords, r.db.Find(&TransferRecords).Error
}

func (r *GormTransferRecordRepository) CalculateTotalPendingTFTAmountPerUser(userID int) (uint64, error) {
	var totalAmount uint64
	err := r.db.Model(&models.TransferRecord{}).
		Select("COALESCE(SUM(tft_amount), 0)").
		Where("user_id = ? AND state = ?", userID, models.PendingState).
		Scan(&totalAmount).Error
	if err != nil {
		return 0, err
	}
	return totalAmount, nil
}

func (r *GormTransferRecordRepository) ListUserTransferRecords(userID int) ([]models.TransferRecord, error) {
	var TransferRecords []models.TransferRecord
	return TransferRecords, r.db.Where("user_id = ?", userID).Find(&TransferRecords).Error
}

func (r *GormTransferRecordRepository) ListPendingTransferRecords() ([]models.TransferRecord, error) {
	var TransferRecords []models.TransferRecord
	return TransferRecords, r.db.Where("state = ?", models.PendingState).Find(&TransferRecords).Error
}

func (r *GormTransferRecordRepository) ListFailedTransferRecords() ([]models.TransferRecord, error) {
	var TransferRecords []models.TransferRecord
	return TransferRecords, r.db.Where("state = ?", models.FailedState).Find(&TransferRecords).Error
}

func (r *GormTransferRecordRepository) UpdateTransferRecordState(recordID int, state models.State, failure string) error {
	return r.db.Model(&models.TransferRecord{}).Where("id = ?", recordID).Updates(
		map[string]interface{}{"state": state, "failure": failure, "updated_at": time.Now()}).Error
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
