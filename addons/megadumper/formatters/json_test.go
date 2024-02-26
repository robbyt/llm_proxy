package formatters

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/robbyt/llm_proxy/schema"
	"github.com/stretchr/testify/assert"
)

func TestJSONFormatter(t *testing.T) {
	container := &schema.LogDumpContainer{
		RequestHeaders:  http.Header{"Header": []string{"Value"}},
		RequestBody:     "Request Body",
		ResponseHeaders: http.Header{"Header": []string{"Value"}},
		ResponseBody:    "Response Body",
	}
	j := &JSON{}

	expectedJSON := `{
	  "timestamp": "0001-01-01T00:00:00Z",
	  "request_headers": {
	    "Header": [
		  "Value"
		]
	  },
	  "request_body": "Request Body",
	  "response_headers": {
		"Header": [
		  "Value"
	    ]
	  },
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
	container := &schema.LogDumpContainer{
		RequestHeaders:  http.Header{},
		ResponseHeaders: http.Header{},
	}
	j := &JSON{}
	expectedJSON := `{
	  "timestamp": "0001-01-01T00:00:00Z",
	  "request_headers": {},
	  "request_body": "",
	  "response_headers": {},
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

	keys := []string{"timestamp", "request_headers", "request_body", "response_headers", "response_body"}
	for _, key := range keys {
		parsedValue, ok := parsedJSON[key]
		if ok {
			expectedValue, ok := expectedParsedJSON[key]
			if ok {
				assert.Equal(t, expectedValue, parsedValue)
			} else {
				t.Errorf("Expected to find %s in expectedParsedJSON", key)
			}
		} else {
			t.Errorf("Expected to find %s in parsedJSON", key)
		}
	}

}

func TestJSONFormatter_implements_Reader(t *testing.T) {
	var _ MegaDumpFormatter = &JSON{}
}
