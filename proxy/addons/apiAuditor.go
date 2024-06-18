package addons

import (
	"fmt"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/proxati/llm_proxy/schema"
	log "github.com/sirupsen/logrus"
)

// APIAuditorAddon log connection and flow
type APIAuditorAddon struct {
	px.BaseAddon
	costCounter *schema.CostCounter
}

func (aud *APIAuditorAddon) Response(f *px.Flow) {
	if f.Response == nil {
		log.Debugf("skipping accounting for nil response: %s", f.Request.URL)
		return
	}
	go func() {
		<-f.Done()

		// Only account when receiving good response codes
		_, shouldAccount := cacheOnlyResponseCodes[f.Response.StatusCode]
		if !shouldAccount {
			log.Debugf("skipping accounting for non-200 response: %s", f.Request.URL)
			return
		}

		// convert the request to an internal TrafficObject
		tObjReq, err := schema.NewProxyRequestFromMITMRequest(f.Request, []string{})
		if err != nil {
			log.Errorf("error creating TrafficObject from request: %s", f.Request.URL)
			return
		}

		// convert the response to an internal TrafficObject
		tObjResp, err := schema.NewProxyResponseFromMITMResponse(f.Response, []string{})
		if err != nil {
			log.Errorf("error creating TrafficObject from response: %s", err)
			return
		}

		// account the cost
		err = aud.costCounter.Add(*tObjReq, *tObjResp)
		if err != nil {
			log.Errorf("error accounting response: %s", err)
		}
		log.Infof("Total cost for this session: %s", aud.costCounter.String())
	}()
}

func (addon *APIAuditorAddon) ServerConnected(connCtx *px.ConnContext) {
	go func() {
		log.InfoFn(func() []interface{} {
			return []interface{}{
				fmt.Sprintf("server connect: %v (%v->%v)",
					connCtx.ServerConn.Address,
					connCtx.ServerConn.Conn.LocalAddr(),
					connCtx.ServerConn.Conn.RemoteAddr(),
				),
			}
		})
	}()
}

func (addon *APIAuditorAddon) ServerDisconnected(connCtx *px.ConnContext) {
	go func() {
		log.InfoFn(func() []interface{} {
			return []interface{}{
				fmt.Sprintf("server disconnect: %v (%v->%v)",
					connCtx.ServerConn.Address,
					connCtx.ServerConn.Conn.LocalAddr(),
					connCtx.ServerConn.Conn.RemoteAddr(),
				),
			}
		})
	}()
}

func NewAPIAuditor() *APIAuditorAddon {
	return &APIAuditorAddon{
		costCounter: schema.NewCostCounter(),
	}
}
