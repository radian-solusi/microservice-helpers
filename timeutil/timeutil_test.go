package timeutil

import (
	"testing"
	"time"
)

func TestLoadLocation(t *testing.T) {
	if _, err := LoadLocation("Asia/Jakarta"); err != nil {
		t.Errorf("valid location: %v", err)
	}
	if _, err := LoadLocation("Invalid/Nowhere"); err == nil {
		t.Error("expected error for invalid location")
	}
}

func TestParse(t *testing.T) {
	got, err := Parse(FormatDateTime, "2024-01-15 10:30:00")
	if err != nil || got.Year() != 2024 {
		t.Errorf("parse valid: %v %v", got, err)
	}
	if _, err := Parse(FormatDateTime, "bad"); err == nil {
		t.Error("expected parse error")
	}
}

func TestFormatConstants(t *testing.T) {
	tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if got := Format(tm, FormatDate); got != "2024-01-15" {
		t.Errorf("FormatDate: got %q", got)
	}
	if got := Format(tm, FormatDateTime); got != "2024-01-15 10:30:00" {
		t.Errorf("FormatDateTime: got %q", got)
	}
	if got := Format(tm, FormatDateTimeFile); got != "20240115103000" {
		t.Errorf("FormatDateTimeFile: got %q", got)
	}
}

func TestUnixToTime(t *testing.T) {
	got := UnixToTime(1705314600)
	if got.Year() != 2024 {
		t.Errorf("unexpected year %d", got.Year())
	}
}

func TestSecondsToDurationRoundTrip(t *testing.T) {
	d := SecondsToDuration(300)
	if d != 5*time.Minute {
		t.Errorf("300s = %v, want 5m0s", d)
	}
	if s := DurationToSeconds(d); s != 300 {
		t.Errorf("DurationToSeconds = %d, want 300", s)
	}
}

func TestUnixToString(t *testing.T) {
	got := UnixToString(1705314600, time.UTC, FormatDateTime)
	if got != "2024-01-15 10:30:00" {
		t.Errorf("got %q", got)
	}
	gotNil := UnixToString(1705314600, nil, FormatDateTime)
	if gotNil != "2024-01-15 10:30:00" {
		t.Errorf("nil location: got %q", gotNil)
	}
}
