package schema

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyStringSlice = []string{}

func TestNewTrafficObject(t *testing.T) {
	t.Run("new traffic object", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
			"Delete-Me":    []string{"too-many-secrets"},
		}
		headersToFilter := []string{"Delete-Me"}

		trafficObject := NewTrafficObject(headers, "hello", headersToFilter)
		assert.Contains(t, trafficObject.Header, "Content-Type")
		assert.NotContains(t, trafficObject.Header, "Delete-Me")
		assert.Equal(t, "hello", trafficObject.Body)
	})
}

func TestNewFromProxyRequest(t *testing.T) {
	t.Run("new from proxy request", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
			"Delete-Me":    []string{"too-many-secrets"},
		}
		headersToFilter := []string{"Delete-Me"}

		url, err := url.Parse("http://example.com")
		require.NoError(t, err)

		request := &px.Request{
			Method: "GET",
			URL:    url,
			Header: headers,
			Body:   []byte("hello"),
			Proto:  "HTTP/1.1",
		}

		trafficObject := NewFromProxyRequest(request, headersToFilter)
		assert.Empty(t, 0, trafficObject.StatusCode, "status code not set on requests")
		assert.Equal(t, "GET", trafficObject.Method)
		assert.Equal(t, "http://example.com", trafficObject.URL.String())
		assert.Contains(t, trafficObject.Header, "Content-Type")
		assert.NotContains(t, trafficObject.Header, "Delete-Me")
		assert.Equal(t, "hello", trafficObject.Body)
		assert.Equal(t, "HTTP/1.1", trafficObject.Proto)
	})
	t.Run("new from proxy request with binary body", func(t *testing.T) {
		request := &px.Request{
			Body: []byte("\x01\x02\x03"),
		}
		trafficObject := NewFromProxyRequest(request, emptyStringSlice)
		assert.NotNil(t, trafficObject)
		assert.Empty(t, trafficObject.Body)
	})
	t.Run("nil request", func(t *testing.T) {
		trafficObject := NewFromProxyRequest(nil, emptyStringSlice)
		assert.Nil(t, trafficObject)
	})
}

func TestNewFromProxyResponse(t *testing.T) {
	t.Run("new from proxy response", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
			"Delete-Me":    []string{"too-many-secrets"},
		}
		headersToFilter := []string{"Delete-Me"}

		response := &px.Response{
			StatusCode: 200,
			Header:     headers,
			Body:       []byte("hello"),
		}

		trafficObject := NewFromProxyResponse(response, headersToFilter)
		assert.Equal(t, 200, trafficObject.StatusCode)
		assert.Equal(t, "hello", trafficObject.Body)
		assert.Equal(t, "", trafficObject.Proto)
		assert.Contains(t, trafficObject.Header, "Content-Type")
		assert.NotContains(t, trafficObject.Header, "Delete-Me")
	})
	t.Run("new from proxy response with binary body", func(t *testing.T) {
		response := &px.Response{
			Body: []byte("\x01\x02\x03"),
		}
		trafficObject := NewFromProxyResponse(response, emptyStringSlice)
		assert.NotNil(t, trafficObject)
		assert.Empty(t, trafficObject.Body)
	})
	t.Run("nil response", func(t *testing.T) {
		trafficObject := NewFromProxyResponse(nil, emptyStringSlice)
		assert.Nil(t, trafficObject)
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
			Header:          headers,
			headersToFilter: headersToFilter,
		}

		trafficObject.filterHeaders()
		assert.Contains(t, trafficObject.Header, "Content-Type")
		assert.NotContains(t, trafficObject.Header, "Delete-Me")
	})
}

func TestTrafficObject_MarshalJSON(t *testing.T) {
	t.Run("marshal json", func(t *testing.T) {
		trafficObject := &TrafficObject{
			Body: "hello",
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"Delete-Me":    []string{"too-many-secrets"},
			},
		}

		jsonBytes, err := json.Marshal(trafficObject)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonBytes)

		expected := `{"body": "hello", "header":{"Content-Type":["application/json"],"Delete-Me":["too-many-secrets"]}}`
		assert.JSONEq(t, expected, string(jsonBytes))
	})
}

func TestTrafficObject_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal json", func(t *testing.T) {
		data := []byte(`{
	"statusCode": 200,
	"method": "GET",
	"url": "http://example.com",
	"header": {
		"Content-Type": ["application/json"],
		"Delete-Me": ["too-many-secrets"]
	},
	"body": "hello"
}`)
		tObj, err := NewFromJSONBytes(data, []string{"Delete-Me"})
		require.NoError(t, err)
		assert.NotNil(t, tObj)
		assert.Equal(t, 200, tObj.StatusCode)
		assert.Equal(t, "GET", tObj.Method)
		assert.Equal(t, "http://example.com", tObj.URL.String())
		assert.Contains(t, tObj.Header, "Content-Type")
		assert.NotContains(t, tObj.Header, "Delete-Me")
		assert.Equal(t, "hello", tObj.Body)
	})

	t.Run("unmarshal empty json", func(t *testing.T) {
		data := []byte(`{}`)
		tObj, err := NewFromJSONBytes(data, emptyStringSlice)
		require.NoError(t, err)
		assert.NotNil(t, tObj)
	})

	t.Run("unmarshal empty json", func(t *testing.T) {
		data := []byte(`{ "header": {"Delete-Me": ["application/json"]} }`)
		tObj, err := NewFromJSONBytes(data, []string{"Delete-Me"})
		require.NoError(t, err)
		assert.NotNil(t, tObj)

	})
}

func TestTrafficObject_UnmarshalJSON_error(t *testing.T) {
	t.Run("missing brackets", func(t *testing.T) {
		data := []byte(`{"statusCode": 200, "method": "GET", "url": "http://example.com", "header": {"Content-Type": ["application/json"]`)
		tObj, err := NewFromJSONBytes(data, []string{"Delete-Me"})
		require.Error(t, err)
		assert.Nil(t, tObj)
	})

	t.Run("invalid status code", func(t *testing.T) {
		data := []byte(`{ "statusCode": "hello" }`)
		tObj, err := NewFromJSONBytes(data, []string{"Delete-Me"})
		require.Error(t, err)
		assert.Nil(t, tObj)
	})

	t.Run("invalid URL", func(t *testing.T) {
		data := []byte(`{ "url": 100 }`)
		tObj, err := NewFromJSONBytes(data, []string{"Delete-Me"})
		require.Error(t, err)
		assert.Nil(t, tObj)
	})

	t.Run("invalid headers", func(t *testing.T) {
		data := []byte(`{"header": {"Content-Type": "should be a list" }}`)
		tObj, err := NewFromJSONBytes(data, []string{})
		require.Error(t, err)
		assert.Nil(t, tObj)
	})

	t.Run("invalid body", func(t *testing.T) {
		data := []byte(`{"body": 100}`)
		tObj, err := NewFromJSONBytes(data, []string{})
		require.Error(t, err)
		assert.Nil(t, tObj)
	})

	t.Run("invalid proto", func(t *testing.T) {
		data := []byte(`{"proto": 100}`)
		tObj, err := NewFromJSONBytes(data, []string{})
		require.Error(t, err)
		assert.Nil(t, tObj)
	})
}
