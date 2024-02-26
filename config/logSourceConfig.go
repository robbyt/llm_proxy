package config

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// LogSourceConfig holds the configuration toggles for logging request and response data
type LogSourceConfig struct {
	LogConnectionStats bool
	LogRequestHeaders  bool
	LogRequestBody     bool
	LogResponseHeaders bool
	LogResponseBody    bool
}

func (l *LogSourceConfig) String() string {
	bytes, err := json.Marshal(l)
	if err != nil {
		log.Error(fmt.Sprintf("Error marshalling LogSourceConfig: %v", err))
		return ""
	}
	return string(bytes)
}
