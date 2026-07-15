package cryptoutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// EncryptLegacyCBC encrypts plaintext with AES-256-CBC using zero padding and a
// crypto/rand IV, producing the legacy wire format
// base64(IV[16] || base64(ciphertext)).
//
// This exists only for compatibility with existing ciphertext. CBC here is
// unauthenticated and provides no tamper detection; new encrypted data should
// use an authenticated construction.
func EncryptLegacyCBC(plaintext, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes for AES-256")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generate IV: %w", err)
	}

	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := append(append([]byte(nil), plaintext...), make([]byte, padding)...)

	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	inner := base64.StdEncoding.EncodeToString(ciphertext)
	outer := append(append([]byte(nil), iv...), []byte(inner)...)
	return base64.StdEncoding.EncodeToString(outer), nil
}

// DecryptLegacyCBC reverses EncryptLegacyCBC. Structurally malformed input
// returns an error and never panics. Because CBC is unauthenticated, tampered
// ciphertext that remains structurally valid decrypts to garbage rather than an
// error.
func DecryptLegacyCBC(encoded string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	outer, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode outer base64: %w", err)
	}
	if len(outer) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := outer[:aes.BlockSize]
	inner, err := base64.StdEncoding.DecodeString(string(outer[aes.BlockSize:]))
	if err != nil {
		return nil, fmt.Errorf("decode inner base64: %w", err)
	}
	if len(inner) == 0 || len(inner)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(inner))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, inner)
	return bytes.TrimRight(plaintext, "\x00"), nil
}
