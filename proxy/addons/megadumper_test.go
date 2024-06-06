package addons

import (
	"testing"

	md "github.com/proxati/llm_proxy/proxy/addons/megadumper"
	"github.com/proxati/llm_proxy/config"
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
