package megadumper

import (
	"testing"

	"github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestNewLogDumpDiskContainer_JSON(t *testing.T) {
	flow := &proxy.Flow{
		Request: &proxy.Request{
			Header: map[string][]string{
				"Content-Type": {"[application/json]"},
			},
			Body: []byte(`{"key": "value"}`),
		},
		Response: &proxy.Response{
			Header: map[string][]string{
				"Content-Type": {"[application/json]"},
			},
			Body: []byte(`{"status": "success"}`),
		},
	}
	var container *LogDumpContainer

	container = NewLogDumpContainer(flow, LogSourceConfig{LogRequestHeaders: true})
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container = NewLogDumpContainer(flow, LogSourceConfig{LogRequestBody: true})
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, `{"key": "value"}`, container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container = NewLogDumpContainer(flow, LogSourceConfig{LogResponseHeaders: true})
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container = NewLogDumpContainer(flow, LogSourceConfig{LogResponseBody: true})
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)

	container = NewLogDumpContainer(flow, LogSourceConfig{
		LogRequestHeaders:  true,
		LogRequestBody:     true,
		LogResponseHeaders: true,
		LogResponseBody:    true,
	})
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, `{"key": "value"}`, container.RequestBody)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)

}
