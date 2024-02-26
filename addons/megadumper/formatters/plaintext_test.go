package formatters

import (
	"net/http"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/robbyt/llm_proxy/schema"
	"github.com/stretchr/testify/assert"
)

func TestPlainText_flatten(t *testing.T) {
	container := &schema.LogDumpContainer{
		RequestHeaders:  http.Header{"ReqHeader": []string{"ReqValue"}},
		RequestBody:     "Request Body",
		ResponseHeaders: http.Header{"RespHeader": []string{"RespValue"}},
		ResponseBody:    "Response Body",
	}
	pt := &PlainText{container}

	expectedResult := []byte("ReqHeader: ReqValue\r\nRequest Body\r\nRespHeader: RespValue\r\nResponse Body\r\n")

	result, err := pt.flatten()
	assert.NoError(t, err)
	spew.Dump(result)
	assert.Equal(t, expectedResult, result)
}

func TestPlainText_flatten_EmptyFields(t *testing.T) {
	pt := &PlainText{}

	expectedResult := []byte("")

	result, err := pt.flatten()
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
