package strutil

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Slugify converts text to a lowercase hyphenated slug, preserving the existing
// behavior: only [a-z0-9] survive; spaces, hyphens, and underscores collapse to
// single hyphens; leading/trailing hyphens are trimmed.
func Slugify(text string) string {
	slug := strings.ToLower(text)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, slug)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

// ContainsSubstr reports whether substr is within s.
func ContainsSubstr(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Contains reports whether list contains str.
func Contains(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// Int64ToString formats i as a base-10 string.
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// ParseInt64 parses s as a base-10 int64, returning an error on invalid input.
// Unlike the legacy helper, it does not silently return 0.
func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseInt64Default parses s as int64, returning def on any error. Preserves the
// legacy "invalid becomes default" behavior for callers that need it.
func ParseInt64Default(s string, def int64) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}

// StringValue dereferences ptr, returning "" if nil.
func StringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// BoolString renders *ptr as "true"/"false", returning "" if nil.
func BoolString(ptr *bool) string {
	if ptr == nil {
		return ""
	}
	return strconv.FormatBool(*ptr)
}

// DefaultValue returns *ptr, or def when ptr is nil or points to "".
func DefaultValue(ptr *string, def string) string {
	if ptr == nil || *ptr == "" {
		return def
	}
	return *ptr
}

// FormatSize renders a byte count using binary units (B, KiB, MiB, GiB, ...).
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

const randomLabelCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomLabel returns prefix followed by length random alphanumeric characters,
// using crypto/rand. A negative length returns an error.
func RandomLabel(prefix string, length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("length must not be negative: %d", length)
	}
	b := make([]byte, length)
	max := big.NewInt(int64(len(randomLabelCharset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generate random label: %w", err)
		}
		b[i] = randomLabelCharset[n.Int64()]
	}
	return prefix + string(b), nil
}

// ExtractIPFromMetadata reads the "ip_address" field from a JSON metadata blob.
// Returns "" for nil, empty, malformed JSON, or a missing field.
func ExtractIPFromMetadata(metadata *string) string {
	if metadata == nil || *metadata == "" {
		return ""
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(*metadata), &m); err != nil {
		return ""
	}
	return m["ip_address"]
}
