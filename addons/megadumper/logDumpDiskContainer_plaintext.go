package megadumper

import (
	"bytes"
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// LogDumpDiskContainer is a struct for holding the various types of data contained within a single request/response
type LogDumpDiskContainer_bytes struct {
	RawLogBytes []byte   `json:"-"`
	logLevel    LogLevel `json:"-"`
}

// Read returns the rawLogBytes if it is not nil, otherwise it calls DumpToJSONBytes and returns the resulting byte array
func (d *LogDumpDiskContainer_bytes) Read() ([]byte, error) {
	if d.RawLogBytes != nil {
		return d.RawLogBytes, nil
	}
	return nil, fmt.Errorf("RawLogBytes is nil")
}

func newLogDumpDiskContainer_bytes(f *px.Flow, logLevel LogLevel) (*LogDumpDiskContainer_bytes, error) {
	dumpContainer := &LogDumpDiskContainer_bytes{logLevel: logLevel}
	// Reference httputil.DumpRequest.
	buf := bytes.NewBuffer(make([]byte, 0))

	// request headers
	switch logLevel {
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
	switch logLevel {
	case WRITE_REQ_BODY_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping request body")
		if f.Request != nil {
			if f.Request != nil && len(f.Request.Body) > 0 && CanPrint(f.Request.Body) {
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
	switch logLevel {
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
	switch logLevel {
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
	dumpContainer.RawLogBytes = buf.Bytes()
	return dumpContainer, nil
}

// NewLogDumpDiskContainer_Bytes is a factory that returns an object that implements the MegaDumperWriter interface
func NewLogDumpDiskContainer_Bytes(f *px.Flow, logLevel LogLevel) (MegaDumperWriter, error) {
	return newLogDumpDiskContainer_bytes(f, logLevel)
}
