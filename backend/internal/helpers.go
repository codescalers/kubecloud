package internal

import (
	"math/rand"
	"net/mail"

	"golang.org/x/crypto/ssh"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Contains[T comparable](elements []T, element T) bool {
	for _, e := range elements {
		if element == e {
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

// IsValidEmail validates an email address using the standard library
func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidateSSH used for validating ssh keys
func ValidateSSH(sshKey string) error {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sshKey))
	return err
}
