package timeutil

import (
	"os"
	"time"
)

// TimeProvider mirrors investor's helper time contract, allowing callers to
// inject time behavior instead of calling time.Now/time.Parse directly.
type TimeProvider interface {
	Now() time.Time
	IntDateToString(int) string
	FormattedDate(string, string) string
	IntToTime(int) time.Time
	IntToDuration(int) time.Duration
	StringToTime(string, string) time.Time
	DurationToInt(time.Duration) int
}

// DefaultTimeProvider is the production TimeProvider implementation.
type DefaultTimeProvider struct{}

func (*DefaultTimeProvider) Now() time.Time { return time.Now() }

// IntDateToString formats a Unix timestamp as RFC3339 in the TZ environment
// location, defaulting to UTC when TZ is unset or invalid.
func (*DefaultTimeProvider) IntDateToString(v int) string {
	name := os.Getenv("TZ")
	loc, err := time.LoadLocation(name)
	if err != nil {
		loc = time.UTC
	}
	return time.Unix(int64(v), 0).In(loc).Format(time.RFC3339)
}

// FormattedDate reparses an RFC3339 string v and renders it with layout,
// returning "" on parse failure.
func (*DefaultTimeProvider) FormattedDate(v, layout string) string {
	parsed, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return ""
	}
	return parsed.Format(layout)
}

func (*DefaultTimeProvider) IntToTime(v int) time.Time { return time.Unix(int64(v), 0) }

func (*DefaultTimeProvider) IntToDuration(v int) time.Duration { return time.Duration(v) * time.Second }

// StringToTime parses v with layout, returning the zero time on failure.
func (*DefaultTimeProvider) StringToTime(v, layout string) time.Time {
	parsed, _ := time.Parse(layout, v)
	return parsed
}

func (*DefaultTimeProvider) DurationToInt(v time.Duration) int { return int(v.Seconds()) }

func (*DefaultTimeProvider) DateTimeToInt(v time.Time) int { return int(v.Unix()) }

func (*DefaultTimeProvider) DateTimeToString(v time.Time) string { return v.Format(FormatDateTime) }

// StringToDateTime parses v using FormatDateTime, returning the zero time on
// failure.
func (*DefaultTimeProvider) StringToDateTime(v string) time.Time {
	parsed, _ := time.Parse(FormatDateTime, v)
	return parsed
}
