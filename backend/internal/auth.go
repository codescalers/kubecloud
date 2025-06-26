package internal

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

var saltLen = 5

func HashAndSaltPassword(password []byte) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return []byte{}, err
	}

	hashedPassword := sha256.Sum256(append(salt, password...))
	return append(salt, hashedPassword[:]...), nil
}

// VerifyPassword checks if given password is same as hashed one
func VerifyPassword(hashedPassword []byte, password string) bool {
	hashedPasswordCopy := make([]byte, len(hashedPassword))

	copy(hashedPasswordCopy, hashedPassword)
	salt := hashedPasswordCopy[:saltLen]

	checkedPass := sha256.Sum256(append(salt, []byte(password)...))
	return bytes.Equal(append(salt, checkedPass[:]...), hashedPassword)
}
