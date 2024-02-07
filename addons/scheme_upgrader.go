package addons

import (
	px "github.com/kardianos/mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// var titleRegexp = regexp.MustCompile("(<title>)(.*?)(</title>)")

type SchemeUpgrader struct {
	px.BaseAddon
	upgraded bool
}

func (c *SchemeUpgrader) Request(f *px.Flow) {
	// upgrade to https
	if f.Request.URL.Scheme == "https" {
		log.Debugf("Upgrading URL scheme from http to https not needed for URL: %s", f.Request.URL)
		c.upgraded = false
		return
	}

	// upgrade the connection from http to https, so when sent upstream it will be encrypted
	f.Request.URL.Scheme = "https"
	c.upgraded = true
}

func (c *SchemeUpgrader) Response(f *px.Flow) {
	if !c.upgraded {
		return
	}
	if f.Response == nil || f.Response.Header == nil {
		return
	}
	f.Response.Header.Add("X-Llm_proxy-scheme-upgraded", "true")
}
