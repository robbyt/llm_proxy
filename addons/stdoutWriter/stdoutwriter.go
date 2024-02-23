package stdoutWriter

import (
	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/robbyt/llm_proxy/schema"
	log "github.com/sirupsen/logrus"
)

const UnknownAddr = "unknown"

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

func NewLogLine(f *px.Flow, doneAt int64) *schema.ConnectionStatsLogger {
	if f == nil {
		log.Error("Flow object is nil")
		return nil
	}

	logOutput := &schema.ConnectionStatsLogger{
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
