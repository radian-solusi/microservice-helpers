package phone

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

var allowedDigits = []string{"2", "3", "4", "5", "6", "7", "8"}

// Validate normalizes an Indonesian phone number to 62-prefixed form and
// validates it. Canonical behavior from the authorization service: strips
// spaces; requires 9..15 chars, digits only; converts a leading "0" to "62";
// prepends "62" when absent; and requires the national leading digit to be 2..8.
func Validate(numberString string) (normalized string, err error) {
	formatted := strings.ReplaceAll(numberString, " ", "")

	if formatted == "" {
		return "", errors.New("phone number is required")
	}
	if len(formatted) < 9 || len(formatted) > 15 {
		return "", errors.New("phone number must be between 9 and 15 digits")
	}
	for _, c := range formatted {
		if !unicode.IsDigit(c) {
			return "", errors.New("phone number cannot contains character")
		}
	}

	if strings.HasPrefix(formatted, "0") {
		formatted = "62" + formatted[1:]
	}
	if formatted[0:2] != "62" {
		formatted = "62" + formatted
	}

	// Guard: after normalization the string always has "62" + at least one
	// digit, so index 2 is safe (min length is 9).
	thirdDigit := string(formatted[2])
	if !slices.Contains(allowedDigits, thirdDigit) {
		var b strings.Builder
		for _, n := range allowedDigits {
			b.WriteString(n)
		}
		return "", fmt.Errorf("Phone number must start with %s", b.String())
	}

	return formatted, nil
}

// countryCodes is the ordered normalization list; longer codes must be checked
// before their prefixes (e.g. "852" before "85").
var countryCodes = []string{
	"+62", "62",
	"+61", "61",
	"+1", "1",
	"+44", "44",
	"+81", "81",
	"+86", "86",
	"+91", "91",
	"+33", "33",
	"+49", "49",
	"+39", "39",
	"+34", "34",
	"+7", "7",
	"+82", "82",
	"+65", "65",
	"+60", "60",
	"+66", "66",
	"+84", "84",
	"+63", "63",
	"+852", "852",
	"+886", "886",
}

// Normalize strips spaces and rewrites a known country code prefix to a leading
// "0". Unknown prefixes and empty input pass through unchanged (empty returns
// ""). Behavior matches the investor service helper.
func Normalize(phonestring string) string {
	phone := strings.ReplaceAll(strings.TrimSpace(phonestring), " ", "")
	if phone == "" {
		return ""
	}
	for _, code := range countryCodes {
		if strings.HasPrefix(phone, code) {
			return "0" + phone[len(code):]
		}
	}
	return phone
}
