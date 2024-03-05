package writers

import (
	"io"
	"path/filepath"

	"github.com/robbyt/llm_proxy/addons/fileUtils"
	md "github.com/robbyt/llm_proxy/addons/megadumper"

	log "github.com/sirupsen/logrus"
)

type ToFile struct {
	targetFileName string
	writer         io.Writer
}

func (t *ToFile) Write(identifier string, bytes []byte) (int, error) {
	bytesWritten, err := t.writer.Write(bytes)
	if err != nil {
		return bytesWritten, err
	}

	err = t.close()
	if err != nil {
		return bytesWritten, err
	}

	return bytesWritten, nil
}

func (t *ToFile) close() error {
	if closer, ok := t.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func newToFile(target string, logFormat md.LogFormat) (*ToFile, error) {
	var fileName string
	var targetExtension string = logFormat.FileExtension() // example: .json or .log

	// add file extension to the target file, if needed
	if filepath.Ext(target) != targetExtension {
		fileName = filepath.Join(target, logFormat.FileExtension())
		log.Warnf("Adding file extension to target: %v", fileName)
	} else {
		fileName = target
	}

	f, err := fileUtils.CreateNewFileFromFilename(fileName)
	if err != nil {
		return nil, err
	}

	return &ToFile{
		targetFileName: fileName,
		writer:         f,
	}, nil
}

func NewToFile(target string, logFormat md.LogFormat) (MegaDumpWriter, error) {
	return newToFile(target, logFormat)
}
