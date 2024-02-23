// Modified from:
package addons

import (
	"time"

	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"

	"github.com/robbyt/llm_proxy/schema"
)

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
		logOutput := schema.NewConnectionStatusContainer(*f, doneAt)
		if logOutput != nil {
			log.Info(logOutput.ToJSONstr())
		}
	}()
}

func NewStdOutLogger() *StdOutLogger {
	return &StdOutLogger{}
}
