package utils

import (
	"unicode"
)

// CanPrint returns a bool if a byte array can be "printed" or written to a log file (filters binary data)
func CanPrint(content []byte) bool {
	for _, c := range string(content) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// CanPrintFast is faster than CanPrint, but less accurate (it only checks the first chunk of the content)
func CanPrintFast(content []byte) (string, bool) {
	contentStr := string(content)
	threshold := len(contentStr) / 3
	if threshold < 1 {
		threshold = 4
	}

	for i, c := range contentStr {
		if i > threshold {
			// reached the threshold, assume it's printable
			break
		}

		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return "", false
		}
	}
	return contentStr, true
}
