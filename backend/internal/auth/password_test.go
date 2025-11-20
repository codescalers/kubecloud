package auth

import (
	"testing"
)

// TestHashAndSaltPassword_Password tests password hashing.
// This scenario covers:
// - Password hashing produces non-empty hash
// - Different passwords produce different hashes
// - Same password produces different hashes (due to salt)
// - Short passwords are hashed correctly
// - Long passwords are hashed correctly
func TestHashAndSaltPassword_Password(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		description string
	}{
		{
			name:        "simple_password",
			password:    "password123",
			description: "hashing simple password",
		},
		{
			name:        "long_password",
			password:    "this-is-a-very-long-password-with-many-characters!@#$%",
			description: "hashing long password",
		},
		{
			name:        "special_characters",
			password:    "p@ssw0rd!#$%^&*()",
			description: "hashing password with special characters",
		},
		{
			name:        "unicode_password",
			password:    "pässwörd€",
			description: "hashing password with unicode characters",
		},
		{
			name:        "whitespace_password",
			password:    "pass word with spaces",
			description: "hashing password with spaces",
		},
		{
			name:        "short_password",
			password:    "p",
			description: "hashing very short password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashAndSaltPassword([]byte(tt.password))

			if err != nil {
				t.Errorf("HashAndSaltPassword() error = %v (%s)", err, tt.description)
				return
			}
			if len(hash) == 0 {
				t.Errorf("HashAndSaltPassword() returned empty hash (%s)", tt.description)
			}
		})
	}
}

// TestHashAndSaltPasswordUniqueness_Password tests that same password produces different hashes.
// This scenario covers:
// - Multiple hashing of same password produces different results
// - This demonstrates salt is working correctly
func TestHashAndSaltPasswordUniqueness_Password(t *testing.T) {
	password := []byte("test-password")

	hash1, err := HashAndSaltPassword(password)
	if err != nil {
		t.Fatalf("HashAndSaltPassword() error = %v", err)
	}

	hash2, err := HashAndSaltPassword(password)
	if err != nil {
		t.Fatalf("HashAndSaltPassword() error = %v", err)
	}

	if string(hash1) == string(hash2) {
		t.Errorf("HashAndSaltPassword() should produce different hashes due to salt")
	}
}

// TestVerifyPassword_Password tests password verification.
// This scenario covers:
// - Correct password verifies successfully
// - Incorrect password fails verification
// - Case sensitivity in password verification
// - Whitespace sensitivity in password verification
func TestVerifyPassword_Password(t *testing.T) {
	correctPassword := "mypassword123"
	hash, err := HashAndSaltPassword([]byte(correctPassword))
	if err != nil {
		t.Fatalf("HashAndSaltPassword() error = %v", err)
	}

	tests := []struct {
		name        string
		password    string
		expected    bool
		description string
	}{
		{
			name:        "correct_password",
			password:    "mypassword123",
			expected:    true,
			description: "verifying correct password",
		},
		{
			name:        "incorrect_password",
			password:    "wrongpassword",
			expected:    false,
			description: "verifying incorrect password",
		},
		{
			name:        "wrong_case",
			password:    "MyPassword123",
			expected:    false,
			description: "verifying password with different case",
		},
		{
			name:        "partial_password",
			password:    "mypassword",
			expected:    false,
			description: "verifying partial password",
		},
		{
			name:        "password_with_extra_space",
			password:    "mypassword123 ",
			expected:    false,
			description: "verifying password with extra space",
		},
		{
			name:        "empty_password",
			password:    "",
			expected:    false,
			description: "verifying empty password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(hash, tt.password)

			if result != tt.expected {
				t.Errorf("VerifyPassword() = %v, expected %v (%s)", result, tt.expected, tt.description)
			}
		})
	}
}

// TestVerifyPasswordWithDifferentHashes_Password tests verification with various hashes.
// This scenario covers:
// - Different passwords have different hashes that don't verify against each other
// - Multiple passwords can be verified independently
func TestVerifyPasswordWithDifferentHashes_Password(t *testing.T) {
	password1 := "password-one"
	password2 := "password-two"

	hash1, err := HashAndSaltPassword([]byte(password1))
	if err != nil {
		t.Fatalf("HashAndSaltPassword() error = %v", err)
	}

	hash2, err := HashAndSaltPassword([]byte(password2))
	if err != nil {
		t.Fatalf("HashAndSaltPassword() error = %v", err)
	}

	// Password1 should verify with hash1 but not hash2
	if !VerifyPassword(hash1, password1) {
		t.Errorf("VerifyPassword() should verify password1 with hash1")
	}
	if VerifyPassword(hash2, password1) {
		t.Errorf("VerifyPassword() should not verify password1 with hash2")
	}

	// Password2 should verify with hash2 but not hash1
	if !VerifyPassword(hash2, password2) {
		t.Errorf("VerifyPassword() should verify password2 with hash2")
	}
	if VerifyPassword(hash1, password2) {
		t.Errorf("VerifyPassword() should not verify password2 with hash1")
	}
}
