package addons

import (
	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/proxati/llm_proxy/schema"
	log "github.com/sirupsen/logrus"
)

// domains supported by this auditor
var auditURLs = map[string]interface{}{
	"api.openai.com": nil,
}

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

		// only account when the request domain is supported
		reqHostname := f.Request.URL.Hostname()
		_, shouldAudit := auditURLs[reqHostname]
		if !shouldAudit {
			log.Debugf("skipping accounting for unsupported API: %s", reqHostname)
			return
		}

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
	}()
}

func NewAPIAuditor() *APIAuditorAddon {
	return &APIAuditorAddon{
		costCounter: schema.NewCostCounter(),
	}
}
