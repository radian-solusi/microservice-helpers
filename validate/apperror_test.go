package validate

import (
	"errors"
	"testing"
)

var errSentinel = errors.New("sentinel")

func TestNewFieldErrorWrapsSentinel(t *testing.T) {
	fe := NewFieldError("website", "url", errSentinel)
	if !errors.Is(fe, errSentinel) {
		t.Fatal("expected errors.Is to reach sentinel via Unwrap")
	}
	if len(fe.Errors) != 1 || fe.Errors[0].Field != "website" || fe.Errors[0].Tag != "url" {
		t.Fatalf("unexpected Errors: %+v", fe.Errors)
	}
	if fe.Error() == "" {
		t.Fatal("Error() must not be empty")
	}
}

func TestNewFieldErrorWithParam(t *testing.T) {
	fe := NewFieldError("minimum_order", "min", errSentinel, "1000")
	if fe.Errors[0].Param != "1000" {
		t.Fatalf("param not set: %+v", fe.Errors[0])
	}
}
