package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// Wire format of Yii2 Security::encryptByKey:
//
//	base64( keySalt[keySize] || hex(HMAC-SHA256)[64] || IV[16] || ciphertext )
//
// Key derivation (kdfHash/macHash = sha256):
//
//	key     = HKDF(secret, salt=keySalt, info="",                 len=keySize)
//	authKey = HKDF(key,    salt=nil,     info="AuthorizationKey", len=keySize)
//	MAC     = hex(HMAC-SHA256(authKey, IV||ciphertext))
const (
	Yii2CipherAES128CBC = "AES-128-CBC"
	Yii2CipherAES192CBC = "AES-192-CBC"
	Yii2CipherAES256CBC = "AES-256-CBC"

	yii2AuthKeyInfo = "AuthorizationKey"
	yii2MACHexLen   = sha256.Size * 2
)

// ErrMACMismatch is returned when the Yii2 MAC does not authenticate the
// ciphertext: either the data was tampered with or the key is wrong.
var ErrMACMismatch = errors.New("cryptoutil: MAC mismatch")

// ErrInvalidYii2Format is returned when the input is not a structurally valid
// Yii2 Security::encryptByKey payload.
var ErrInvalidYii2Format = errors.New("cryptoutil: invalid Yii2 ciphertext format")

// yii2KeySize maps a Yii2 $cipher name to its AES key size. An empty name
// defaults to AES-256-CBC, which is what e-IPO configures.
func yii2KeySize(cipherName string) (int, error) {
	switch cipherName {
	case "", Yii2CipherAES256CBC:
		return 32, nil
	case Yii2CipherAES192CBC:
		return 24, nil
	case Yii2CipherAES128CBC:
		return 16, nil
	default:
		return 0, fmt.Errorf("cryptoutil: unsupported Yii2 cipher %q", cipherName)
	}
}

func yii2DeriveKeys(secret, keySalt []byte, keySize int) (key, authKey []byte, err error) {
	key, err = hkdf.Key(sha256.New, secret, keySalt, "", keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("derive key: %w", err)
	}
	authKey, err = hkdf.Key(sha256.New, key, nil, yii2AuthKeyInfo, keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("derive auth key: %w", err)
	}
	return key, authKey, nil
}

func yii2MAC(authKey, data []byte) string {
	m := hmac.New(sha256.New, authKey)
	m.Write(data)
	return hex.EncodeToString(m.Sum(nil))
}

// EncryptYii2Legacy encrypts plaintext in the Yii2 Security::encryptByKey wire
// format so PHP (e-IPO) can read it back. cipherName may be empty for the
// e-IPO default AES-256-CBC.
func EncryptYii2Legacy(plaintext, secret []byte, cipherName string) (string, error) {
	keySize, err := yii2KeySize(cipherName)
	if err != nil {
		return "", err
	}

	keySalt := make([]byte, keySize)
	if _, err := rand.Read(keySalt); err != nil {
		return "", fmt.Errorf("generate key salt: %w", err)
	}
	key, authKey, err := yii2DeriveKeys(secret, keySalt, keySize)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("generate IV: %w", err)
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	out := make([]byte, 0, len(keySalt)+yii2MACHexLen+len(iv)+len(ciphertext))
	out = append(out, keySalt...)
	out = append(out, yii2MAC(authKey, append(append([]byte(nil), iv...), ciphertext...))...)
	out = append(out, iv...)
	out = append(out, ciphertext...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptYii2Legacy reverses EncryptYii2Legacy and reads data produced by Yii2
// Security::encryptByKey. The MAC is verified before any plaintext is
// returned; a wrong key or tampered payload yields ErrMACMismatch.
func DecryptYii2Legacy(encoded string, secret []byte, cipherName string) ([]byte, error) {
	keySize, err := yii2KeySize(cipherName)
	if err != nil {
		return nil, err
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: decode base64: %v", ErrInvalidYii2Format, err)
	}
	minLen := keySize + yii2MACHexLen + aes.BlockSize + aes.BlockSize
	if len(raw) < minLen || (len(raw)-minLen)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("%w: unexpected length %d", ErrInvalidYii2Format, len(raw))
	}

	keySalt := raw[:keySize]
	mac := string(raw[keySize : keySize+yii2MACHexLen])
	authenticated := raw[keySize+yii2MACHexLen:]
	iv := authenticated[:aes.BlockSize]
	ciphertext := authenticated[aes.BlockSize:]

	key, authKey, err := yii2DeriveKeys(secret, keySalt, keySize)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare([]byte(mac), []byte(yii2MAC(authKey, authenticated))) != 1 {
		return nil, ErrMACMismatch
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	return unpadded, nil
}

func pkcs7Pad(src []byte, blockSize int) []byte {
	n := blockSize - len(src)%blockSize
	out := make([]byte, len(src)+n)
	copy(out, src)
	for i := len(src); i < len(out); i++ {
		out[i] = byte(n)
	}
	return out
}

func pkcs7Unpad(src []byte, blockSize int) ([]byte, error) {
	if len(src) == 0 || len(src)%blockSize != 0 {
		return nil, errors.New("cryptoutil: invalid padded length")
	}
	n := int(src[len(src)-1])
	if n == 0 || n > blockSize || n > len(src) {
		return nil, errors.New("cryptoutil: invalid PKCS#7 padding")
	}
	for _, b := range src[len(src)-n:] {
		if int(b) != n {
			return nil, errors.New("cryptoutil: invalid PKCS#7 padding")
		}
	}
	return src[:len(src)-n], nil
}
