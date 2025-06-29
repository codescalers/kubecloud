package internal

import (
	"math/rand"
	"net/mail"
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

// isValidEmail validates an email address using the standard library
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
