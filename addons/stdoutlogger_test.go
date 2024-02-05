package addons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogStdOutLine_toJSONstr(t *testing.T) {
	line := &logStdOutLine{
		ClientAddress: "127.0.0.1",
		Method:        "GET",
		URL:           "http://example.com",
		StatusCode:    200,
		ContentLength: 13,
		Duration:      100,
		ContentType:   "text/plain",
		XreqID:        "123",
	}

	expected := `{"client_address":"127.0.0.1","method":"GET","url":"http://example.com","status_code":200,"content_length":13,"duration_ms":100,"content_type":"text/plain","x_request_id":"123"}`
	assert.Equal(t, expected, line.toJSONstr())
}
