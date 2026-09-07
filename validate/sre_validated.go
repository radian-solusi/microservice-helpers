package validate

func IsSREIpo(sre string) bool {
	// 004 or 064
	if len(sre) < 5 {
		return false
	}

	last5 := sre[len(sre)-5:]
	kodeAngka := last5[:3]

	switch kodeAngka {
	case "004", "064":
		return true
	default:
		return false
	}
}

func IsSREEbus(sre string) bool {
	// 074
	if len(sre) < 5 {
		return false
	}

	last5 := sre[len(sre)-5:]
	kodeAngka := last5[:3]
	return kodeAngka == "074"
}
