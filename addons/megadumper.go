package addons

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"

	"github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// LogLevel is an enum for the logging style of the dumper
type LogLevel int

const (
	// logs only request headers
	WRITE_REQ_HEADERS_ONLY LogLevel = iota

	// logs both request and response bodies, this is the most common use case
	WRITE_REQ_BODY_AND_RESP_BODY

	// logs request headers, response headers, and response body
	WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY

	// logs request headers, request body, response headers, and response body
	WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY
)

type MegaDumper struct {
	proxy.BaseAddon
	singleLogFileTarget io.Writer
	logFilename         string
	logTarget           string
	logLevel            LogLevel
}

func (d *MegaDumper) Requestheaders(f *proxy.Flow) {
	go func() {
		<-f.Done()
		d.Write(f)
	}()
}

func (d MegaDumper) Write(f *proxy.Flow) error {
	bytesToWrite, err := d.prepDumpBytes(f)
	if err != nil {
		return err
	}

	if d.logTarget != "" {
		// multiple log file mode enabled, will create a new singleLogFileTarget for each request
		d.logFilename = fmt.Sprintf("%s/%s.log", d.logTarget, f.Id)

		// check if the file exists
		_, err := os.Stat(d.logFilename)
		if err == nil {
			log.Warnf("log file already exists, appending: %v", d.logFilename)
		}

		d.singleLogFileTarget, err = newFile(d.logFilename)
		if err != nil {
			return err
		}
	}

	return d.diskWriter(bytesToWrite)
}

func (d MegaDumper) diskWriter(buf *bytes.Buffer) error {
	if d.singleLogFileTarget == nil {
		return fmt.Errorf("internal error: singleLogFileTarget is not set")
	}

	log.Debugf("Writing to log file: %v", d.logFilename)
	_, err := d.singleLogFileTarget.Write(buf.Bytes())
	return err
}

// prepDumpBytes is a blocking call, run by .Write after <-f.Done()
func (d *MegaDumper) prepDumpBytes(f *proxy.Flow) (*bytes.Buffer, error) {
	// Reference httputil.DumpRequest.
	buf := bytes.NewBuffer(make([]byte, 0))

	// request headers
	switch d.logLevel {
	case WRITE_REQ_HEADERS_ONLY, WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping request headers")
		if f.Request != nil {
			err := f.Request.Header.WriteSubset(buf, nil) // writing response headers
			if err != nil {
				// continue here, if unable to store the full response
				log.Error(err)
			}
			buf.WriteString("\r\n")
		} else {
			log.Error("request is nil, unable to write request headers")
		}
	}

	// request body
	switch d.logLevel {
	case WRITE_REQ_BODY_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping request body")
		if f.Request != nil {
			if f.Request != nil && len(f.Request.Body) > 0 && canPrint(f.Request.Body) {
				/*
					bod, err := f.Request.Raw().GetBody()
					if err != nil {
						log.Error(err)
						break
					}

					_, err = io.ReadAll(bod)
					if err != nil {
						log.Error(err)
						break
					}

					// buf.Write(bodyBytes)
					//
				*/
				_, err := buf.Write(f.Request.Body)
				if err != nil {
					log.Error(err)
					break
				}
				buf.WriteString("\r\n\r\n")
			}
		} else {
			log.Error("request is nil, unable to write request headers")
		}
	}

	// response headers
	switch d.logLevel {
	case WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping response headers")
		if f.Response != nil {
			err := f.Response.Header.WriteSubset(buf, nil) // writing response headers
			if err != nil {
				// continue here, if unable to store the full response
				log.Error(err)
			}
			buf.WriteString("\r\n")

		}
	}

	// response body
	switch d.logLevel {
	case WRITE_REQ_BODY_AND_RESP_BODY, WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping response body")
		if f.Response.Body != nil && len(f.Response.Body) > 0 && f.Response.IsTextContentType() {
			body, err := f.Response.DecodedBody()
			if err == nil && body != nil && len(body) > 0 {
				buf.Write(body)
				buf.WriteString("\r\n\r\n")
			}
		}
	}

	buf.WriteString("\r\n\r\n")
	return buf, nil
}

func canPrint(content []byte) bool {
	for _, c := range string(content) {
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func newFile(fileName string) (*os.File, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %v: %w", fileName, err)
	}
	return file, nil
}

func newDumper(out io.Writer, lvl LogLevel) *MegaDumper {
	return &MegaDumper{singleLogFileTarget: out, logLevel: lvl}
}

// NewDumperWithFilename creates a dumper that writes all logs to a single file
func NewDumperWithFilename(filename string, lvl LogLevel) (*MegaDumper, error) {
	out, err := newFile(filename)
	if err != nil {
		return nil, err
	}
	return newDumper(out, lvl), nil
}

// NewDumperWithLogRoot creates a new dumper that creates a new log file for each request
func NewDumperWithLogRoot(logRoot string, lvl LogLevel) (*MegaDumper, error) {
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

	return &MegaDumper{logLevel: lvl, logTarget: logRoot}, nil
}
