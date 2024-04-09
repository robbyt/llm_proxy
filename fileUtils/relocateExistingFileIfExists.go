package fileUtils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// RelocateExistingFileIfExists checks if a file exists, and if it does, renames it to a unique name
func RelocateExistingFileIfExists(fileName string) error {
	if FileExists(fileName) {
		relocatedFile := CreateUniqueFileName(filepath.Dir(fileName), filepath.Base(fileName), "", 0)
		log.Warnf("file already exists, relocating: %s -> %s", fileName, relocatedFile)
		return os.Rename(fileName, relocatedFile)
	}
	return nil
}
