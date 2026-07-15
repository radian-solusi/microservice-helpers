package otp

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

var testKey = []byte("01234567890123456789012345678901") // 32 bytes

func TestGenerateTOTPSecret(t *testing.T) {
	if _, _, err := GenerateTOTPSecret("", "acct"); err == nil {
		t.Error("expected error for empty issuer")
	}
	if _, _, err := GenerateTOTPSecret("E-IPO", ""); err == nil {
		t.Error("expected error for empty account")
	}
	secret, url, err := GenerateTOTPSecret("E-IPO", "user@example.com")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if secret == "" || !strings.HasPrefix(url, "otpauth://") {
		t.Errorf("bad output: secret=%q url=%q", secret, url)
	}
	if !strings.Contains(url, "E-IPO") {
		t.Errorf("url missing issuer: %q", url)
	}
}

func TestVerifyTOTPCode(t *testing.T) {
	if VerifyTOTPCode("", "123") || VerifyTOTPCode("SECRET", "") {
		t.Error("empty inputs should be false")
	}
	secret, _, _ := GenerateTOTPSecret("E-IPO", "user")
	code, _ := totp.GenerateCode(secret, time.Now())
	if !VerifyTOTPCode(secret, code) {
		t.Error("expected valid code to verify")
	}
}

func TestGenerateNumericOTP(t *testing.T) {
	if _, err := GenerateNumericOTP(0); err == nil {
		t.Error("expected error for length 0")
	}
	if _, err := GenerateNumericOTP(-1); err == nil {
		t.Error("expected error for negative length")
	}
	otp, err := GenerateNumericOTP(6)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(otp) != 6 {
		t.Errorf("length = %d, want 6", len(otp))
	}
	for _, r := range otp {
		if r < '0' || r > '9' {
			t.Errorf("non-digit %q in %q", r, otp)
		}
	}
}

func TestPreAuthTokenRoundTrip(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	token, err := GeneratePreAuthToken("user-123", testKey, now, 5*time.Minute)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	uid, err := VerifyPreAuthToken(token, testKey, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if uid != "user-123" {
		t.Errorf("uid = %q, want user-123", uid)
	}
}

func TestPreAuthTokenExpiry(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	token, _ := GeneratePreAuthToken("u", testKey, now, 1*time.Minute)
	// exactly at expiry boundary: now.Unix() > exp is false -> still valid
	atExp := now.Add(1 * time.Minute)
	if _, err := VerifyPreAuthToken(token, testKey, atExp); err != nil {
		t.Errorf("token at exact expiry should be valid: %v", err)
	}
	// one second past expiry -> invalid
	past := now.Add(1*time.Minute + 1*time.Second)
	if _, err := VerifyPreAuthToken(token, testKey, past); err == nil {
		t.Error("expected expired error")
	}
}

func TestPreAuthTokenInvalidInputs(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if _, err := GeneratePreAuthToken("", testKey, now, time.Minute); err == nil {
		t.Error("expected error for empty user")
	}
	if _, err := GeneratePreAuthToken("u", []byte("short"), now, time.Minute); err == nil {
		t.Error("expected error for bad key")
	}
	if _, err := GeneratePreAuthToken("u", testKey, now, 0); err == nil {
		t.Error("expected error for non-positive ttl")
	}
	if _, err := VerifyPreAuthToken("", testKey, now); err == nil {
		t.Error("expected error for empty token")
	}

	token, _ := GeneratePreAuthToken("u", testKey, now, time.Minute)
	wrongKey := []byte("abcdefghijklmnopqrstuvwxyz012345")
	if _, err := VerifyPreAuthToken(token, wrongKey, now); err == nil {
		t.Error("expected error for wrong key")
	}
	tampered := token[:len(token)-2] + "AA"
	if _, err := VerifyPreAuthToken(tampered, testKey, now); err == nil {
		t.Error("expected error for tampered token")
	}
}
