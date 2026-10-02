package cryptoutil

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

// Vectors produced by Yii2 2.0.54 Security::encryptByKey with cipher
// AES-256-CBC (e-IPO common/config/main.php). Do not regenerate.
var yii2Vectors = []struct {
	name          string
	plaintextB64  string
	key           string
	ciphertextB64 string
}{
	{
		name:          "phone",
		plaintextB64:  "MDgxMjM0NTY3ODkw",
		key:           "aK9$mL2@pQ7!vX5#bN8&wE4*jH6%yF1z",
		ciphertextB64: "CvozxvA9TL/vC9Cjm94mqOgne2bCWNXuCYS8QRbDEEwxYzg5MzQzM2M0NmJjMTI0YjBmZDkyZWY0NmU3OTk2OTZkZDgzZmJmY2Q4ZjU3NWUzY2YxNDI3NzUwZDBlMjk13z+goY/sMf5VxrsgzX7oBUPxGqr9f6WgwUZ03IJ8J+U=",
	},
	{
		name:          "e164 phone other key",
		plaintextB64:  "KzYyODEyMzQ1Njc4OTA=",
		key:           "BcdEFghijklmnopqrstuvwxyz0987654",
		ciphertextB64: "yyKdA8p3VOmG8V3ewCRESMpva4gVTGNen3Fl2orvF1diNzg2ZTZkNTNlZTYyMDJmYmRkMzZkOTM5YTcwZTJmNmMyMzA1MmRhNGQ2NTZjOTk5M2Q1ZWE3MjRjNGRhOTU3UaVi9Z9QmDrVDLtO1pHKIEBSHWbDZ6c7Mzg4lNnvq1M=",
	},
	{
		name:          "empty plaintext",
		plaintextB64:  "",
		key:           "aK9$mL2@pQ7!vX5#bN8&wE4*jH6%yF1z",
		ciphertextB64: "i9hnMbp1GKpvXo7Z+FyNvbXULzJkB4CuClTrlHrv3GYyMzI1YzlmZTU0YTgyN2ZlZmNhM2M5MmEzZGYyZGJhNzgzM2JkYzBkNWYxMjIwNTQxOGZiZDY4YjFhMzU2NzhkdrCLJyUALofXixw2OO924zu/whn5b4nGTRTxPrdimIE=",
	},
}

func TestDecryptYii2Legacy_PHPVectors(t *testing.T) {
	for _, v := range yii2Vectors {
		t.Run(v.name, func(t *testing.T) {
			want, err := base64.StdEncoding.DecodeString(v.plaintextB64)
			if err != nil {
				t.Fatalf("bad fixture: %v", err)
			}
			got, err := DecryptYii2Legacy(v.ciphertextB64, []byte(v.key), Yii2CipherAES256CBC)
			if err != nil {
				t.Fatalf("DecryptYii2Legacy: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("plaintext = %q, want %q", got, want)
			}
		})
	}
}

func TestDecryptYii2Legacy_WrongKey(t *testing.T) {
	v := yii2Vectors[0]
	_, err := DecryptYii2Legacy(v.ciphertextB64, []byte("0000000000000000000000000000000x"), "")
	if !errors.Is(err, ErrMACMismatch) {
		t.Fatalf("err = %v, want ErrMACMismatch", err)
	}
}

func TestDecryptYii2Legacy_BadFormat(t *testing.T) {
	for _, in := range []string{"not base64!!", "", base64.StdEncoding.EncodeToString([]byte("short"))} {
		if _, err := DecryptYii2Legacy(in, []byte("key"), ""); !errors.Is(err, ErrInvalidYii2Format) {
			t.Fatalf("input %q: err = %v, want ErrInvalidYii2Format", in, err)
		}
	}
}

func TestYii2Legacy_RoundTrip(t *testing.T) {
	key := []byte("aK9$mL2@pQ7!vX5#bN8&wE4*jH6%yF1z")
	for _, cipherName := range []string{"", Yii2CipherAES128CBC, Yii2CipherAES192CBC, Yii2CipherAES256CBC} {
		for _, pt := range [][]byte{nil, []byte("081234567890"), bytes.Repeat([]byte("a"), 16), bytes.Repeat([]byte("b"), 100)} {
			enc, err := EncryptYii2Legacy(pt, key, cipherName)
			if err != nil {
				t.Fatalf("encrypt(%q): %v", cipherName, err)
			}
			got, err := DecryptYii2Legacy(enc, key, cipherName)
			if err != nil {
				t.Fatalf("decrypt(%q): %v", cipherName, err)
			}
			if !bytes.Equal(got, pt) && !(len(got) == 0 && len(pt) == 0) {
				t.Fatalf("round-trip = %q, want %q", got, pt)
			}
		}
	}
}

func TestEncryptYii2Legacy_UnsupportedCipher(t *testing.T) {
	if _, err := EncryptYii2Legacy(nil, []byte("k"), "AES-256-GCM"); err == nil {
		t.Fatal("want error for unsupported cipher")
	}
}

// Yii2 payload layout is keySalt||hex MAC||IV||ciphertext; the legacy Go
// format must not accidentally parse it, otherwise auto-detect would return
// garbage instead of falling through.
func TestLegacyCBC_RejectsYii2Payload(t *testing.T) {
	v := yii2Vectors[0]
	if _, err := DecryptLegacyCBC(v.ciphertextB64, []byte(v.key)); err == nil {
		t.Fatal("DecryptLegacyCBC accepted a Yii2 payload")
	}
}
