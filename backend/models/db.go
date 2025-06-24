package models

// DB interface for databases
type DB interface {
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(ID string) (User, error)
	UpdateUserByID(user *User) error
	UpdatePassword(email, hashedPassword, salt string) error
	ListAllUsers() ([]User, error)
	DeleteUserByID(userID string) error
	CreateVoucher(voucher *Voucher) error
	ListAllVouchers() ([]Voucher, error)
	CreateTransaction(transaction *Transaction) error
	CreditUserBalance(userID int, amount float64) error
}
