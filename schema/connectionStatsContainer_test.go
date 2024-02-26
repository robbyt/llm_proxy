package schema

import (
	"net/http"
	"net/url"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestLogStdOutLine_toJSONstr(t *testing.T) {
	line := &ConnectionStatsContainer{
		ClientAddress:       "127.0.0.1",
		Method:              "GET",
		URL:                 "http://example.com",
		StatusCode:          200,
		ContentLength:       13,
		Duration:            100,
		ResponseContentType: "text/plain",
	}

	expected := `{"client_address":"127.0.0.1","method":"GET","url":"http://example.com","status_code":200,"content_length":13,"duration_ms":100,"response_content_type":"text/plain"}`
	assert.Equal(t, expected, line.ToJSONstr())
}

func TestNewLogLine(t *testing.T) {
	// Create a mock Flow object
	f := px.Flow{
		Request: &px.Request{
			Method: "GET",
			URL:    &url.URL{Scheme: "https", Host: "example.com", Path: "/testpath"},
		},
		Response: &px.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}, "X-Request-Id": []string{"1234"}},
			Body:       nil, // You can also mock a Body here
		},
		Id: uuid.NewV4(),
	}

	logLine := NewConnectionStatusContainerWithDuration(f, 100)
	assert.NotNil(t, logLine)
	assert.Equal(t, "unknown", logLine.ClientAddress)
	assert.Equal(t, "GET", logLine.Method)
	assert.Equal(t, "https://example.com/testpath", logLine.URL)
	assert.Equal(t, 200, logLine.StatusCode)
	assert.Equal(t, "application/json", logLine.ResponseContentType)
	assert.Equal(t, f.Id.String(), logLine.ProxyID)
	assert.Equal(t, int64(100), logLine.Duration)
}
