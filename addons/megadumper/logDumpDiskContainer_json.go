package megadumper

import (
	"bytes"
	"encoding/json"
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// LogDumpDiskContainer_JSON is a struct for holding
type LogDumpDiskContainer_JSON struct {
	RequestHeaders  string   `json:"request_headers,omitempty"`
	RequestBody     string   `json:"request_body,omitempty"`
	ResponseHeaders string   `json:"response_headers,omitempty"`
	ResponseBody    string   `json:"response_body,omitempty"`
	logLevel        LogLevel `json:"-"`
}

// dumpToJSONBytes converts the requestLogDump struct to a byte array, omitting fields that are empty
func (d *LogDumpDiskContainer_JSON) dumpToJSONBytes() ([]byte, error) {
	j, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal requestLogDump to JSON: %w", err)
	}
	return j, nil
}

// Read returns the JSON representation of the LogDumpDiskContainer_JSON (json formatted byte array)
func (d *LogDumpDiskContainer_JSON) Read() ([]byte, error) {
	// TODO only return valid fields by logLevel
	return d.dumpToJSONBytes()
}

// NewLogDumpDiskContainer_JSON takes a proxy.Flow and a LogLevel, and returns a LogDumpDiskContainer which can be used for dumping to JSON
func newLogDumpDiskContainer_JSON(f *px.Flow, logLevel LogLevel) (*LogDumpDiskContainer_JSON, error) {
	dumpContainer := &LogDumpDiskContainer_JSON{logLevel: logLevel}

	// request headers
	switch logLevel {
	case WRITE_REQ_HEADERS_ONLY, WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping request headers")
		if f.Request != nil {
			buf := new(bytes.Buffer)
			if err := f.Request.Header.WriteSubset(buf, nil); err != nil {
				log.Error(err)
			} else {
				dumpContainer.RequestHeaders = buf.String()
			}
		} else {
			log.Error("request is nil, unable to write request headers")
		}
	}

	// request body
	switch logLevel {
	case WRITE_REQ_BODY_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping request body")
		if f.Request != nil {
			// TODO: CanPrint converts to a string, so no need to do it twice
			if f.Request != nil && len(f.Request.Body) > 0 && CanPrint(f.Request.Body) {
				dumpContainer.RequestBody = string(f.Request.Body)
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
			buf := new(bytes.Buffer)
			err := f.Response.Header.WriteSubset(buf, nil) // writing response headers
			if err != nil {
				// continue here, if unable to store the full response
				log.Error(err)
			} else {
				dumpContainer.ResponseHeaders = buf.String()
			}

		}
	}

	// response body
	switch logLevel {
	case WRITE_REQ_BODY_AND_RESP_BODY, WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY:
		log.Debug("Dumping response body")
		if f.Response.Body != nil && len(f.Response.Body) > 0 && f.Response.IsTextContentType() {
			body, err := f.Response.DecodedBody()
			if err == nil && body != nil && len(body) > 0 {
				dumpContainer.ResponseBody = string(body)
			}
		}
	}

	return dumpContainer, nil
}

func NewLogDumpDiskContainer_JSON(f *px.Flow, logLevel LogLevel) (MegaDumperWriter, error) {
	return newLogDumpDiskContainer_JSON(f, logLevel)
}
