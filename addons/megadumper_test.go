package addons

import (
	"net/url"
	"strings"
	"testing"

	px "github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestMegaDumper_prepDumpBytes_Header(t *testing.T) {
	dumper := &MegaDumper{logLevel: LogLevel(WRITE_REQ_HEADERS_ONLY)}
	req := &px.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: "example.com"},
		Body:   []byte("Hello!"),
	}
	res := &px.Response{
		StatusCode: 200,
		Body:       []byte("world!"),
		Header:     map[string][]string{"Content-Type": {"text/plain"}},
	}
	flow := &px.Flow{
		Request:  req,
		Response: res,
	}

	buf, err := dumper.prepDumpBytes(flow)

	assert.NoError(t, err)
	resp := strings.ReplaceAll(buf.String(), "\r\n", "")
	assert.Contains(t, resp, "GET /")
	assert.Contains(t, resp, "Host: example.com")

	// not writing response headers and body
	assert.NotContains(t, resp, "world!")
	assert.NotContains(t, resp, "200 OK")
}

func TestMegaDumper_prepDumpBytes_HeaderAndBody(t *testing.T) {
	dumper := &MegaDumper{logLevel: LogLevel(WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY)}
	req := &px.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: "example.com"},
		Body:   []byte("Hello!"),
	}
	res := &px.Response{
		StatusCode: 200,
		Body:       []byte("world!"),
		Header:     map[string][]string{"Content-Type": {"text/plain"}},
	}
	flow := &px.Flow{
		Request:  req,
		Response: res,
	}

	buf, err := dumper.prepDumpBytes(flow)

	assert.NoError(t, err)
	resp := strings.ReplaceAll(buf.String(), "\r\n", "")
	assert.Contains(t, resp, "GET /")
	assert.Contains(t, resp, "Host: example.com")
	assert.Contains(t, resp, "world!")
	assert.Contains(t, resp, "200 OK")
}

// TestMegaDumper_prepDumpBytes_HeaderAndBody_WrongContentType tests the case when the response has
// an incompatible content type, it should not write the response body
func TestMegaDumper_prepDumpBytes_HeaderAndBody_WrongContentType(t *testing.T) {
	dumper := &MegaDumper{logLevel: LogLevel(WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY)}
	req := &px.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: "example.com"},
		Body:   []byte("Hello!"),
	}
	res := &px.Response{
		StatusCode: 200,
		Body:       []byte("world!"),
		Header:     map[string][]string{"Content-Type": {"bin/hex"}},
	}
	flow := &px.Flow{
		Request:  req,
		Response: res,
	}

	buf, err := dumper.prepDumpBytes(flow)

	assert.NoError(t, err)
	resp := strings.ReplaceAll(buf.String(), "\r\n", "")
	assert.Contains(t, resp, "GET /")
	assert.Contains(t, resp, "Host: example.com")
	assert.NotContains(t, resp, "world!")
	assert.Contains(t, resp, "200 OK")
}

func TestCanPrint(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "Printable content",
			content:  []byte("Hello, World!"),
			expected: true,
		},
		{
			name:     "Non-printable content",
			content:  []byte("Hello, \x00World!"),
			expected: false,
		},
		{
			name:     "Whitespace content",
			content:  []byte("Hello, \tWorld!"),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := canPrint(test.content)
			assert.Equal(t, test.expected, result)
		})
	}
}
