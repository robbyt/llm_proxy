package addons

import (
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestSchemeUpgrader_Request(t *testing.T) {
	upgrader := &SchemeUpgrader{}
	req := &px.Request{
		URL: &url.URL{
			Scheme: "http",
		},
		Method: "GET",
	}
	flow := &px.Flow{
		Request: req,
	}

	upgrader.Request(flow)

	assert.Equal(t, "https", flow.Request.URL.Scheme)
	assert.True(t, upgrader.upgraded)
}
