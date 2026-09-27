package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken generates a cryptographically secure random token of specified byte length
func GenerateRandomToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
