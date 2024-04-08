package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderString(t *testing.T) {
	t.Run("valid headers", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
			"User-Agent":   []string{"test-agent"},
		}
		expected := "Content-Type: application/json\r\nUser-Agent: test-agent\r\n"
		assert.Equal(t, expected, HeaderString(headers))
	})

	t.Run("empty headers", func(t *testing.T) {
		headers := http.Header{}
		expected := ""
		assert.Equal(t, expected, HeaderString(headers))
	})
}
