package writers

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// createUniqueFileName creates a new unused filename at the target directory
func createUniqueFileName(targetDir, identifier, fileExtension string, attempt int) string {
	// check if the file already exists
	var fileName string
	if attempt < 1 {
		fileName = filepath.Join(targetDir, fmt.Sprintf("%s.%s", identifier, fileExtension))
	} else if attempt > 10000 {
		panic("Too many attempts to create a unique filename, crashing to prevent infinite loop.")
	} else {
		fileName = filepath.Join(targetDir, fmt.Sprintf("%s-%v.%s", identifier, attempt, fileExtension))
	}

	if fileExists(fileName) {
		log.Warnf("File %s already exists, trying again...", fileName)
		return createUniqueFileName(targetDir, identifier, fileExtension, attempt+1)
	}

	return fileName
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

// createNewFileFromFilename given a file name, this creates a new file on-disk for writing logs.
func createNewFileFromFilename(fileName string) (*os.File, error) {
	log.Debugf("Creating/opening file: %v", fileName)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %v: %w", fileName, err)
	}
	return file, nil
}

// dirExistsOrCreate checks if a directory exists, and creates it if it doesn't
func dirExistsOrCreate(path string) error {
	log.Debugf("Checking if log directory exists: %s", path)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If it doesn't exist, create it
			err := os.MkdirAll(path, 0750)
			if err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
			log.Infof("Log directory created successfully: %s", path)
		} else {
			// If os.Stat failed for another reason, return the error
			return fmt.Errorf("failed to check if log directory exists: %w", err)
		}
	}
	return nil
}
