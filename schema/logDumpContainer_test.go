package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"

	"github.com/robbyt/llm_proxy/config"
)

func TestNewLogDumpDiskContainer_JSON(t *testing.T) {
	flow := &px.Flow{
		Request: &px.Request{
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			},
			Header: http.Header{
				"Content-Type": []string{"[application/json]"},
			},
			Body: []byte(`{"key": "value"}`),
		},
		Response: &px.Response{
			Header: http.Header{
				"Content-Type": []string{"[application/json]"},
			},
			Body: []byte(`{"status": "success"}`),
		},
	}
	var container *LogDumpContainer

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogRequestHeaders: true}, 0, []string{}, []string{})
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.Request.HeadersString())
	assert.Equal(t, "", container.Request.Body)
	assert.Equal(t, "", container.Response.HeadersString())
	assert.Equal(t, "", container.Response.Body)

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogRequestBody: true}, 0, []string{}, []string{})
	assert.Equal(t, "", container.Request.HeadersString())
	assert.Equal(t, `{"key": "value"}`, container.Request.Body)
	assert.Equal(t, "", container.Response.HeadersString())
	assert.Equal(t, "", container.Response.Body)

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogResponseHeaders: true}, 0, []string{}, []string{})
	assert.Equal(t, "", container.Request.HeadersString())
	assert.Equal(t, "", container.Request.Body)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.Response.HeadersString())
	assert.Equal(t, "", container.Response.Body)

	container = NewLogDumpContainer(*flow, config.LogSourceConfig{LogResponseBody: true}, 0, []string{}, []string{})
	assert.Equal(t, "", container.Request.HeadersString())
	assert.Equal(t, "", container.Request.Body)
	assert.Equal(t, "", container.Response.HeadersString())
	assert.Equal(t, `{"status": "success"}`, container.Response.Body)

	container = NewLogDumpContainer(
		*flow,
		config.LogSourceConfig{
			LogRequestHeaders:  true,
			LogRequestBody:     true,
			LogResponseHeaders: true,
			LogResponseBody:    true,
		},
		0,
		[]string{},
		[]string{},
	)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.Request.HeadersString())
	assert.Equal(t, `{"key": "value"}`, container.Request.Body)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.Response.HeadersString())
	assert.Equal(t, `{"status": "success"}`, container.Response.Body)

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
