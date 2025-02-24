package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func ShortID(length int) (string, error) {
	// Generate random bytes
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base64 (URL-safe) and trim padding
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}
