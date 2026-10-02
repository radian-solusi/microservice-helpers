package helpers

import (
	"errors"
	"testing"

	"github.com/radian-solusi/microservice-helpers/cryptoutil"
)

const (
	testAppKey    = "12345678901234567890123456789012"
	testLegacyKey = "aK9$mL2@pQ7!vX5#bN8&wE4*jH6%yF1z"
)

func newHelpersWithLegacyKey(t *testing.T, legacyKey string) *Helpers {
	t.Helper()
	extra := ""
	if legacyKey != "" {
		extra = "legacy_app_key=\"" + legacyKey + "\"\n"
	}
	root := writeConfig(t, extra)
	return NewHelpers(WithConfigStart(root), WithEnvLookup(fileConfigLookup))
}

// No legacy_app_key: EncryptLegacy must keep the old Go-native behaviour under AppKey.
func TestEncryptLegacy_WithoutLegacyKeyKeepsNativeFormat(t *testing.T) {
	h := newHelpersWithLegacyKey(t, "")
	plaintext := "081234567890"

	encoded, err := h.EncryptLegacy([]byte(plaintext))
	if err != nil {
		t.Fatalf("EncryptLegacy: %v", err)
	}
	if _, err := cryptoutil.DecryptLegacyCBC(encoded, []byte(testAppKey)); err != nil {
		t.Fatalf("want go-native format under AppKey: %v", err)
	}
	got, err := h.DecryptLegacy(encoded)
	if err != nil || string(got) != plaintext {
		t.Fatalf("DecryptLegacy = %q, %v", got, err)
	}
}

// Ciphertext produced by e-IPO (Yii2 format) under the legacy key.
func TestDecryptLegacy_ReadsYii2WithLegacyKey(t *testing.T) {
	h := newHelpersWithLegacyKey(t, testLegacyKey)
	plaintext := "081234567890"

	encoded, err := cryptoutil.EncryptYii2Legacy([]byte(plaintext), []byte(testLegacyKey), "")
	if err != nil {
		t.Fatalf("EncryptYii2Legacy: %v", err)
	}
	got, err := h.DecryptLegacy(encoded)
	if err != nil {
		t.Fatalf("DecryptLegacy: %v", err)
	}
	if string(got) != plaintext {
		t.Fatalf("plaintext = %q, want %q", got, plaintext)
	}

	// AppKey path must not read legacy data.
	if _, err := h.Decrypt(encoded, nil); err == nil {
		t.Fatal("want error decrypting legacy data with AppKey")
	}
}

func TestDecryptLegacy_WrongKeyReturnsMACMismatch(t *testing.T) {
	h := newHelpersWithLegacyKey(t, "wrong-key-wrong-key-wrong-key-32")

	encoded, err := cryptoutil.EncryptYii2Legacy([]byte("081234567890"), []byte(testLegacyKey), "")
	if err != nil {
		t.Fatalf("EncryptYii2Legacy: %v", err)
	}
	_, err = h.DecryptLegacy(encoded)
	if !errors.Is(err, cryptoutil.ErrMACMismatch) {
		t.Fatalf("err = %v, want ErrMACMismatch", err)
	}
}

func TestEncryptLegacy_RoundTripYii2(t *testing.T) {
	h := newHelpersWithLegacyKey(t, testLegacyKey)
	plaintext := "081234567890"

	encoded, err := h.EncryptLegacy([]byte(plaintext))
	if err != nil {
		t.Fatalf("EncryptLegacy: %v", err)
	}
	// e-IPO must be able to read it: it is the Yii2 wire format under the legacy key.
	if _, err := cryptoutil.DecryptYii2Legacy(encoded, []byte(testLegacyKey), ""); err != nil {
		t.Fatalf("want Yii2 format under legacy key: %v", err)
	}
	got, err := h.DecryptLegacy(encoded)
	if err != nil || string(got) != plaintext {
		t.Fatalf("DecryptLegacy = %q, %v", got, err)
	}
}
