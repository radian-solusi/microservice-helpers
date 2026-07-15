package cryptoutil

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateSecureToken returns a URL-safe base64 token backed by byteLength
// random bytes from crypto/rand. A negative length returns an error.
func GenerateSecureToken(byteLength int) (string, error) {
	if byteLength < 0 {
		return "", fmt.Errorf("byte length must not be negative: %d", byteLength)
	}
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
