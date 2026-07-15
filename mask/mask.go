package mask

import "strings"

const maskPlaceholder = "***"

// MaskEmail masks an email address for logging, preserving only the first
// character of the local part and the full domain.
func MaskEmail(s string) string {
	if s == "" {
		return ""
	}
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return maskPlaceholder
	}
	local, domain := s[:at], s[at+1:]
	if !strings.Contains(domain, ".") {
		return maskPlaceholder
	}
	return local[:1] + maskPlaceholder + "@" + domain
}

// MaskPhone masks a phone number for logging, preserving the country code,
// first national digit, and last three digits.
func MaskPhone(s string) string {
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "+") {
		return maskPlaceholder
	}
	digits := digitsOnly(s)
	if len(digits) < 6 {
		return maskPlaceholder
	}
	cc := digits[:2]
	first := digits[2:3]
	last3 := digits[len(digits)-3:]
	return "+" + cc + " " + first + "** **** *" + last3
}

// MaskOTPRef masks an OTP-reference / session-id identifier. Only a short
// prefix is kept for in-log correlation.
func MaskOTPRef(s string) string {
	if s == "" {
		return ""
	}
	const keep = 6
	if len(s) <= keep {
		return maskPlaceholder
	}
	return s[:keep] + maskPlaceholder
}

// MaskByKey masks value when the key suggests it carries PII. Email-shaped keys
// use MaskEmail; phone-shaped keys use MaskPhone; other keys pass through.
func MaskByKey(key, value string) string {
	k := strings.ToLower(key)
	switch {
	case strings.Contains(k, "email") || k == "to" || k == "recipient":
		return MaskEmail(value)
	case strings.Contains(k, "phone") || strings.Contains(k, "msisdn"):
		return MaskPhone(value)
	default:
		return value
	}
}

func digitsOnly(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
