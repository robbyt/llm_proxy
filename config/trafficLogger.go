// Config related to the *output* of the proxy, for writing request logs
package config

type trafficLogger struct {
	OutputDir           string   // Directory to write logs
	WriteJsonFormatLogs bool     // if true, write logs in JSON format
	LogReqHeaders       bool     // if true, log request headers
	LogReqBody          bool     // if true, log request body
	LogRespHeaders      bool     // if true, log response headers
	LogRespBody         bool     // if true, log response body
	FilterReqHeaders    []string // if set, request headers that match these strings will not be logged
	FilterRespHeaders   []string // if set, response headers that match these strings will not be logged
}
