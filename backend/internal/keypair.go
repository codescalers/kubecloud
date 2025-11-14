package internal

import (
	"errors"
	"strings"

	"github.com/vedhavyas/go-subkey"
	"github.com/vedhavyas/go-subkey/sr25519"
)

// Blockchain-related constants
const (
	// SS58AddressFormat is the format used for Substrate-based chain addresses
	SS58AddressFormat = 42
)

var (
	errMnemonicEmpty    = errors.New("mnemonic cannot be empty")
	errMnemonicTooShort = errors.New("mnemonic must be at least 12 words")
)

// validateMnemonic checks if the mnemonic is non-empty and at least 12 words
func validateMnemonic(mnemonic string) error {
	mnemonic = strings.TrimSpace(mnemonic)
	if mnemonic == "" {
		return errMnemonicEmpty
	}
	words := strings.Fields(mnemonic)
	if len(words) < 12 {
		return errMnemonicTooShort
	}
	return nil
}

// KeyPairFromMnemonic derives a keypair from a mnemonic
func KeyPairFromMnemonic(mnemonic string) (subkey.KeyPair, error) {
	if err := validateMnemonic(mnemonic); err != nil {
		return nil, err
	}
	keyPair, err := sr25519.Scheme{}.FromPhrase(strings.TrimSpace(mnemonic), "")
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}

// AccountAddressFromKeypair returns the SS58 address for a given keypair
func AccountAddressFromKeypair(keyPair subkey.KeyPair) (string, error) {
	return keyPair.SS58Address(SS58AddressFormat)
}

// AccountFromMnemonic returns the SS58 address from a mnemonic
func AccountFromMnemonic(mnemonic string) (string, error) {
	keyPair, err := KeyPairFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}
	return AccountAddressFromKeypair(keyPair)
}
