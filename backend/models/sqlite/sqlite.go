package sqlite

import (
	"database/sql"
	"fmt"
	"kubecloud/models"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Sqlite struct implements db interface with sqlite
type Sqlite struct {
	db    *gorm.DB
	mutex sync.Mutex
}

// SQLDB returns the underlying *sql.DB for health checks
func (s *Sqlite) SQLDB() (*sql.DB, error) {
	return s.db.DB()
}

// NewSqliteStorage connects to the database file
func NewSqliteStorage(file string) (*Sqlite, error) {
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate models
	err = db.AutoMigrate(&models.User{}, &models.Voucher{}, models.Transaction{}, models.Invoice{}, models.NodeItem{}, models.UserNodes{}, &models.Notification{}, &models.SSHKey{}, &models.Cluster{})
	if err != nil {
		return nil, err
	}

	return &Sqlite{db: db}, nil
}

// Close closes the database connection
func (s *Sqlite) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RegisterUser registers a new user to the system
func (s *Sqlite) RegisterUser(user *models.User) error {
	return s.db.Create(user).Error
}

// GetUserByEmail returns user by its email if found
func (s *Sqlite) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := s.db.First(&user, "email = ?", email)
	return user, query.Error
}

// GetUserByEmail returns user by its email if found
func (s *Sqlite) GetUserByID(userID int) (models.User, error) {
	var user models.User
	query := s.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (s *Sqlite) UpdateUserByID(user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
}

// UpdatePassword updates password of user by its email
func (s *Sqlite) UpdatePassword(email string, hashedPassword []byte) error {
	result := s.db.Model(&models.User{}).
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

// UpdateUserVerification updates verified status of user by its ID
func (s *Sqlite) UpdateUserVerification(userID int, verified bool) error {
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("verified", verified)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with ID %d", userID)
	}

	return nil
}

// ListAllUsers lists all users in system
func (s *Sqlite) ListAllUsers() ([]models.User, error) {
	var users []models.User

	err := s.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil

}

// DeleteUserByID deletes user by its ID
func (s *Sqlite) DeleteUserByID(userID int) error {
	return s.db.Where("id = ?", userID).Delete(&models.User{}).Error
}

// CreateVoucher creates new voucher in system
func (s *Sqlite) CreateVoucher(voucher *models.Voucher) error {
	return s.db.Create(voucher).Error
}

// ListAllVouchers gets all vouchers in system
func (s *Sqlite) ListAllVouchers() ([]models.Voucher, error) {
	var vouchers []models.Voucher

	err := s.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

// GetVoucherByCode returns voucher by its code
func (s *Sqlite) GetVoucherByCode(code string) (models.Voucher, error) {
	var voucher models.Voucher
	query := s.db.First(&voucher, "code = ?", code)
	return voucher, query.Error
}

// RedeemVoucher updates status if voucher
func (s *Sqlite) RedeemVoucher(code string) error {
	result := s.db.Model(&models.Voucher{}).
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
func (s *Sqlite) CreateTransaction(transaction *models.Transaction) error {
	return s.db.Create(transaction).Error
}

// CreditUserBalance add credited balance to user by its ID
func (s *Sqlite) CreditUserBalance(userID int, amount uint64) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}

// CreateInvoice creates new invoice
func (s *Sqlite) CreateInvoice(invoice *models.Invoice) error {
	return s.db.Create(&invoice).Error
}

// GetInvoice returns an invoice by ID
func (s *Sqlite) GetInvoice(id int) (models.Invoice, error) {
	var invoice models.Invoice
	err := s.db.First(&invoice, id).Error
	if err != nil {
		return models.Invoice{}, err
	}

	var nodes []models.NodeItem
	if err = s.db.Model(&invoice).Association("Nodes").Find(&nodes); err != nil {
		return models.Invoice{}, err
	}

	invoice.Nodes = nodes
	return invoice, nil
}

// ListUserInvoices returns all invoices of user
func (s *Sqlite) ListUserInvoices(userID int) ([]models.Invoice, error) {
	var invoices []models.Invoice
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
func (s *Sqlite) ListInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice
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

func (s *Sqlite) UpdateInvoicePDF(id int, data []byte) error {
	return s.db.Model(&models.Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"file_data": data}).Error
}

// CreateUserNode creates new node record for user
func (s *Sqlite) CreateUserNode(userNode *models.UserNodes) error {
	return s.db.Create(&userNode).Error
}

// ListUserNodes returns all nodes records for user by its ID
func (s *Sqlite) ListUserNodes(userID int) ([]models.UserNodes, error) {
	var userNodes []models.UserNodes
	return userNodes, s.db.Where("user_id = ?", userID).Find(&userNodes).Error
}

// CreateNotification creates a new notification
func (s *Sqlite) CreateNotification(notification *models.Notification) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.db.Create(notification).Error
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *Sqlite) GetUserNotifications(userID string, limit, offset int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkNotificationAsRead marks a specific notification as read
func (s *Sqlite) MarkNotificationAsRead(notificationID uint, userID string) error {
	now := time.Now()
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or access denied")
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func (s *Sqlite) MarkAllNotificationsAsRead(userID string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, models.NotificationStatusUnread).
		Updates(map[string]interface{}{
			"status":  models.NotificationStatusRead,
			"read_at": &now,
		}).Error
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func (s *Sqlite) GetUnreadNotificationCount(userID string) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, models.NotificationStatusUnread).
		Count(&count).Error
	return count, err
}

// DeleteNotification deletes a notification for a user
func (s *Sqlite) DeleteNotification(notificationID uint, userID string) error {
	return s.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{}).Error
}

