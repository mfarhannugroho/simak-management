package document

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateQRToken generates a cryptographically random, tamper-proof token
func GenerateQRToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback
		return fmt.Sprintf("smk-%d", hex.EncodeToString(bytes))
	}
	return fmt.Sprintf("smk-%s", hex.EncodeToString(bytes))
}

// GenerateQRCodeBase64 generates a base64 encoded PNG data URI for the given string content
func GenerateQRCodeBase64(content string) (string, error) {
	pngBytes, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}
