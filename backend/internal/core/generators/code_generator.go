package generators

import (
	"math"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateVoucherCode generates a random alphanumeric voucher code of the specified length
// that users can redeem. For example, a 8-character voucher code: "aB3xK9mL"
func GenerateVoucherCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GenerateVerificationCode generates a random numeric code of the specified length
// for user email/phone verification during signup. For example, a 4-digit code: 7392
func GenerateVerificationCode(length int) int {
	min := int(math.Pow(10, float64(length-1)))
	max := int(math.Pow(10, float64(length)) - 1)
	return rand.Intn(max-min) + min
}
