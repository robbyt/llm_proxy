package megadumper

// LogFormat is an enum for the format of the logs
type LogFormat int

const (
	// logs in LogFormat_JSON format
	LogFormat_JSON LogFormat = iota

	// logs in plain text format
	LogFormat_PLAINTEXT
)
