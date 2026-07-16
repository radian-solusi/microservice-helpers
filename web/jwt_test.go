package web

import (
	"testing"
	"time"
)

func TestJWTRoundTrip(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	j, err := NewJWT(key)
	if err != nil {
		t.Fatal(err)
	}
	type claims struct {
		UserID string `json:"user_id"`
	}
	token, err := j.Generate(claims{UserID: "u1"}, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	var got claims
	if err := j.Parse(token, &got); err != nil {
		t.Fatal(err)
	}
	if got.UserID != "u1" {
		t.Fatalf("got %q", got.UserID)
	}
}

func TestJWTRejectsExpired(t *testing.T) {
	key := []byte("12345678901234567890123456789012")
	j, _ := NewJWT(key)
	token, _ := j.Generate(map[string]string{"user_id": "u1"}, time.Now().Add(-time.Hour))
	var sink map[string]any
	if err := j.Parse(token, &sink); err == nil {
		t.Fatal("expected expiry error")
	}
}

func TestNewJWTRejectsShortKey(t *testing.T) {
	if _, err := NewJWT([]byte("short")); err == nil {
		t.Fatal("expected error")
	}
}
