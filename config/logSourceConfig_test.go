package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogSourceConfig_String(t *testing.T) {
	config := LogSourceConfig{
		LogConnectionStats: true,
		LogRequestHeaders:  true,
		LogRequestBody:     true,
		LogResponseHeaders: true,
		LogResponseBody:    true,
	}

	expected := `{"LogConnectionStats":true,"LogRequestHeaders":true,"LogRequestBody":true,"LogResponseHeaders":true,"LogResponseBody":true}`
	assert.Equal(t, expected, config.String())
}
