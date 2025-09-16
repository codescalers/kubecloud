package internal

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"kubecloud/internal/logger"
	"kubecloud/models"

	"sync"

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
		return nil, errors.New("mnemonic encryption passphrase is not configured in config")
	}

	if userIdentifier == "" {
		return nil, errors.New("user identifier cannot be empty")
	}

	if cm.config.MnemonicEncryptionSalt == "" {
		return nil, errors.New("mnemonic encryption salt is not configured in config")
	}

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

func (cm *CryptoManager) encrypt(plainText string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)

	return ciphertext, nil
}

func (cm *CryptoManager) decrypt(encryptedBytes []byte, key []byte) (string, error) {
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

	nonceSize := aesGCM.NonceSize()
	if len(encryptedBytes) < nonceSize {
		return "", ErrInvalidData
	}

	nonce, ciphertext := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

func (cm *CryptoManager) getMnemonicKey(userAddress string) ([]byte, error) {
	if cm.config.MnemonicEncryptionPassphrase == "" {
		return nil, errors.New("mnemonic encryption passphrase is not configured in config")
	}

	return cm.deriveKey(cm.config.MnemonicEncryptionPassphrase, userAddress)
}

func (cm *CryptoManager) EncryptMnemonic(plainText string, userAddress string) ([]byte, error) {
	key, err := cm.getMnemonicKey(userAddress)
	if err != nil {
		return nil, err
	}
	return cm.encrypt(plainText, key)
}

func (cm *CryptoManager) DecryptMnemonic(encryptedBytes []byte, userAddress string) (string, error) {
	key, err := cm.getMnemonicKey(userAddress)
	if err != nil {
		return "", err
	}
	return cm.decrypt(encryptedBytes, key)
}

func (cm *CryptoManager) EnsureMnemonicsEncrypted(ctx context.Context, db models.DB) error {
	users, err := db.ListAllUsers()
	if err != nil {
		return fmt.Errorf("ensure encryption: list users failed: %w", err)
	}

	const maxWorkers = 16
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i := range users {
		u := users[i]
		wg.Add(1)
		sem <- struct{}{}
		go func(u models.User) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return
			default:
			}

			if len(u.Mnemonic) == 0 {
				return
			}

			// Derive account address if missing and mnemonic appears to be plaintext
			if len(u.AccountAddress) == 0 {
				addr, err := AccountFromMnemonic(string(u.Mnemonic))
				if err != nil {
					logger.GetLogger().Error().Err(err).Int("user_id", u.ID).Msg("failed to derive account address from mnemonic")
					return
				}
				u.AccountAddress = addr
			}

			if _, err := cm.DecryptMnemonic(u.Mnemonic, u.AccountAddress); err == nil {
				return
			}

			encryptedMnemonic, err := cm.EncryptMnemonic(string(u.Mnemonic), u.AccountAddress)
			if err != nil {
				return
			}
			u.Mnemonic = encryptedMnemonic
			_ = db.UpdateUserByID(&u)
		}(u)
	}

	wg.Wait()
	return nil
}
