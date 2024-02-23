package formatters

import (
	"encoding/json"
	"testing"

	"github.com/robbyt/llm_proxy/schema"
	"github.com/stretchr/testify/assert"
)

func TestJSONFormatter(t *testing.T) {
	container := &schema.LogDumpContainer{
		RequestHeaders:  "Request Headers",
		RequestBody:     "Request Body",
		ResponseHeaders: "Response Headers",
		ResponseBody:    "Response Body",
	}
	j := &JSON{}

	expectedJSON := `{
	  "request_headers": "Request Headers",
	  "request_body": "Request Body",
	  "response_headers": "Response Headers",
	  "response_body": "Response Body"
	}`

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedParsedJSON, parsedJSON)
}

func TestJSONFormatter_Empty(t *testing.T) {
	container := &schema.LogDumpContainer{}
	j := &JSON{}
	expectedJSON := `{
	  "request_headers": "",
	  "request_body": "",
	  "response_headers": "",
	  "response_body": ""
	}`

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	assert.Equal(t, expectedParsedJSON, parsedJSON)
}

func TestJSONFormatter_implements_Reader(t *testing.T) {
	var _ MegaDumpFormatter = &JSON{}
}
