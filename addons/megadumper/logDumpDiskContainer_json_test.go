package megadumper

import (
	"encoding/json"
	"testing"

	"github.com/kardianos/mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

func TestLogDumpDiskContainer_DumpToJSONBytes(t *testing.T) {
	container := &LogDumpDiskContainer_JSON{
		RequestHeaders:  "Request Headers",
		RequestBody:     "Request Body",
		ResponseHeaders: "Response Headers",
		ResponseBody:    "Response Body",
	}

	expectedJSON := `{
	  "request_headers": "Request Headers",
	  "request_body": "Request Body",
	  "response_headers": "Response Headers",
	  "response_body": "Response Body"
	}`

	jsonBytes, err := container.Read()
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedParsedJSON, parsedJSON)
}

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

	container, err := newLogDumpDiskContainer_JSON(flow, WRITE_REQ_HEADERS_ONLY)
	assert.NoError(t, err)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, "", container.ResponseBody)

	container, err = newLogDumpDiskContainer_JSON(flow, WRITE_REQ_BODY_AND_RESP_BODY)
	assert.NoError(t, err)
	assert.Equal(t, "", container.RequestHeaders)
	assert.Equal(t, `{"key": "value"}`, container.RequestBody)
	assert.Equal(t, "", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)

	container, err = newLogDumpDiskContainer_JSON(flow, WRITE_REQ_HEADERS_ALSO_RESP_HEADERS_AND_RESP_BODY)
	assert.NoError(t, err)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, "", container.RequestBody)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)

	container, err = newLogDumpDiskContainer_JSON(flow, WRITE_REQ_HEADERS_AND_REQ_BODY_ALSO_RESP_HEADERS_AND_RESP_BODY)
	assert.NoError(t, err)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.RequestHeaders)
	assert.Equal(t, `{"key": "value"}`, container.RequestBody)
	assert.Equal(t, "Content-Type: [application/json]\r\n", container.ResponseHeaders)
	assert.Equal(t, `{"status": "success"}`, container.ResponseBody)
}
