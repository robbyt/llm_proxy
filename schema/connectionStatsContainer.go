package schema

import (
	"encoding/json"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

const UnknownAddr = "unknown"

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

func (obj *ConnectionStatsContainer) ToJSON() []byte {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Failed to marshal object to JSON: %v", err)
		return []byte("{}")
	}
	return jsonData
}

func (obj *ConnectionStatsContainer) ToJSONstr() string {
	return string(obj.ToJSON())
}

func getClientAddr(f *px.Flow) string {
	if f == nil || f.ConnContext == nil || f.ConnContext.ClientConn == nil || f.ConnContext.ClientConn.Conn == nil {
		// Ugh != nil
		return UnknownAddr
	}
	remote := f.ConnContext.ClientConn.Conn.RemoteAddr()
	if remote == nil {
		return UnknownAddr
	}
	return remote.String()
}

func newConnectionStatusContainer(f *px.Flow) *ConnectionStatsContainer {
	logOutput := &ConnectionStatsContainer{
		ClientAddress: getClientAddr(f),
		Method:        f.Request.Method,
		URL:           f.Request.URL.String(),
		ProxyID:       f.Id.String(),
	}

	if f.Response != nil {
		logOutput.StatusCode = f.Response.StatusCode
	}

	if f.Response != nil && f.Response.Body != nil {
		logOutput.ContentLength = len(f.Response.Body)
	}

	if f.Response != nil && f.Response.Header != nil {
		logOutput.ContentType = f.Response.Header.Get("Content-Type")
		logOutput.XreqID = f.Response.Header.Get("X-Request-Id")
	}

	return logOutput
}

// NewConnectionStatusContainerWithDuration is a slightly leaky abstraction, the doneAt param is for logging
// the entire session length, and comes from the proxy addon layer.
func NewConnectionStatusContainerWithDuration(f px.Flow, doneAt int64) *ConnectionStatsContainer {
	logOutput := newConnectionStatusContainer(&f)
	logOutput.Duration = doneAt
	return logOutput
}
