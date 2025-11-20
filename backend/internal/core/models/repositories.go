package models

import "time"

// UserRepository defines operations for user data persistence
type UserRepository interface {
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(userID int) (User, error)
	UpdateUserByID(user *User) error
	DeductUserBalance(user *User, amount uint64) error
	ListAllUsers() ([]User, error)
	ListAdmins() ([]User, error)
	DeleteUserByID(userID int) error
	CreditUserBalance(userID int, amount uint64) error

	// stats methods
	CountAllUsers() (int64, error)

	// Usage calculation time methods
	GetUserLastCalcTime(userID int) (time.Time, error)
	UpdateUserLastCalcTime(userID int, lastCalcTime time.Time) error

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
	UpdateCluster(contractsRepo ContractDataRepository, cluster *Cluster) error
	DeleteCluster(userID int, projectName string) error
	DeleteAllUserClusters(contractsRepo ContractDataRepository, userID int) error

	// stats methods
	CountAllClusters() (int64, error)
	ListAllClusters() ([]Cluster, error)
}

// ContractDataRepository defines operations for contract data persistence
type ContractDataRepository interface {
	CreateUserContractData(contractData *UserContractData) error
	DeleteUserContract(contractID uint64) error
	ListUserRentedNodes(userID int) ([]UserContractData, error)
	GetUserNodeByNodeID(nodeID uint64) (UserContractData, error)
	GetUserNodeByContractID(contractID uint64) (UserContractData, error)
	ListAllReservedNodes() ([]UserContractData, error)
	ListAllContractsInPeriod(userID int, start, end time.Time) ([]UserContractData, error)
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

// TransferRecordRepository defines operations for transfer record data persistence
type TransferRecordRepository interface {
	CreateTransferRecord(record *TransferRecord) error
	ListTransferRecords() ([]TransferRecord, error)
	ListUserTransferRecords(userID int) ([]TransferRecord, error)
	ListPendingTransferRecords() ([]TransferRecord, error)
	ListFailedTransferRecords() ([]TransferRecord, error)
	UpdateTransferRecordState(recordID int, state State, failure string) error
	CalculateTotalPendingTFTAmountPerUser(userID int) (uint64, error)
}

// SettingsRepository defines operations for settings data persistence
type SettingsRepository interface {
	GetSetting(name string) (string, error)
	SetSetting(name, value string) error
	SetMaintenanceMode(enabled bool) error
	GetMaintenanceMode() (bool, error)
}
