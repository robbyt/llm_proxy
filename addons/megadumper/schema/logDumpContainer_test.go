package schema

import (
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/robbyt/llm_proxy/config"
	"github.com/stretchr/testify/assert"
)

func TestNewLogDumpDiskContainer_JSON(t *testing.T) {
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
	var container *LogDumpContainer

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogRequestHeaders: true}, []string{}, []string{})
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogRequestBody: true}, []string{}, []string{})
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, `{"key": "value"}`, container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogResponseHeaders: true}, []string{}, []string{})
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogResponseBody: true}, []string{}, []string{})
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)

	container = NewLogDumpContainer(
		*flow,
		config.LogSourceConfig{
			LogRequestHeaders:  true,
			LogRequestBody:     true,
			LogResponseHeaders: true,
			LogResponseBody:    true,
		},
		[]string{},
		[]string{},
	)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, `{"key": "value"}`, container.RequestBody)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)

}

func TestValidateFlowObj(t *testing.T) {
	t.Run("flow is nil", func(t *testing.T) {
		logSources := config.LogSourceConfig{
			LogRequestHeaders:  true,
			LogRequestBody:     true,
			LogResponseHeaders: true,
			LogResponseBody:    true,
		}
		// validateFlowObj(nil, &logSources)
		ls := validateFlowObj(nil, logSources)
		assert.False(t, ls.LogRequestHeaders)
		assert.False(t, ls.LogRequestBody)
		assert.False(t, ls.LogResponseHeaders)
		assert.False(t, ls.LogResponseBody)
	})

	t.Run("request is nil", func(t *testing.T) {
		logSources := config.LogSourceConfig{
			LogRequestHeaders: true,
			LogRequestBody:    true,
		}
		flow := &px.Flow{}
		ls := validateFlowObj(flow, logSources)
		assert.False(t, ls.LogRequestHeaders)
		assert.False(t, ls.LogRequestBody)
	})

	t.Run("response is nil", func(t *testing.T) {
		logSources := config.LogSourceConfig{
			LogResponseHeaders: true,
			LogResponseBody:    true,
		}
		flow := &px.Flow{}
		ls := validateFlowObj(flow, logSources)
		assert.False(t, ls.LogResponseHeaders)
		assert.False(t, ls.LogResponseBody)
	})

}
