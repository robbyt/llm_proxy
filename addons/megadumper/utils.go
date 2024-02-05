package megadumper

import (
	"fmt"
	"os"
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

// NewFile given a file name, this creates a new file on-disk for writing logs.
func NewFile(fileName string) (*os.File, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %v: %w", fileName, err)
	}
	return file, nil
}
