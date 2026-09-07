package validate

import "regexp"

var pattern = regexp.MustCompile(`^[A-Za-z]{2}\d+[A-Za-z]{2}(\d{3})\d*$`)

func IsSREIpo(sre string) bool {
	// 004 or 064
	matches := pattern.FindStringSubmatch(sre)
	if len(matches) < 2 {
		return false
	}

	switch matches[1] {
	case "004", "064":
		return true
	default:
		return false
	}
}

func IsSREEbus(sre string) bool {
	// 074
	matches := pattern.FindStringSubmatch(sre)
	if len(matches) < 2 {
		return false
	}
	return matches[1] == "074"
}
