// Modified from:
package addons

import (
	"encoding/json"
	"time"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

type logStdOutLine struct {
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

func (obj *logStdOutLine) toJSONstr() string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Failed to marshal object to JSON: %v", err)
	}
	return string(jsonData)
}

// StdOutLogger log connection and flow
type StdOutLogger struct {
	px.BaseAddon
}

func (addon *StdOutLogger) ClientConnected(client *px.ClientConn) {
	log.Debugf("client connect: %v", client.Conn.RemoteAddr())
}

func (addon *StdOutLogger) ClientDisconnected(client *px.ClientConn) {
	log.Debugf("client disconnect: %v", client.Conn.RemoteAddr())
}

func (addon *StdOutLogger) ServerConnected(connCtx *px.ConnContext) {
	log.Debugf(
		"server connect: %v (%v->%v)",
		connCtx.ServerConn.Address,
		connCtx.ServerConn.Conn.LocalAddr(),
		connCtx.ServerConn.Conn.RemoteAddr(),
	)
}

func (addon *StdOutLogger) ServerDisconnected(connCtx *px.ConnContext) {
	log.Debugf(
		"server disconnect: %v (%v->%v)",
		connCtx.ServerConn.Address,
		connCtx.ServerConn.Conn.LocalAddr(),
		connCtx.ServerConn.Conn.RemoteAddr(),
	)
}

func (addon *StdOutLogger) Requestheaders(f *px.Flow) {
	start := time.Now()
	go func() {
		<-f.Done()
		doneAt := time.Since(start).Milliseconds()
		logOutput := &logStdOutLine{
			ClientAddress: f.ConnContext.ClientConn.Conn.RemoteAddr().String(),
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

		log.Info(logOutput.toJSONstr())
	}()
}
