package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateRandomHex generates a cryptographically secure random hex string of given byte length.
func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CurrentTimestamp returns ISO8601/RFC3339 formatted string.
func CurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
