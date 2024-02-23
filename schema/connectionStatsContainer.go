package schema

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type ConnectionStatsContainer struct {
	ClientAddress string `json:"client_address"`
	Method        string `json:"method"`
	URL           string `json:"url"`
	StatusCode    int    `json:"status_code"`
	ContentLength int    `json:"content_length"`
	Duration      int64  `json:"duration_ms"`
	ContentType   string `json:"content_type,omitempty"`
	XreqID        string `json:"x_request_id,omitempty"`
	ProxyID       string `json:"proxy_id,omitempty"`
}

func (obj *ConnectionStatsContainer) ToJSONstr() string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Failed to marshal object to JSON: %v", err)
	}
	return string(jsonData)
}
