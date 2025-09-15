package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidKeyLength = errors.New("invalid key length, must be 32 bytes for AES-256-GCM")
	ErrInvalidData      = errors.New("invalid encrypted data format")
)

type CryptoManager struct {
	config Configuration
}

func NewCryptoManager(config Configuration) *CryptoManager {
	return &CryptoManager{
		config: config,
	}
}

func (cm *CryptoManager) deriveKey(passphrase string, userIdentifier string) ([]byte, error) {
	if passphrase == "" {
		return nil, errors.New("passphrase cannot be empty")
	}

	if userIdentifier == "" {
		return nil, errors.New("user identifier cannot be empty")
	}

	if cm.config.MnemonicEncryptionSalt == "" {
		return nil, errors.New("mnemonic encryption salt is not configured - set MNEMONIC_ENCRYPTION_SALT environment variable")
	}

	// Create deterministic salt using master salt + user identifier
	// This separates salt from passphrase for better security
	saltBase := fmt.Sprintf("%s_user_%s", cm.config.MnemonicEncryptionSalt, userIdentifier)
	salt := sha256.Sum256([]byte(saltBase))

	key := argon2.IDKey(
		[]byte(passphrase),
		salt[:],
		cm.config.Argon2.Time,
		cm.config.Argon2.Memory,
		cm.config.Argon2.Threads,
		32,
	)

	return key, nil
}

func (cm *CryptoManager) encrypt(plainText string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (cm *CryptoManager) decrypt(encryptedText string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKeyLength
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidData
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func (cm *CryptoManager) EncryptMnemonic(plainText string, userAddress string) (string, error) {
	if cm.config.MnemonicEncryptionPassphrase == "" {
		return "", errors.New("mnemonic encryption passphrase is not configured - set MNEMONIC_ENCRYPTION_PASSPHRASE environment variable")
	}

	key, err := cm.deriveKey(cm.config.MnemonicEncryptionPassphrase, userAddress)
	if err != nil {
		return "", err
	}
	return cm.encrypt(plainText, key)
}

func (cm *CryptoManager) DecryptMnemonic(encryptedText string, userAddress string) (string, error) {
	if cm.config.MnemonicEncryptionPassphrase == "" {
		return "", errors.New("mnemonic encryption passphrase is not configured - set MNEMONIC_ENCRYPTION_PASSPHRASE environment variable")
	}

	key, err := cm.deriveKey(cm.config.MnemonicEncryptionPassphrase, userAddress)
	if err != nil {
		return "", err
	}
	return cm.decrypt(encryptedText, key)
}
