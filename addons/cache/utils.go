package cache

import (
	"encoding/base64"
	"path/filepath"
	"strings"
)

// ConvertIDtoFileName converts a ID string into a filename string by replacing several characters
func ConvertIDtoFileName(dbFileDir, identifier string) string {
	identifier = strings.ReplaceAll(identifier, "https://", "")
	identifier = strings.ReplaceAll(identifier, "http://", "")
	encodedString := base64.URLEncoding.EncodeToString([]byte(identifier))
	return filepath.Join(dbFileDir, encodedString)
}
