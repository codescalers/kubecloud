package internal

import (
	"crypto/rand"
	"math/big"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// IsAdmin checks if provided email in admins list or not
func IsAdmin(email string, admins []string) bool {
	for _, adminEmail := range admins {
		if email == adminEmail {
			return true
		}
	}
	return false
}

// GenerateRandomVoucher generates a cryptographically secure random voucher
func GenerateRandomVoucher(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		b[i] = letterBytes[num.Int64()]
	}
	return string(b)
}
