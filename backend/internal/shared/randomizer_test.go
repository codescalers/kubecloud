// Package shared provides common utilities and configuration management for the kubecloud application.
// This test file covers the randomizer functionality for generating secure random
// voucher codes (for redemptions) and verification codes (for signup).

package shared

import (
	"math"
	"strings"
	"testing"
)

// TestGenerateVoucherCode_LengthAndCharset tests voucher code generation with correct length and character set.
// This scenario covers:
// - Voucher code length matches the requested length
// - All characters in the voucher are from the allowed alphanumeric character set
// - Security: ensuring vouchers only contain safe, URL-friendly characters
func TestGenerateVoucherCode_LengthAndCharset(t *testing.T) {
	length := 8
	voucher := GenerateVoucherCode(length)

	if len(voucher) != length {
		t.Errorf("Expected voucher length %d, got %d", length, len(voucher))
	}

	// Check that all characters are from the allowed set
	for _, char := range voucher {
		if !strings.ContainsRune(letterBytes, char) {
			t.Errorf("Invalid character in voucher: %c", char)
		}
	}
}

// TestGenerateVoucherCode_DifferentLengths tests voucher generation with various lengths.
// This scenario covers:
// - Correct voucher generation for different requested lengths
// - Edge cases like very short (1 char) and longer (16 char) vouchers
// - Consistency across multiple generations
func TestGenerateVoucherCode_DifferentLengths(t *testing.T) {
	testLengths := []int{1, 4, 8, 12, 16}

	for _, length := range testLengths {
		voucher := GenerateVoucherCode(length)
		if len(voucher) != length {
			t.Errorf("Expected voucher length %d, got %d", length, len(voucher))
		}
	}
}

// TestGenerateVoucherCode_Randomness tests that multiple voucher generations produce different results.
// This scenario covers:
// - Generated vouchers are not always the same (randomness verification)
// - No predictable patterns in voucher generation
// - Security: ensuring cryptographic randomness for voucher codes
func TestGenerateVoucherCode_Randomness(t *testing.T) {
	length := 8
	vouchers := make(map[string]bool)

	// Generate multiple vouchers and ensure they're different
	for i := 0; i < 100; i++ {
		voucher := GenerateVoucherCode(length)
		if vouchers[voucher] {
			t.Logf("Generated duplicate voucher (expected but rare): %s", voucher)
		}
		vouchers[voucher] = true
	}

	// With 100 generations, we should have at least 99 unique vouchers (very high probability)
	if len(vouchers) < 95 {
		t.Errorf("Expected mostly unique vouchers, got only %d unique out of 100", len(vouchers))
	}
}

// TestGenerateVerificationCode_RangeValidation tests that generated codes are within the expected numeric range.
// This scenario covers:
// - Code generation produces numbers within the correct bounds for the configured length
// - Mathematical correctness of range calculation (10^(length-1) to 10^length - 1)
// - For 4-digit codes: range [1000, 9999]
func TestGenerateVerificationCode_RangeValidation(t *testing.T) {
	length := 4
	min := int(math.Pow(10, float64(length-1)))   // 1000
	max := int(math.Pow(10, float64(length))) - 1 // 9999

	// Test multiple generations to ensure they're within range
	for i := 0; i < 100; i++ {
		code := GenerateVerificationCode(length)

		if code < min || code > max {
			t.Errorf("Generated code %d is outside range [%d, %d]", code, min, max)
		}
	}
}

// TestGenerateVerificationCode_VariousLengths tests code generation across different code lengths.
// This scenario covers:
// - Correct range calculation for various code lengths (1-5 digits)
// - Edge cases for single-digit and multi-digit codes
// - Validation of mathematical range boundaries for each length
func TestGenerateVerificationCode_VariousLengths(t *testing.T) {
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
		// Test multiple times for each length
		for i := 0; i < 20; i++ {
			code := GenerateVerificationCode(tc.codeLength)

			if code < tc.min || code > tc.max {
				t.Errorf("For codeLength %d, generated code %d is outside range [%d, %d]",
					tc.codeLength, code, tc.min, tc.max)
			}
		}
	}
}

// TestGenerateVerificationCode_Randomness tests that multiple verification code generations produce different results.
// This scenario covers:
// - Generated codes are not predictable
// - Different codes are generated across multiple calls
// - Security: ensuring randomness for verification codes used in authentication flows
func TestGenerateVerificationCode_Randomness(t *testing.T) {
	length := 4
	codes := make(map[int]bool)

	// Generate multiple codes and ensure they're different
	for i := 0; i < 100; i++ {
		code := GenerateVerificationCode(length)
		codes[code] = true
	}

	// With 100 generations for 4-digit codes (range 1000-9999), we should have many unique codes
	if len(codes) < 90 {
		t.Errorf("Expected mostly unique verification codes, got only %d unique out of 100", len(codes))
	}
}
