package config

import "fmt"

// LogSourceConfig holds the configuration toggles for logging request and response data
type LogSourceConfig struct {
	LogRequestHeaders  bool
	LogRequestBody     bool
	LogResponseHeaders bool
	LogResponseBody    bool
}

func (l *LogSourceConfig) String() string {
	return fmt.Sprintf("LogRequestHeaders: %v, LogRequestBody: %v, LogResponseHeaders: %v, LogResponseBody: %v",
		l.LogRequestHeaders, l.LogRequestBody, l.LogResponseHeaders, l.LogResponseBody)
}
