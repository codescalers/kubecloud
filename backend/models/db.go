package models

import (
	"context"
	"fmt"
	"kubecloud/internal/utils"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

// DB interface for databases
type DB interface {
	Ping(ctx context.Context) error
	Close() error
	GetDB() *gorm.DB
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(userID int) (User, error)
	UpdateUserByID(user *User) error
	UpdatePassword(email string, hashedPassword []byte) error
	ListAllUsers() ([]User, error)
	ListAdmins() ([]User, error)
	DeleteUserByID(userID int) error
	CreateVoucher(voucher *Voucher) error
	ListAllVouchers() ([]Voucher, error)
	GetVoucherByCode(code string) (Voucher, error)
	RedeemVoucher(code string) error
	CreateTransaction(transaction *Transaction) error
	CreditUserBalance(userID int, amount uint64) error
	CreateInvoice(invoice *Invoice) error
	GetInvoice(id int) (Invoice, error)
	ListUserInvoices(userID int) ([]Invoice, error)
	ListInvoices() ([]Invoice, error)
	UpdateInvoicePDF(id int, data []byte) error
	CreateUserNode(userNode *UserNodes) error
	DeleteUserNode(contractID uint64) error
	ListUserNodes(userID int) ([]UserNodes, error)
	GetUserNodeByNodeID(nodeID uint64) (UserNodes, error)
	GetUserNodeByContractID(contractID uint64) (UserNodes, error)
	ListAllReservedNodes() ([]UserNodes, error)
	// SSH Key methods
	CreateSSHKey(sshKey *SSHKey) error
	ListUserSSHKeys(userID int) ([]SSHKey, error)
	DeleteSSHKey(sshKeyID int, userID int) error
	GetSSHKeyByID(sshKeyID int, userID int) (SSHKey, error)
	// Notification methods
	CreateNotification(notification *Notification) error
	GetUserNotifications(userID int, limit, offset int) ([]Notification, error)
	GetUnreadNotifications(userID int, limit, offset int) ([]Notification, error)
	MarkNotificationAsRead(notificationID string, userID int) error
	MarkNotificationAsUnread(notificationID string, userID int) error
	MarkAllNotificationsAsRead(userID int) error
	DeleteNotification(notificationID string, userID int) error
	DeleteAllNotifications(userID int) error
	// Cluster methods
	CreateCluster(userID int, cluster *Cluster) error
	ListUserClusters(userID int) ([]Cluster, error)
	GetClusterByName(userID int, projectName string) (Cluster, error)
	UpdateCluster(cluster *Cluster) error
	DeleteCluster(userID int, projectName string) error
	DeleteAllUserClusters(userID int) error
	// pending records methods
	CreatePendingRecord(record *PendingRecord) error
	ListAllPendingRecords() ([]PendingRecord, error)
	ListOnlyPendingRecords() ([]PendingRecord, error)
	ListUserPendingRecords(userID int) ([]PendingRecord, error)
	UpdatePendingRecordTransferredAmount(id int, amount uint64) error
	// stats methods
	CountAllUsers() (int64, error)
	CountAllClusters() (int64, error)
	ListAllClusters() ([]Cluster, error)
}

func NewDB(dsn string, cfg DBPoolConfig) (DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	switch u.Scheme {
	case "postgres":
		return NewPostgresDB(dsn, cfg)
	case "sqlite", "sqlite3":
		path, err := utils.ExpandPath(u.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path: %w", err)
		}
		return NewSqliteDB(path)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}
