package formatters

import (
	"testing"

	"github.com/robbyt/llm_proxy/schema"
	"github.com/stretchr/testify/assert"
)

func TestPlainText_flatten(t *testing.T) {
	container := &schema.LogDumpContainer{
		RequestHeaders:  "Request Headers",
		RequestBody:     "Request Body",
		ResponseHeaders: "Response Headers",
		ResponseBody:    "Response Body",
	}
	pt := &PlainText{container}

	expectedResult := []byte("Request Headers\r\nRequest Body\r\nResponse Headers\r\nResponse Body\r\n")

	result, err := pt.flatten()
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestPlainText_flatten_EmptyFields(t *testing.T) {
	pt := &PlainText{}

	expectedResult := []byte("")

	result, err := pt.flatten()
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
