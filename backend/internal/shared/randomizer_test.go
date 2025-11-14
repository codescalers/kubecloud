package shared

import (
	"math"
	"strings"
	"testing"
)

func TestNewRandomizer(t *testing.T) {
	r := NewRandomizer(10, 4)
	if r.voucherLength != 10 {
		t.Errorf("Expected voucherLength 10, got %d", r.voucherLength)
	}
	if r.codeLength != 4 {
		t.Errorf("Expected codeLength 4, got %d", r.codeLength)
	}
}

func TestGenerateRandomVoucher(t *testing.T) {
	r := NewRandomizer(8, 4)

	voucher := r.GenerateRandomVoucher()

	if len(voucher) != 8 {
		t.Errorf("Expected voucher length 8, got %d", len(voucher))
	}

	// Check that all characters are from the allowed set
	for _, char := range voucher {
		if !strings.ContainsRune(letterBytes, char) {
			t.Errorf("Invalid character in voucher: %c", char)
		}
	}
}

func TestGenerateRandomCode(t *testing.T) {
	r := NewRandomizer(10, 4)

	// Test multiple generations to ensure they're within range
	for i := 0; i < 100; i++ {
		code := r.GenerateRandomCode()
		min := int(math.Pow(10, float64(4-1)))   // 1000
		max := int(math.Pow(10, float64(4))) - 1 // 9999

		if code < min || code > max {
			t.Errorf("Generated code %d is outside range [%d, %d]", code, min, max)
		}
	}
}

func TestGenerateRandomCodeDifferentLengths(t *testing.T) {
	testCases := []struct {
		codeLength int
		min        int
		max        int
	}{
		{1, 0, 9},
		{2, 10, 99},
		{3, 100, 999},
		{4, 1000, 9999},
		{5, 10000, 99999},
	}

	for _, tc := range testCases {
		r := NewRandomizer(10, tc.codeLength)
		code := r.GenerateRandomCode()

		if code < tc.min || code > tc.max {
			t.Errorf("For codeLength %d, generated code %d is outside range [%d, %d]",
				tc.codeLength, code, tc.min, tc.max)
		}
	}
}

func TestRandomizerInterface(t *testing.T) {
	var r Randomizer = NewRandomizer(5, 3)

	// Test that it implements the interface
	voucher := r.GenerateRandomVoucher()
	if len(voucher) != 5 {
		t.Errorf("Expected voucher length 5, got %d", len(voucher))
	}

	code := r.GenerateRandomCode()
	if code < 100 || code > 999 {
		t.Errorf("Generated code %d is outside expected range [100, 999]", code)
	}
}
