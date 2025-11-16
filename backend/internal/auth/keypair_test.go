package auth

import (
	"testing"
)

// TestValidateMnemonic tests mnemonic validation.
// This scenario covers:
// - Valid 12-word mnemonic passes validation
// - Empty mnemonic fails validation
// - Whitespace-only mnemonic fails validation
// - Mnemonic with fewer than 12 words fails validation
// - Mnemonic with extra whitespace is handled correctly
func TestValidateMnemonic(t *testing.T) {
	tests := []struct {
		name        string
		mnemonic    string
		expectError bool
		errorMsg    string
		description string
	}{
		{
			name:        "valid_12_word_mnemonic",
			mnemonic:    "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			expectError: false,
			description: "validating standard 12-word mnemonic",
		},
		{
			name:        "valid_mnemonic_with_spaces",
			mnemonic:    "  abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about  ",
			expectError: false,
			description: "validating mnemonic with trimmed whitespace",
		},
		{
			name:        "empty_mnemonic",
			mnemonic:    "",
			expectError: true,
			errorMsg:    "mnemonic cannot be empty",
			description: "validating empty mnemonic",
		},
		{
			name:        "whitespace_only",
			mnemonic:    "   ",
			expectError: true,
			errorMsg:    "mnemonic cannot be empty",
			description: "validating whitespace-only mnemonic",
		},
		{
			name:        "too_few_words",
			mnemonic:    "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
			expectError: true,
			errorMsg:    "mnemonic must be at least 12 words",
			description: "validating mnemonic with 10 words",
		},
		{
			name:        "single_word",
			mnemonic:    "abandon",
			expectError: true,
			errorMsg:    "mnemonic must be at least 12 words",
			description: "validating single-word mnemonic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMnemonic(tt.mnemonic)

			if (err != nil) != tt.expectError {
				t.Errorf("validateMnemonic() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
				return
			}

			if tt.expectError && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("validateMnemonic() error message = %q, want %q (%s)", err.Error(), tt.errorMsg, tt.description)
			}
		})
	}
}

// TestKeyPairFromMnemonic tests keypair derivation from mnemonic.
// This scenario covers:
// - Valid mnemonic derives keypair successfully
// - Invalid mnemonic fails
// - Empty mnemonic fails
// - Same mnemonic always produces same keypair
func TestKeyPairFromMnemonic(t *testing.T) {
	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	tests := []struct {
		name        string
		mnemonic    string
		expectError bool
		description string
	}{
		{
			name:        "valid_mnemonic",
			mnemonic:    validMnemonic,
			expectError: false,
			description: "deriving keypair from valid mnemonic",
		},
		{
			name:        "empty_mnemonic",
			mnemonic:    "",
			expectError: true,
			description: "deriving keypair from empty mnemonic",
		},
		{
			name:        "too_few_words",
			mnemonic:    "abandon abandon abandon",
			expectError: true,
			description: "deriving keypair from mnemonic with too few words",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair, err := KeyPairFromMnemonic(tt.mnemonic)

			if (err != nil) != tt.expectError {
				t.Errorf("KeyPairFromMnemonic() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
				return
			}

			if !tt.expectError {
				if keyPair == nil {
					t.Errorf("KeyPairFromMnemonic() returned nil keypair (%s)", tt.description)
				}
			}
		})
	}
}

// TestKeyPairFromMnemonicConsistency tests that same mnemonic produces same keypair.
// This scenario covers:
// - Multiple derivations of same mnemonic produce identical keypairs
// - Keypair has consistent public key
func TestKeyPairFromMnemonicConsistency(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	keyPair1, err := KeyPairFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("KeyPairFromMnemonic() error = %v", err)
	}

	keyPair2, err := KeyPairFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("KeyPairFromMnemonic() error = %v", err)
	}

	// Get SS58 addresses to compare
	addr1, err := AccountAddressFromKeypair(keyPair1)
	if err != nil {
		t.Fatalf("AccountAddressFromKeypair() error = %v", err)
	}

	addr2, err := AccountAddressFromKeypair(keyPair2)
	if err != nil {
		t.Fatalf("AccountAddressFromKeypair() error = %v", err)
	}

	if addr1 != addr2 {
		t.Errorf("KeyPairFromMnemonic() produces inconsistent addresses: %s vs %s", addr1, addr2)
	}
}

// TestAccountAddressFromKeypair tests SS58 address derivation.
// This scenario covers:
// - Keypair derives valid SS58 address
// - Address starts with character for format 42
// - Address is non-empty string
func TestAccountAddressFromKeypair(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	keyPair, err := KeyPairFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("KeyPairFromMnemonic() error = %v", err)
	}

	address, err := AccountAddressFromKeypair(keyPair)

	if err != nil {
		t.Errorf("AccountAddressFromKeypair() error = %v", err)
		return
	}
	if address == "" {
		t.Errorf("AccountAddressFromKeypair() returned empty address")
	}
	// SS58 format 42 addresses are valid substrate addresses (they contain letters and numbers)
	if len(address) < 47 {
		t.Errorf("AccountAddressFromKeypair() address length too short = %s (expected ~47 chars)", address)
	}
}

// TestAccountFromMnemonic tests direct SS58 address derivation from mnemonic.
// This scenario covers:
// - Valid mnemonic derives address successfully
// - Invalid mnemonic fails
// - Empty mnemonic fails
// - Same mnemonic produces same address
// - Address is in SS58 format
func TestAccountFromMnemonic(t *testing.T) {
	validMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	tests := []struct {
		name        string
		mnemonic    string
		expectError bool
		description string
	}{
		{
			name:        "valid_mnemonic",
			mnemonic:    validMnemonic,
			expectError: false,
			description: "deriving address from valid mnemonic",
		},
		{
			name:        "empty_mnemonic",
			mnemonic:    "",
			expectError: true,
			description: "deriving address from empty mnemonic",
		},
		{
			name:        "mnemonic_too_short",
			mnemonic:    "abandon abandon",
			expectError: true,
			description: "deriving address from mnemonic with too few words",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := AccountFromMnemonic(tt.mnemonic)

			if (err != nil) != tt.expectError {
				t.Errorf("AccountFromMnemonic() error = %v, expectError %v (%s)", err, tt.expectError, tt.description)
				return
			}

			if !tt.expectError {
				if address == "" {
					t.Errorf("AccountFromMnemonic() returned empty address (%s)", tt.description)
				}
			}
		})
	}
}

// TestAccountFromMnemonicConsistency tests that same mnemonic produces same address.
// This scenario covers:
// - Multiple derivations produce identical addresses
// - Address is deterministic and reproducible
func TestAccountFromMnemonicConsistency(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	address1, err := AccountFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("AccountFromMnemonic() error = %v", err)
	}

	address2, err := AccountFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("AccountFromMnemonic() error = %v", err)
	}

	if address1 != address2 {
		t.Errorf("AccountFromMnemonic() produces inconsistent results: %s vs %s", address1, address2)
	}
}
