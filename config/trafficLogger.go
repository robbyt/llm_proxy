package config

// trafficLogger handles config related to the *output* of the proxy traffic, for writing request/response logs
type trafficLogger struct {
	OutputDir           string   // Directory to write logs
	WriteJsonFormatLogs bool     // if true, write logs in JSON format
	NoLogReqHeaders     bool     // if true, log request headers
	NoLogReqBody        bool     // if true, log request body
	NoLogRespHeaders    bool     // if true, log response headers
	NoLogRespBody       bool     // if true, log response body
	FilterReqHeaders    []string // if set, request headers that match these strings will not be logged
	FilterRespHeaders   []string // if set, response headers that match these strings will not be logged
}
