package megadumper

import (
	"unicode"
)

// CanPrintString returns a bool if a string can be "printed" or written to a log file (filters binary data)
func CanPrintString(content string) bool {
	for _, c := range content {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

// CanPrint returns a bool if a byte array can be "printed" or written to a log file (filters binary data)
func CanPrint(content []byte) bool {
	for _, c := range string(content) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}
