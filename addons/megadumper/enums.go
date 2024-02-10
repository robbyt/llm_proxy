package megadumper

// LogSource is an enum for selecting which fields will be stored in the output logs
type LogSource int

const (
	// LogRequestHeaders logs the request headers
	LogRequestHeaders LogSource = iota

	// LogRequestBody logs the request body
	LogRequestBody

	// LogResponseHeaders logs the response headers
	LogResponseHeaders

	// LogResponseBody logs the response body
	LogResponseBody
)

func (s LogSource) String() string {
	switch s {
	case LogRequestHeaders:
		return "RequestHeaders"
	case LogRequestBody:
		return "RequestBody"
	case LogResponseHeaders:
		return "ResponseHeaders"
	case LogResponseBody:
		return "ResponseBody"
	default:
		return ""
	}
}

// LogSourceFromBools returns a slice of LogSource enums based on the boolean values passed in (from cfg)
func LogSourceFromBools(logReqHeaders, logReqBody, logRespHeaders, logRespBody bool) []LogSource {
	var sources []LogSource
	if logReqHeaders {
		sources = append(sources, LogRequestHeaders)
	}
	if logReqBody {
		sources = append(sources, LogRequestBody)
	}
	if logRespHeaders {
		sources = append(sources, LogResponseHeaders)
	}
	if logRespBody {
		sources = append(sources, LogResponseBody)
	}
	return sources
}

// LogFormat is an enum for the format of the logs
type LogFormat int

func (f LogFormat) FileExtension() string {
	switch f {
	case Format_JSON:
		return "json"
	case Format_PLAINTEXT:
		return "log"
	default:
		return ""
	}
}

const (
	// Format_JSON logs in Format_JSON format
	Format_JSON LogFormat = iota

	// Format_PLAINTEXT logs in plain text format
	Format_PLAINTEXT
)

// LogDestination is an enum for the destination for where the logs are stored
type LogDestination int

const (
	// WriteToFile logs to a single file
	WriteToFile LogDestination = iota

	// WriteToDir logs to a directory
	WriteToDir
)
