package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateRandomToken generates a cryptographically secure random token of specified byte length
func GenerateRandomToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GeneratePostID returns a prefixed unique ID for a post
func GeneratePostID() string {
	token, err := GenerateRandomToken(8)
	if err != nil {
		return fmt.Sprintf("post_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("post_%s", token)
}

func GenerateChatID() string {
	token, err := GenerateRandomToken(8)
	if err != nil {
		return fmt.Sprintf("chat_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("chat_%s", token)
}

func GenerateMessageID() string {
	token, err := GenerateRandomToken(8)
	if err != nil {
		return fmt.Sprintf("message_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("message_%s", token)
}