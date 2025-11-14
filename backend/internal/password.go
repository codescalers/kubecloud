package internal

import (
	"golang.org/x/crypto/bcrypt"
)

// HashAndSaltPassword hashes given password and append salt to it
func HashAndSaltPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

}

// VerifyPassword checks if given password is same as hashed one
func VerifyPassword(hashedPassword []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil

}
