package cryptoutil

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// Checksum returns a hex MD5 over the JSON encoding of data. This is a
// compatibility checksum for change detection, not a security primitive; MD5 is
// not collision resistant.
func Checksum(data any) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal data to JSON: %w", err)
	}
	hash := md5.Sum(jsonBytes)
	return fmt.Sprintf("%x", hash), nil
}
