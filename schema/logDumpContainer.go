package schema

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/schema/utils"
)

const SchemaVersion string = "v1"

type TrafficObject struct {
	Headers         http.Header `json:"headers"`
	Body            string      `json:"body"`
	headersToFilter []string    `json:"-"`
}

// HeadersString returns the headers as a flat string
func (t *TrafficObject) HeadersString() string {
	buf := new(bytes.Buffer)
	if err := t.Headers.WriteSubset(buf, nil); err != nil {
		return ""
	}
	return buf.String()
}

func (t *TrafficObject) filterHeaders() {
	log.Debugf("Filtering headers from log output: %v", t.headersToFilter)
	for _, header := range t.headersToFilter {
		t.Headers.Del(header)
	}
}

// LogDumpContainer holds the request and response data for a given flow
type LogDumpContainer struct {
	SchemaVersion   string                    `json:"schema,omitempty"`
	Timestamp       time.Time                 `json:"timestamp,omitempty"`
	ConnectionStats *ConnectionStatsContainer `json:"connection_stats,omitempty"`
	Request         *TrafficObject            `json:"request,omitempty"`
	Response        *TrafficObject            `json:"response,omitempty"`
	flow            *px.Flow                  `json:"-"`
}

func (d *LogDumpContainer) loadRequestHeaders() {
	d.Request.Headers = d.flow.Request.Header
}

func (d *LogDumpContainer) loadRequestBody() error {
	// TODO CanPrint converts to a string, so there's no point in doing it twice
	if utils.CanPrint(d.flow.Request.Body) {
		d.Request.Body = string(d.flow.Request.Body)
	}
	return nil
}

func (d *LogDumpContainer) loadResponseHeaders() {
	d.Response.Headers = d.flow.Response.Header
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
		d.Response.Body = string(body)
	}
	return nil
}

// validateFlowObj checks if various fields in the f flow are populated, and adjusts the logSources object accordingly
func validateFlowObj(f *px.Flow, logSources config.LogSourceConfig) config.LogSourceConfig {
	if f == nil {
		log.Error("flow is nil, unable to extract data")
		return config.LogSourceConfig{}
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
func NewLogDumpContainer(f px.Flow, logSources config.LogSourceConfig, doneAt int64, filterReqHeaders, filterRespHeaders []string) *LogDumpContainer {
	logSources = validateFlowObj(&f, logSources) // disable logging of fields that are not present in the flow
	dumpContainer := &LogDumpContainer{
		SchemaVersion: SchemaVersion,
		Timestamp:     time.Now(),
		flow:          &f,
		Request:       &TrafficObject{headersToFilter: filterReqHeaders},
		Response:      &TrafficObject{headersToFilter: filterRespHeaders},
	}
	errors := make([]error, 0)

	if logSources.LogConnectionStats {
		log.Debug("Dumping connection stats")
		dumpContainer.ConnectionStats = NewConnectionStatusContainerWithDuration(f, doneAt)
	}

	if logSources.LogRequestHeaders {
		log.Debug("Dumping request headers")
		if dumpContainer.flow.Request != nil && dumpContainer.flow.Request.Header != nil {
			dumpContainer.loadRequestHeaders()
			dumpContainer.Request.filterHeaders()
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
		if dumpContainer.flow.Response != nil && dumpContainer.flow.Response.Header != nil {
			dumpContainer.loadResponseHeaders()
			dumpContainer.Response.filterHeaders()
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
