package shared

import (
	"math"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Randomizer interface {
	GenerateRandomVoucher() string
	GenerateRandomCode() int
}

type randomizer struct {
	voucherLength int
	codeLength    int
}

func NewRandomizer(voucherLength int, codeLength int) randomizer {
	return randomizer{
		voucherLength: voucherLength,
		codeLength:    codeLength,
	}
}

// GenerateRandomVoucher generates a random voucher
func (r randomizer) GenerateRandomVoucher() string {
	b := make([]byte, r.voucherLength)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GenerateRandomCode generates random code of 4 digits
func (r randomizer) GenerateRandomCode() int {
	min := int(math.Pow(10, float64(r.codeLength-1)))
	max := int(math.Pow(10, float64(r.codeLength)) - 1)
	return rand.Intn(max-min) + min
}
