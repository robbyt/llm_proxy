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
		Request: &schema.TrafficObject{
			Headers: http.Header{"ReqHeader": []string{"ReqValue"}},
			Body:    "Request Body",
		},
		Response: &schema.TrafficObject{
			Headers: http.Header{"RespHeader": []string{"RespValue"}},
			Body:    "Response Body",
		},
	}
	j := &JSON{}

	expectedJSON := `{
	  "timestamp": "0001-01-01T00:00:00Z",
	  "request": {
		"body": "Request Body",
		"headers": { "ReqHeader": [ "ReqValue" ] }
	  },
	  "response": {
		"body": "Response Body",
		"headers": { "RespHeader": [ "RespValue" ] }
	  }
	}`

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	keys := []string{"timestamp", "request", "response"}
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

func TestJSONFormatter_Empty(t *testing.T) {
	container := &schema.LogDumpContainer{
		Request:  &schema.TrafficObject{},
		Response: &schema.TrafficObject{},
	}
	j := &JSON{}
	expectedJSON := `{
		"timestamp": "0001-01-01T00:00:00Z",
		"request": {
			"body": "",
			"headers": null
		},
		"response": {
			"body": "",
			"headers": null
		}
	  }`

	jsonBytes, err := j.Read(container)
	assert.NoError(t, err)

	var parsedJSON map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsedJSON)
	assert.NoError(t, err)

	expectedParsedJSON := make(map[string]interface{})
	err = json.Unmarshal([]byte(expectedJSON), &expectedParsedJSON)
	assert.NoError(t, err)

	keys := []string{"timestamp", "request", "response"}
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
