package megadumper

import (
	"net/http"
	"testing"

	"github.com/kardianos/mitmproxy/proxy"
	px "github.com/kardianos/mitmproxy/proxy"

	"github.com/stretchr/testify/assert"
)

func TestNewLogDumpDiskContainer_bytes(t *testing.T) {
	t.Run("should create LogDumpDiskContainer_bytes correctly", func(t *testing.T) {
		// Setup
		f := &proxy.Flow{
			Request: &px.Request{
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: []byte("Request Body"),
			},
			Response: &px.Response{
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: []byte("Response Body"),
			},
		}
		logLevel := WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY

		// Execute
		container, err := newLogDumpDiskContainer_bytes(f, logLevel)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, container)
		assert.Equal(t, logLevel, container.logLevel)
		assert.Contains(t, string(container.RawLogBytes), "Request Body")
		assert.Contains(t, string(container.RawLogBytes), "Response Body")
	})
}
