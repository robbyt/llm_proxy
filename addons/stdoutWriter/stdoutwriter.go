package stdoutWriter

import (
	"encoding/json"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

const UnknownAddr = "unknown"

type LogLine struct {
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

func (obj *LogLine) ToJSONstr() string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Failed to marshal object to JSON: %v", err)
	}
	return string(jsonData)
}

func getClientAddr(f *px.Flow) string {
	if f == nil || f.ConnContext == nil || f.ConnContext.ClientConn == nil || f.ConnContext.ClientConn.Conn == nil {
		return UnknownAddr
	}
	remote := f.ConnContext.ClientConn.Conn.RemoteAddr()
	if remote == nil {
		return UnknownAddr
	}
	return remote.String()
}

func NewLogLine(f *px.Flow, doneAt int64) *LogLine {
	if f == nil {
		log.Error("Flow object is nil")
		return nil
	}

	logOutput := &LogLine{
		ClientAddress: getClientAddr(f),
		Method:        f.Request.Method,
		URL:           f.Request.URL.String(),
		Duration:      doneAt,
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
