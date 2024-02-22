package megadumper

import (
	"bytes"
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/robbyt/llm_proxy/config"
	log "github.com/sirupsen/logrus"
)

const SchemaVersion string = "v1"

// LogDumpContainer holds the request and response data for a given flow
type LogDumpContainer struct {
	SchemaVersion     string   `json:"schema"`
	RequestHeaders    string   `json:"request_headers"`
	RequestBody       string   `json:"request_body"`
	ResponseHeaders   string   `json:"response_headers"`
	ResponseBody      string   `json:"response_body"`
	filterReqHeaders  []string `json:"-"`
	filterRespHeaders []string `json:"-"`
	flow              *px.Flow `json:"-"`
}

func (d *LogDumpContainer) loadRequestHeaders() error {
	buf := new(bytes.Buffer)
	if err := d.flow.Request.Header.WriteSubset(buf, nil); err != nil {
		// error filling the buffer from the headers
		return err
	}
	d.RequestHeaders = buf.String()
	return nil
}

func (d *LogDumpContainer) loadRequestBody() error {
	// TODO CanPrint converts to a string, so there's no point in doing it twice
	if CanPrint(d.flow.Request.Body) {
		d.RequestBody = string(d.flow.Request.Body)
	}
	return nil
}

func (d *LogDumpContainer) loadResponseHeaders() error {
	buf := new(bytes.Buffer)
	if err := d.flow.Response.Header.WriteSubset(buf, nil); err != nil {
		// error filling the buffer from the headers
		return err
	}
	d.ResponseHeaders = buf.String()
	return nil
}

func (d *LogDumpContainer) loadResponseBody() error {
	if d.flow.Response.Body == nil {
		return fmt.Errorf("response body is nil, unable to extract response body")
	}

	if !d.flow.Response.IsTextContentType() {
		return fmt.Errorf("response body is not text, unable to extract response body")
	}

	body, err := d.flow.Response.DecodedBody()
	if err != nil {
		return fmt.Errorf("error decoding response body: %s", err)
	}

	if body != nil {
		d.ResponseBody = string(body)
	}
	return nil
}

func (d *LogDumpContainer) runResponseHeadersFilter() {
	if d.flow.Response != nil || d.flow.Response.Header != nil {
		log.Debugf("Filtering response headers from log output: %v", d.filterRespHeaders)
		for _, header := range d.filterRespHeaders {
			d.flow.Response.Header.Del(header)
		}
	}
}

func (d *LogDumpContainer) runRequestHeadersFilter() {
	if d.flow.Request != nil || d.flow.Request.Header != nil {
		log.Debugf("Filtering request headers from log output: %v", d.filterReqHeaders)
		for _, header := range d.filterReqHeaders {
			d.flow.Request.Header.Del(header)
		}
	}
}

// validateFlowObj checks if various fields in the f flow are populated, and adjusts the logSources object accordingly
func validateFlowObj(f *px.Flow, logSources config.LogSourceConfig) config.LogSourceConfig {
	if f == nil {
		log.Error("flow is nil, unable to extract data")
		return config.LogSourceConfig{
			LogRequestHeaders:  false,
			LogRequestBody:     false,
			LogResponseHeaders: false,
			LogResponseBody:    false,
		}
	}

	// request validation
	if f.Request == nil {
		log.Error("request is nil, disabling request data extraction")
		logSources.LogRequestHeaders = false
		logSources.LogRequestBody = false
	} else if f.Request.Header == nil {
		log.Error("request headers are nil, disabling request headers extraction")
		logSources.LogRequestHeaders = false
	} else if f.Request.Body == nil {
		log.Error("request body is nil, disabling request body extraction")
		logSources.LogRequestBody = false
	}

	// response validation
	if f.Response == nil {
		log.Error("response is nil, disabling response data extraction")
		logSources.LogResponseHeaders = false
		logSources.LogResponseBody = false
	} else if f.Response.Header == nil {
		log.Error("response headers are nil, disabling response headers extraction")
		logSources.LogResponseHeaders = false
	} else if f.Response.Body == nil {
		log.Error("response body is nil, disabling response body extraction")
		logSources.LogResponseBody = false
	}
	return logSources
}

// NewLogDumpContainer returns a LogDumpContainer with *only* the fields requested in logSources populated
func NewLogDumpContainer(f px.Flow, logSources config.LogSourceConfig, filterReqHeaders, filterRespHeaders []string) *LogDumpContainer {
	logSources = validateFlowObj(&f, logSources) // disable logging of fields that are not present in the flow
	dumpContainer := &LogDumpContainer{
		SchemaVersion:     SchemaVersion,
		flow:              &f,
		filterReqHeaders:  filterReqHeaders,
		filterRespHeaders: filterRespHeaders,
	}
	errors := make([]error, 0)

	if logSources.LogRequestHeaders {
		log.Debug("Dumping request headers")
		dumpContainer.runRequestHeadersFilter()
		err := dumpContainer.loadRequestHeaders()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if logSources.LogRequestBody {
		log.Debug("Dumping request body")
		err := dumpContainer.loadRequestBody()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if logSources.LogResponseHeaders {
		log.Debug("Dumping response headers")
		dumpContainer.runResponseHeadersFilter()
		err := dumpContainer.loadResponseHeaders()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if logSources.LogResponseBody {
		log.Debug("Dumping response body")
		err := dumpContainer.loadResponseBody()
		if err != nil {
			errors = append(errors, err)
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
