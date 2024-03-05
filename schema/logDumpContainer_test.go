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
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			},
			Header: http.Header{
				"Content-Type":      []string{"[application/json]"},
				"Delete-Me-Request": []string{"too-many-secrets"},
			},
			Body: []byte(`{"key": "value"}`),
		},
		Response: &px.Response{
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
		ClientAddress:       "unknown",
		Method:              "",
		URL:                 "http://example.com/",
		ResponseCode:        0,
		ContentLength:       21,
		Duration:            0,
		ResponseContentType: "[application/json]",
		ProxyID:             "00000000-0000-0000-0000-000000000000",
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
		expectedRequestHeaders  string
		expectedRequestBody     string
		expectedResponseHeaders string
		expectedResponseBody    string
	}{
		{
			name: "all fields enabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequestBody:     true,
				LogResponseHeaders: true,
				LogResponseBody:    true,
			},
			filterReqHeaders:        []string{},
			filterRespHeaders:       []string{},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "Content-Type: [application/json]\r\nDelete-Me-Request: too-many-secrets\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseHeaders: "Content-Type: [application/json]\r\nDelete-Me-Response: too-many-secrets\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: false,
				LogRequestHeaders:  false,
				LogRequestBody:     false,
				LogResponseHeaders: false,
				LogResponseBody:    false,
			},
			filterReqHeaders:        []string{},
			filterRespHeaders:       []string{},
			expectedConnectionStats: (*ConnectionStatsContainer)(nil), // weird way to assert nil
			expectedRequestHeaders:  "",
			expectedRequestBody:     "",
			expectedResponseHeaders: "",
			expectedResponseBody:    "",
		},
		{
			name: "all fields enabled, with filter",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequestBody:     true,
				LogResponseHeaders: true,
				LogResponseBody:    true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, request headers disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  false,
				LogRequestBody:     true,
				LogResponseHeaders: true,
				LogResponseBody:    true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, request body disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequestBody:     false,
				LogResponseHeaders: true,
				LogResponseBody:    true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     "",
			expectedResponseHeaders: "Content-Type: [application/json]\r\n",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response headers disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequestBody:     true,
				LogResponseHeaders: false,
				LogResponseBody:    true,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
			expectedRequestHeaders:  "Content-Type: [application/json]\r\n",
			expectedRequestBody:     `{"key": "value"}`,
			expectedResponseHeaders: "",
			expectedResponseBody:    `{"status": "success"}`,
		},
		{
			name: "all fields enabled, with filter, response body disabled",
			flow: getDefaultFlow(),
			logSources: config.LogSourceConfig{
				LogConnectionStats: true,
				LogRequestHeaders:  true,
				LogRequestBody:     true,
				LogResponseHeaders: true,
				LogResponseBody:    false,
			},
			filterReqHeaders:        []string{"Delete-Me-Request"},
			filterRespHeaders:       []string{"Delete-Me-Response"},
			expectedConnectionStats: getDefaultConnectionStats(),
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
			assert.Equal(t, tc.expectedRequestHeaders, container.Request.HeadersString())
			assert.Equal(t, tc.expectedRequestBody, container.Request.Body)
			assert.Equal(t, tc.expectedResponseHeaders, container.Response.HeadersString())
			assert.Equal(t, tc.expectedResponseBody, container.Response.Body)
		})
	}
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

func TestTrafficObject_filterHeaders(t *testing.T) {
	t.Run("filter headers", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
			"Delete-Me":    []string{"too-many-secrets"},
		}
		headersToFilter := []string{"Delete-Me"}

		trafficObject := &TrafficObject{
			Headers:         headers,
			headersToFilter: headersToFilter,
		}

		trafficObject.filterHeaders()
		assert.Contains(t, trafficObject.Headers, "Content-Type")
		assert.NotContains(t, trafficObject.Headers, "Delete-Me")
	})
}
