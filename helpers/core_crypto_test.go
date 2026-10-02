package helpers

import (
	"testing"

	"github.com/radian-solusi/microservice-helpers/cryptoutil"
)

func TestHelpersDecrypt_AutoDetectsBothFormats(t *testing.T) {
	h := NewHelpers()
	key := "aK9$mL2@pQ7!vX5#bN8&wE4*jH6%yF1z"
	plaintext := "081234567890"

	native, err := h.Encrypt([]byte(plaintext), &key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	yii2, err := cryptoutil.EncryptYii2Legacy([]byte(plaintext), []byte(key), "")
	if err != nil {
		t.Fatalf("EncryptYii2Legacy: %v", err)
	}

	for name, encoded := range map[string]string{"go-native": native, "yii2": yii2} {
		got, err := h.Decrypt(encoded, &key)
		if err != nil {
			t.Fatalf("%s: Decrypt: %v", name, err)
		}
		if string(got) != plaintext {
			t.Fatalf("%s: plaintext = %q, want %q", name, got, plaintext)
		}
	}

	if _, err := h.Decrypt("not a ciphertext", &key); err == nil {
		t.Fatal("want error for unparsable input")
	}
}
