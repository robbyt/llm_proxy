// Modified from:
package addons

import (
	"fmt"
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
	log.DebugFn(func() []interface{} {
		return []interface{}{fmt.Sprintf("client connect: %v", client.Conn.RemoteAddr())}
	})
}

func (addon *StdOutLogger) ClientDisconnected(client *px.ClientConn) {
	log.DebugFn(func() []interface{} {
		return []interface{}{fmt.Sprintf("client disconnect: %v", client.Conn.RemoteAddr())}
	})
}

func (addon *StdOutLogger) ServerConnected(connCtx *px.ConnContext) {
	log.DebugFn(func() []interface{} {
		return []interface{}{
			fmt.Sprintf("server connect: %v (%v->%v)",
				connCtx.ServerConn.Address,
				connCtx.ServerConn.Conn.LocalAddr(),
				connCtx.ServerConn.Conn.RemoteAddr(),
			),
		}
	})
}

func (addon *StdOutLogger) ServerDisconnected(connCtx *px.ConnContext) {
	log.DebugFn(func() []interface{} {
		return []interface{}{
			fmt.Sprintf("server disconnect: %v (%v->%v)",
				connCtx.ServerConn.Address,
				connCtx.ServerConn.Conn.LocalAddr(),
				connCtx.ServerConn.Conn.RemoteAddr(),
			),
		}
	})
}

func (addon *StdOutLogger) Requestheaders(f *px.Flow) {
	start := time.Now()
	go func() {
		<-f.Done()
		// InfoFn will only render if logging is verbose enough
		log.InfoFn(func() []interface{} {
			doneAt := time.Since(start).Milliseconds()
			logOutput := schema.NewConnectionStatusContainerWithDuration(*f, doneAt)
			if logOutput == nil {
				return nil
			}

			return []interface{}{logOutput.ToJSONstr()}
		})
	}()
}

func NewStdOutLogger() *StdOutLogger {
	return &StdOutLogger{}
}
