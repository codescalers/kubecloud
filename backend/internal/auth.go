package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateSalt generates a random salt for password hashing
func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// HashPassword hashes a password with a given salt
func HashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// VerifyPassword checks that hashed password is valid
func VerifyPassword(hashedPassword, plainPassword, salt string) bool {
	generated := HashPassword(plainPassword, salt)
	return generated == hashedPassword

}
