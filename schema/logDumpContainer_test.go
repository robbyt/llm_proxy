package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"

	"github.com/robbyt/llm_proxy/config"
)

func getDefaultFlow() px.Flow {
	return px.Flow{
		Request: &px.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			},
			Proto: "HTTP/1.1",
			Header: http.Header{
				"Content-Type":      []string{"[application/json]"},
				"Delete-Me-Request": []string{"too-many-secrets"},
			},
			Body: []byte(`{"key": "value"}`),
		},
		Response: &px.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":       []string{"[application/json]"},
				"Delete-Me-Response": []string{"too-many-secrets"},
			},
			Body: []byte(`{"status": "success"}`),
		},
	}
}

func getDefaultConnectionStats() *ConnectionStatsContainer {
	return &ConnectionStatsContainer{
		ClientAddress: "unknown",
		URL:           "http://example.com/",
		Duration:      0,
		ProxyID:       "00000000-0000-0000-0000-000000000000",
	}
}

func TestNewLogDumpDiskContainer_JSON(t *testing.T) {
	testCases := []struct {
		name                    string
		flow                    px.Flow
		logSources              config.LogSourceConfig
		filterReqHeaders        []string
		filterRespHeaders       []string
		expectedConnectionStats *ConnectionStatsContainer
		expectedRequestMethod   string
		expectedRequestURL      string
		expectedRequestProto    string
		expectedRequestHeaders  string
		expectedRequestBody     string
		expectedResponseCode    int
		expectedResponseHeaders string
		expectedResponseBody    string
	}{
		{
			name: "all fields enabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        []string{},
			filterRespHeaders:       []string{},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\nDelete-Me-Request: too-many-secrets\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "Content-Type: [application/json]\r\nDelete-Me-Response: too-many-secrets\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: false,
				LogRequestHeaders:  false,
				LogRequest:         false,
				LogResponseHeaders: false,
				LogResponse:        false,
			},
			filterReqHeaders:        []string{},
			filterRespHeaders:       []string{},
			expectedConnectionStats: (*ConnectionStatsContainer)(nil), // weird way to assert nil
		},
		{
			name: "all fields enabled, with filter",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, request headers disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  false,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, request disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         false,
				LogResponseHeaders: true,
				LogResponse:        true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseCode:    http.StatusOK,
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response headers disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: false,
				LogResponse:        true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseCode:    http.StatusOK,
			expectedResponseHeaders: "",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequest:         true,
				LogResponseHeaders: true,
				LogResponse:        false,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestMethod:   "GET",
			expectedRequestURL:      "http://example.com/",
			expectedRequestProto:    "HTTP/1.1",
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			container := NewLogDumpContainer(tc.flow, tc.logSources, 0, tc.filterReqHeaders, tc.filterRespHeaders)
			assert.Equal(t, tc.expectedConnectionStats, container.ConnectionStats)
			assert.Equal(t, tc.expectedRequestMethod, container.Request.Method)
			assert.Equal(t, tc.expectedRequestURL, container.Request.URL.String())
			assert.Equal(t, tc.expectedRequestProto, container.Request.Proto)
			assert.Equal(t, tc.expectedRequestHeaders, container.Request.HeaderString())
			assert.Equal(t, tc.expectedRequestBody, container.Request.Body)
			assert.Equal(t, tc.expectedResponseCode, container.Response.StatusCode)
			assert.Equal(t, tc.expectedResponseHeaders, container.Response.HeaderString())
			assert.Equal(t, tc.expectedResponseBody, container.Response.Body)
		})
	}
}

func TestValidateFlowObj(t *testing.T) {
	t.Run("flow is nil", func(t *testing.T) {
		logSources := config.LogSourceConfig{
			LogRequestHeaders:  true,
			LogRequest:         true,
			LogResponseHeaders: true,
			LogResponse:        true,
		}
		// validateFlowObj(nil, &logSources)
		ls := validateFlowObj(nil, logSources)
		assert.False(t, ls.LogRequestHeaders)
		assert.False(t, ls.LogRequest)
		assert.False(t, ls.LogResponseHeaders)
		assert.False(t, ls.LogResponse)
	})

	t.Run("request is nil", func(t *testing.T) {
		logSources := config.LogSourceConfig{
			LogRequestHeaders: true,
			LogRequest:        true,
		}
		flow := &px.Flow{}
		ls := validateFlowObj(flow, logSources)
		assert.False(t, ls.LogRequestHeaders)
		assert.False(t, ls.LogRequest)
	})

	t.Run("response is nil", func(t *testing.T) {
		logSources := config.LogSourceConfig{
			LogResponseHeaders: true,
			LogResponse:        true,
		}
		flow := &px.Flow{}
		ls := validateFlowObj(flow, logSources)
		assert.False(t, ls.LogResponseHeaders)
		assert.False(t, ls.LogResponse)
	})

}
