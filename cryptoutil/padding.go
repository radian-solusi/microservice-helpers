package cryptoutil

import "bytes"

func Unpad(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	n := int(src[len(src)-1])
	if n > len(src) {
		return nil
	}
	return src[:len(src)-n]
}

func ZeroUnpad(src []byte) []byte {
	return bytes.TrimRight(src, "\x00")
}