// CreateSSHKey creates a new SSH key for a user
func (s *Sqlite) CreateSSHKey(sshKey *models.SSHKey) error {
	sshKey.CreatedAt = time.Now()
	sshKey.UpdatedAt = time.Now()
	return s.db.Create(sshKey).Error
}

// ListUserSSHKeys returns all SSH keys for a user
func (s *Sqlite) ListUserSSHKeys(userID int) ([]models.SSHKey, error) {
	var sshKeys []models.SSHKey
	err := s.db.Where("user_id = ?", userID).Find(&sshKeys).Error
	if err != nil {
		return nil, err
	}
	return sshKeys, nil
}

// DeleteSSHKey deletes an SSH key by ID for a specific user
func (s *Sqlite) DeleteSSHKey(sshKeyID int, userID int) error {
	result := s.db.Where("id = ? AND user_id = ?", sshKeyID, userID).Delete(&models.SSHKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no SSH key found with ID %d for user %d", sshKeyID, userID)
	}
	return nil
}

// GetSSHKeyByID returns an SSH key by ID for a specific user
func (s *Sqlite) GetSSHKeyByID(sshKeyID int, userID int) (models.SSHKey, error) {
	var sshKey models.SSHKey
	query := s.db.Where("id = ? AND user_id = ?", sshKeyID, userID).First(&sshKey)
	return sshKey, query.Error
}

// CreateCluster creates a new cluster in the database
func (s *Sqlite) CreateCluster(cluster *models.Cluster) error {
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	return s.db.Create(cluster).Error
}

// ListUserClusters returns all clusters for a specific user
func (s *Sqlite) ListUserClusters(userID string) ([]models.Cluster, error) {
	var clusters []models.Cluster
	query := s.db.Where("user_id = ?", userID).Find(&clusters)
	return clusters, query.Error
}

// GetClusterByName returns a cluster by name for a specific user
func (s *Sqlite) GetClusterByName(userID string, projectName string) (models.Cluster, error) {
	var cluster models.Cluster
	query := s.db.Where("user_id = ? AND project_name = ?", userID, projectName).First(&cluster)
	return cluster, query.Error
}

// UpdateCluster updates an existing cluster
func (s *Sqlite) UpdateCluster(cluster *models.Cluster) error {
	cluster.UpdatedAt = time.Now()
	return s.db.Model(&models.Cluster{}).
		Where("user_id = ? AND project_name = ?", cluster.UserID, cluster.ProjectName).
		Updates(cluster).Error
}

// DeleteCluster deletes a cluster by name for a specific user
func (s *Sqlite) DeleteCluster(userID string, projectName string) error {
	return s.db.Where("user_id = ? AND project_name = ?", userID, projectName).Delete(&models.Cluster{}).Error
}

// GetDBStats returns open and idle connection counts for metrics
func (s *Sqlite) GetDBStats() (open int, idle int, err error) {
	sqldb, err := s.SQLDB()
	if err != nil {
		return 0, 0, err
	}
	stats := sqldb.Stats()
	return stats.OpenConnections, stats.Idle, nil
}
