package models

// UserRepository defines operations for user data persistence
type UserRepository interface {
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(userID int) (User, error)
	UpdateUserByID(user *User) error
	ListAllUsers() ([]User, error)
	ListAdmins() ([]User, error)
	DeleteUserByID(userID int) error
	CreditUserBalance(userID int, amount uint64) error
	ListRemainingWorkflowsByUserID(userID int) ([]GormWorkflowRecord, error)

	// stats methods
	CountAllUsers() (int64, error)

	// SSH Key methods
	CreateSSHKey(sshKey *SSHKey) error
	ListUserSSHKeys(userID int) ([]SSHKey, error)
	DeleteSSHKey(sshKeyID int, userID int) (string, error)
	GetSSHKeyByID(sshKeyID int, userID int) (SSHKey, error)
}

// ClusterRepository defines operations for cluster data persistence
type ClusterRepository interface {
	CreateCluster(userID int, cluster *Cluster) error
	ListUserClusters(userID int) ([]Cluster, error)
	GetClusterByName(userID int, projectName string) (Cluster, error)
	UpdateCluster(cluster *Cluster) error
	DeleteCluster(userID int, projectName string) error
	DeleteAllUserClusters(userID int) error

	// stats methods
	CountAllClusters() (int64, error)
	ListAllClusters() ([]Cluster, error)
}

// UserNodesRepository defines operations for user nodes data persistence
type UserNodesRepository interface {
	CreateUserNode(userNode *UserNodes) error
	DeleteUserNode(contractID uint64) error
	ListUserNodes(userID int) ([]UserNodes, error)
	GetUserNodeByNodeID(nodeID uint64) (UserNodes, error)
	GetUserNodeByContractID(contractID uint64) (UserNodes, error)
	ListAllReservedNodes() ([]UserNodes, error)
}

// VoucherRepository defines operations for voucher data persistence
type VoucherRepository interface {
	CreateVoucher(voucher *Voucher) error
	ListAllVouchers() ([]Voucher, error)
	GetVoucherByCode(code string) (Voucher, error)
	RedeemVoucher(code string) error
}

// TransactionRepository defines operations for transaction data persistence
type TransactionRepository interface {
	CreateTransaction(transaction *Transaction) error
}

// InvoiceRepository defines operations for invoice data persistence
type InvoiceRepository interface {
	CreateInvoice(invoice *Invoice) error
	GetInvoice(id int) (Invoice, error)
	ListUserInvoices(userID int) ([]Invoice, error)
	ListInvoices() ([]Invoice, error)
	UpdateInvoicePDF(id int, data []byte) error
}

// NotificationRepository defines operations for notification data persistence
type NotificationRepository interface {
	CreateNotification(notification *Notification) error
	GetUserNotifications(userID int, limit, offset int) ([]Notification, error)
	GetUnreadNotifications(userID int, limit, offset int) ([]Notification, error)
	MarkNotificationAsRead(notificationID string, userID int) error
	MarkNotificationAsUnread(notificationID string, userID int) error
	MarkAllNotificationsAsRead(userID int) error
	DeleteNotification(notificationID string, userID int) error
	DeleteAllNotifications(userID int) error
}

// PendingRecordRepository defines operations for pending record data persistence
type PendingRecordRepository interface {
	CreatePendingRecord(record *PendingRecord) error
	ListAllPendingRecords() ([]PendingRecord, error)
	ListOnlyPendingRecords() ([]PendingRecord, error)
	ListUserPendingRecords(userID int) ([]PendingRecord, error)
	UpdatePendingRecordTransferredAmount(id int, amount uint64) error
}

// SettingsRepository defines operations for settings data persistence
type SettingsRepository interface {
	GetSetting(name string) (string, error)
	SetSetting(name, value string) error
	SetMaintenanceMode(enabled bool) error
	GetMaintenanceMode() (bool, error)
}
