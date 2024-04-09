package fileUtils

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// CreateNewFileFromFilename given a file name, this creates a new file on-disk for writing logs.
func CreateNewFileFromFilename(fileName string) (*os.File, error) {
	log.Debugf("Creating/opening file: %v", fileName)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %v: %w", fileName, err)
	}
	return file, nil
}
