package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// Generate a random string
func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length/2)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes)[:length], nil
}
