package models

// DB interface for databases
type DB interface {
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
}
