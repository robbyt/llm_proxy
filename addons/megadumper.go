package addons

import (
	"fmt"
	"io"
	"os"

	"github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	md "github.com/robbyt/llm_proxy/addons/megadumper"
)

type MegaDumper struct {
	proxy.BaseAddon
	singleLogFileTarget io.Writer
	logFilename         string
	logTarget           string
	logLevel            md.LogLevel
	logFormat           md.LogFormat
}

func (d *MegaDumper) logExtension() string {
	switch d.logFormat {
	case md.LogFormat_JSON:
		return "json"
	case md.LogFormat_PLAINTEXT:
		return "log"
	default:
		return ""
	}
}

// Requestheaders is a callback for the Requestheaders event
func (d *MegaDumper) Requestheaders(f *proxy.Flow) {
	go func() {
		<-f.Done()
		d.writeLog(f)
	}()
}

func (d MegaDumper) writeLog(f *proxy.Flow) error {
	var logDump md.MegaDumperWriter
	var err error

	switch d.logFormat {
	case md.LogFormat_JSON:
		logDump, err = md.NewLogDumpDiskContainer_JSON(f, d.logLevel)
	case md.LogFormat_PLAINTEXT:
		logDump, err = md.NewLogDumpDiskContainer_Bytes(f, d.logLevel)
	}

	if err != nil {
		return err
	}
	bytesToWrite, err := logDump.Read()
	if err != nil {
		return err
	}

	if d.logTarget != "" {
		// multiple log file mode enabled, will create a new singleLogFileTarget for each request
		d.logFilename = fmt.Sprintf("%s/%s.%s", d.logTarget, f.Id, d.logExtension())

		// check if the file exists
		_, err := os.Stat(d.logFilename)
		if err == nil {
			log.Warnf("log file already exists, appending: %v", d.logFilename)
		}

		d.singleLogFileTarget, err = md.NewFile(d.logFilename)
		if err != nil {
			return err
		}
	}

	return d.diskWriter(bytesToWrite)
}

func (d MegaDumper) diskWriter(bytes []byte) error {
	if d.singleLogFileTarget == nil {
		return fmt.Errorf("internal error: singleLogFileTarget is not set")
	}

	log.Debugf("Writing to log file: %v", d.logFilename)
	_, err := d.singleLogFileTarget.Write(bytes)
	return err
}

// newDumper is an abstract factory for creating a MegaDumper, configured for logging to a single file
func newDumper(out io.Writer, lvl md.LogLevel, logFormat md.LogFormat) *MegaDumper {
	return &MegaDumper{singleLogFileTarget: out, logLevel: lvl, logFormat: logFormat}
}

// NewDumperWithFilename creates a dumper, and a single file, and writes all logs to that one file
func NewDumperWithFilename(filename string, lvl md.LogLevel, logFormat md.LogFormat) (*MegaDumper, error) {
	out, err := md.NewFile(filename)
	if err != nil {
		return nil, err
	}
	return newDumper(out, lvl, logFormat), nil
}

// NewDumperWithLogRoot creates a new dumper that creates a new log file for each request
func NewDumperWithLogRoot(logRoot string, lvl md.LogLevel, logFormat md.LogFormat) (*MegaDumper, error) {
	// Check if the log directory exists
	_, err := os.Stat(logRoot)
	if err != nil {
		if os.IsNotExist(err) {
			// If it doesn't exist, create it
			err := os.MkdirAll(logRoot, 0750)
			if err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}
			log.Infof("Log directory %v created successfully", logRoot)
		} else {
			// If os.Stat failed for another reason, return the error
			return nil, fmt.Errorf("failed to check if log directory exists: %w", err)
		}
	}

	return &MegaDumper{
		logLevel:  lvl,
		logTarget: logRoot,
		logFormat: logFormat,
	}, nil
}
