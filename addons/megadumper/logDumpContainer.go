package megadumper

import (
	"bytes"
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// LogDumpContainer holds the request and response data for a given flow
type LogDumpContainer struct {
	RequestHeaders  string   `json:"request_headers,omitempty"`
	RequestBody     string   `json:"request_body,omitempty"`
	ResponseHeaders string   `json:"response_headers,omitempty"`
	ResponseBody    string   `json:"response_body,omitempty"`
	flow            *px.Flow `json:"-"`
}

func (d *LogDumpContainer) loadRequestHeaders() error {
	if d.flow == nil {
		return fmt.Errorf("flow is nil, unable to extract request headers")
	}
	if d.flow.Request == nil {
		return fmt.Errorf("request is nil, unable to extract request headers")
	}
	if d.flow.Request.Header == nil {
		return fmt.Errorf("request headers are nil, unable to extract request headers")
	}

	buf := new(bytes.Buffer)
	if err := d.flow.Request.Header.WriteSubset(buf, nil); err != nil {
		log.Error(err)
	} else {
		d.RequestHeaders = buf.String()
	}
	return nil
}

func (d *LogDumpContainer) loadRequestBody() error {
	if d.flow == nil {
		return fmt.Errorf("flow is nil, unable to extract request body")
	}
	if d.flow.Request == nil {
		return fmt.Errorf("request is nil, unable to extract request body")
	}
	if d.flow.Request.Body == nil {
		return fmt.Errorf("request body is nil, unable to extract request body")
	}

	// TODO CanPrint converts to a string, so there's no point in doing it twice
	if len(d.flow.Request.Body) > 0 && CanPrint(d.flow.Request.Body) {
		d.RequestBody = string(d.flow.Request.Body)
	}
	return nil
}

func (d *LogDumpContainer) loadResponseHeaders() error {
	if d.flow == nil {
		return fmt.Errorf("flow is nil, unable to extract response headers")
	}
	if d.flow.Response == nil {
		return fmt.Errorf("response is nil, unable to extract response headers")
	}
	if d.flow.Response.Header == nil {
		return fmt.Errorf("response headers are nil, unable to extract response headers")
	}

	buf := new(bytes.Buffer)
	if err := d.flow.Response.Header.WriteSubset(buf, nil); err != nil {
		log.Error(err)
	} else {
		d.ResponseHeaders = buf.String()
	}
	return nil
}

func (d *LogDumpContainer) loadResponseBody() error {
	if d.flow == nil {
		return fmt.Errorf("flow is nil, unable to extract response body")
	}
	if d.flow.Response == nil {
		return fmt.Errorf("response is nil, unable to extract response body")
	}
	if d.flow.Response.Body == nil {
		return fmt.Errorf("response body is nil, unable to extract response body")
	}

	if d.flow.Response.Body != nil && len(d.flow.Response.Body) > 0 && d.flow.Response.IsTextContentType() {
		body, err := d.flow.Response.DecodedBody()
		if err == nil && body != nil && len(body) > 0 {
			d.ResponseBody = string(body)
		}
	}
	return nil
}

// NewLogDumpContainer returns a LogDumpContainer with *only* the fields requested in logSources populated
func NewLogDumpContainer(f *px.Flow, logSources []LogSource) *LogDumpContainer {
	dumpContainer := &LogDumpContainer{flow: f}
	errors := make([]error, len(logSources))

	for _, lsource := range logSources {
		switch lsource {
		case LogRequestHeaders:
			log.Debug("Dumping request headers")
			err := dumpContainer.loadRequestHeaders()
			if err != nil {
				errors = append(errors, err)
			}

		case LogRequestBody:
			log.Debug("Dumping request body")
			err := dumpContainer.loadRequestBody()
			if err != nil {
				errors = append(errors, err)
			}

		case LogResponseHeaders:
			log.Debug("Dumping response headers")
			err := dumpContainer.loadResponseHeaders()
			if err != nil {
				errors = append(errors, err)
			}

		case LogResponseBody:
			log.Debug("Dumping response body")
			err := dumpContainer.loadResponseBody()
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	for _, err := range errors {
		if err != nil {
			// TODO: need to consider how to handle errors here
			log.Error(err)
		}
	}

	return dumpContainer
}
