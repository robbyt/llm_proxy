package megadumper

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
