package internal

import (
	"math/rand"
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

// GenerateRandomVoucher generates a random voucher
func GenerateRandomVoucher(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GenerateRandomCode generates random code of 4 digits
func GenerateRandomCode() int {
	min := 1000
	max := 9999
	return rand.Intn(max-min) + min
}
