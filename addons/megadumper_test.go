package addons

import (
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"

	md "github.com/robbyt/llm_proxy/addons/megadumper"
	"github.com/robbyt/llm_proxy/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMegaDirDumper_JSON_LogDir(t *testing.T) {
	logTarget := "/tmp/logs"
	logFormat := md.Format_JSON
	logSources := config.LogSourceConfig{}
	logDestinations := []md.LogDestination{md.WriteToDir}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaDirDumper(logTarget, logFormat, logSources, logDestinations, filterReqHeaders, filterRespHeaders)

	assert.NoError(t, err)
	assert.NotNil(t, mda)
	assert.Equal(t, logSources, mda.logSources)
	assert.Len(t, mda.writers, 1)
}

func TestNewMegaDirDumper_TXT_LOGFILE(t *testing.T) {
	logTarget := "/tmp/logs"
	logFormat := md.Format_PLAINTEXT
	logSources := config.LogSourceConfig{}
	logDestinations := []md.LogDestination{md.WriteToFile}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaDirDumper(logTarget, logFormat, logSources, logDestinations, filterReqHeaders, filterRespHeaders)

	assert.NoError(t, err)
	assert.NotNil(t, mda)
	assert.Equal(t, logSources, mda.logSources)
	assert.Len(t, mda.writers, 1)
}

func TestRequestheadersMethod(t *testing.T) {
	logTarget := "/tmp/logs"
	logFormat := md.Format_JSON
	logSources := config.LogSourceConfig{}
	logDestinations := []md.LogDestination{md.WriteToDir}
	filterReqHeaders := []string{}
	filterRespHeaders := []string{}

	mda, err := NewMegaDirDumper(logTarget, logFormat, logSources, logDestinations, filterReqHeaders, filterRespHeaders)

	assert.NoError(t, err)
	assert.NotNil(t, mda)

	flow := &px.Flow{
		Request: &px.Request{
			Header: map[string][]string{
				"Content-Type": {"[application/json]"},
			},
			Body: []byte(`{"key": "value"}`),
		},
		Response: &px.Response{
			Header: map[string][]string{
				"Content-Type": {"[application/json]"},
			},
			Body: []byte(`{"status": "success"}`),
		},
	}
	mda.Requestheaders(flow)
}
