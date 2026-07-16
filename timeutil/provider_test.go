package timeutil

import (
	"testing"
	"time"
)

func TestDefaultTimeProvider(t *testing.T) {
	var _ TimeProvider = (*DefaultTimeProvider)(nil)
	p := &DefaultTimeProvider{}
	if p.IntToDuration(3) != 3*time.Second {
		t.Fatal("duration")
	}
	if p.DurationToInt(2500*time.Millisecond) != 2 {
		t.Fatal("seconds")
	}
	if p.IntToTime(10).Unix() != 10 {
		t.Fatal("unix")
	}
	if got := p.DateTimeToString(time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)); got != "2026-07-16 01:02:03" {
		t.Fatalf("got %q", got)
	}
	if got := p.StringToTime("bad", time.RFC3339); !got.IsZero() {
		t.Fatalf("got %v", got)
	}
}
