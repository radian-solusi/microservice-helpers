package cryptoutil

import (
	"bytes"
	"testing"
)

func TestPadding(t *testing.T) {
	if got := Unpad([]byte{'a', 'b', 2, 2}); !bytes.Equal(got, []byte("ab")) {
		t.Fatalf("got %v", got)
	}
	if got := Unpad([]byte{'a', 9}); got != nil {
		t.Fatalf("got %v", got)
	}
	if got := ZeroUnpad([]byte{'a', 'b', 0, 0}); !bytes.Equal(got, []byte("ab")) {
		t.Fatalf("got %v", got)
	}
}
