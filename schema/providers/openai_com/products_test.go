package openai_com

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEmbeddedDataJSON(t *testing.T) {
	assert.NotEmpty(t, API_Endpoint_Data, "init() populates this variable")

	// Reset API_Endpoint_Pricing before test
	API_Endpoint_Data = nil

	err := loadEmbeddedDataJSON()
	assert.Nil(t, err, "Expected no error loading data.json, but got an error")

	assert.NotEmpty(t, API_Endpoint_Data, "Expected API_Endpoint_Pricing to be populated, but it was empty")
}
