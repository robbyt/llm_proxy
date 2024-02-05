package megadumper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogDumpDiskContainer_DumpToJSONBytes(t *testing.T) {
	container := &LogDumpDiskContainer{
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

	jsonBytes, err := container.DumpToJSONBytes()
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedParsedJSON, parsedJSON)
}
