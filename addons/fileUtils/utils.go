package fileUtils

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CreateUniqueFileName creates a new unused filename at the target directory.
// targetDir: the directory to create the file in.
// identifier: the base name of the file.
// fileExtension: the file extension suffix.
// attempt: the current attempt number, to prevent infinite loops.
func CreateUniqueFileName(
	targetDir, identifier, fileExtension string,
	attempt int,
) string {
	// check if the file already exists
	var fileName string
	if attempt < 1 {
		fileName = filepath.Join(targetDir, fmt.Sprintf("%s.%s", identifier, fileExtension))
	} else if attempt > 100000 {
		panic("Too many attempts to create a unique filename, crashing to prevent infinite loop.")
	} else {
		fileName = filepath.Join(targetDir, fmt.Sprintf("%s-%v.%s", identifier, attempt, fileExtension))
	}

	if FileExists(fileName) {
		log.Warnf("File %s already exists, trying again...", fileName)
		return CreateUniqueFileName(targetDir, identifier, fileExtension, attempt+1)
	}

	return fileName
}

func FileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

// CreateNewFileFromFilename given a file name, this creates a new file on-disk for writing logs.
func CreateNewFileFromFilename(fileName string) (*os.File, error) {
	log.Debugf("Creating/opening file: %v", fileName)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %v: %w", fileName, err)
	}
	return file, nil
}

// DirExistsOrCreate checks if a directory exists, and creates it if it doesn't
func DirExistsOrCreate(dir string) error {
	log.Debugf("Checking if directory exists: %s", dir)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// If it doesn't exist, create it
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			log.Infof("Log directory created successfully: %s", dir)
		} else {
			// If os.Stat failed for another reason, return the error
			return fmt.Errorf("failed to check if directory exists: %w", err)
		}
	}
	return nil
}

// RelocateExistingFileIfExists checks if a file exists, and if it does, renames it to a unique name
func RelocateExistingFileIfExists(fileName string) error {
	if FileExists(fileName) {
		relocatedFile := CreateUniqueFileName(filepath.Dir(fileName), filepath.Base(fileName), "", 0)
		log.Warnf("file already exists, relocating: %s -> %s", fileName, relocatedFile)
		return os.Rename(fileName, relocatedFile)
	}
	return nil
}

// ConvertIDtoFileName converts a ID string into a filename string by replacing several characters
func ConvertIDtoFileName(dbFileDir, identifier string) string {
	identifier = strings.ReplaceAll(identifier, "https://", "")
	identifier = strings.ReplaceAll(identifier, "http://", "")
	encodedString := base64.URLEncoding.EncodeToString([]byte(identifier))
	return filepath.Join(dbFileDir, encodedString)
}
