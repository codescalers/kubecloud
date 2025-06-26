package models

// DB interface for databases
type DB interface {
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(userID int) (User, error)
	UpdateUserByID(user *User) error
	UpdatePassword(email string, hashedPassword []byte) error
	UpdateUserVerification(userID int, verified bool) error
	ListAllUsers() ([]User, error)
	DeleteUserByID(userID int) error
	CreateVoucher(voucher *Voucher) error
	ListAllVouchers() ([]Voucher, error)
	CreateTransaction(transaction *Transaction) error
	CreditUserBalance(userID int, amount float64) error
}
