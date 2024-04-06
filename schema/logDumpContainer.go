package schema

import (
	"fmt"
	"net/url"
	"time"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/config"
	"github.com/robbyt/llm_proxy/schema/utils"
)

const SchemaVersion string = "v2"

// LogDumpContainer holds the request and response data for a given flow
type LogDumpContainer struct {
	SchemaVersion   string                    `json:"schema,omitempty"`
	Timestamp       time.Time                 `json:"timestamp,omitempty"`
	ConnectionStats *ConnectionStatsContainer `json:"connection_stats,omitempty"`
	Request         *TrafficObject            `json:"request,omitempty"`
	Response        *TrafficObject            `json:"response,omitempty"`
	flow            *px.Flow                  `json:"-"`
}

func (d *LogDumpContainer) loadRequestMethod() error {
	if d.flow.Request == nil {
		return fmt.Errorf("request is nil, unable to extract request method")
	}
	d.Request.Method = d.flow.Request.Method
	return nil
}

func (d *LogDumpContainer) loadRequestURL() error {
	if d.flow.Request == nil || d.flow.Request.URL == nil {
		return fmt.Errorf("request URL is nil, unable to extract request URL")
	}
	d.Request.URL = d.flow.Request.URL
	return nil
}

func (d *LogDumpContainer) loadRequestProto() error {
	if d.flow.Request == nil {
		return fmt.Errorf("request is nil, unable to extract request proto")
	}
	d.Request.Proto = d.flow.Request.Proto
	return nil
}

func (d *LogDumpContainer) loadRequestHeaders() {
	d.Request.Header = d.flow.Request.Header
}

func (d *LogDumpContainer) loadRequestBody() error {
	// TODO CanPrint converts to a string, so there's no point in doing it twice
	bodyStr, canPrint := utils.CanPrintFast(d.flow.Request.Body)
	if !canPrint {
		return fmt.Errorf("request body is not printable: %s", d.flow.Request.URL)
	}
	d.Request.Body = bodyStr
	return nil
}

func (d *LogDumpContainer) loadResponseStatusCode() error {
	if d.flow.Response == nil {
		return fmt.Errorf("response is nil, unable to extract response status code")
	}
	d.Response.StatusCode = d.flow.Response.StatusCode
	return nil
}

func (d *LogDumpContainer) loadResponseHeaders() error {
	if d.flow.Response.Header == nil {
		return fmt.Errorf("response headers are nil, unable to extract response headers")
	}
	d.Response.Header = d.flow.Response.Header
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
		logSources.LogRequest = false
	} else if f.Request.Header == nil {
		log.Error("request headers are nil, disabling request headers extraction")
		logSources.LogRequestHeaders = false
	} else if f.Request.Body == nil {
		log.Error("request body is nil, disabling request body extraction")
		logSources.LogRequest = false
	}

	// response validation
	if f.Response == nil {
		log.Error("response is nil, disabling response data extraction")
		logSources.LogResponseHeaders = false
		logSources.LogResponse = false
	} else if f.Response.Header == nil {
		log.Error("response headers are nil, disabling response headers extraction")
		logSources.LogResponseHeaders = false
	} else if f.Response.Body == nil {
		log.Error("response body is nil, disabling response body extraction")
		logSources.LogResponse = false
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
		Request: &TrafficObject{
			headersToFilter: filterReqHeaders,
		},
		Response: &TrafficObject{
			headersToFilter: filterRespHeaders,
		},
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

	if logSources.LogRequest {
		log.Debug("Dumping request body")
		if err := dumpContainer.loadRequestMethod(); err != nil {
			errors = append(errors, err)
		}
		if err := dumpContainer.loadRequestURL(); err != nil {
			errors = append(errors, err)
		}
		if err := dumpContainer.loadRequestProto(); err != nil {
			errors = append(errors, err)
		}
		if err := dumpContainer.loadRequestBody(); err != nil {
			errors = append(errors, err)
		}
	} else {
		// NPE defense
		dumpContainer.Request.URL = &url.URL{}
	}

	if logSources.LogResponseHeaders {
		log.Debug("Dumping response headers")
		if dumpContainer.flow.Response != nil && dumpContainer.flow.Response.Header != nil {
			err := dumpContainer.loadResponseHeaders()
			if err != nil {
				errors = append(errors, err)
			} else {
				dumpContainer.Response.filterHeaders()
			}
		}
	}

	if logSources.LogResponse {
		log.Debug("Dumping response body")
		err := dumpContainer.loadResponseBody()
		if err != nil {
			errors = append(errors, err)
		}
		err = dumpContainer.loadResponseStatusCode()
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
