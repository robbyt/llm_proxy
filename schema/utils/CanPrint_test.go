package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanPrintString(t *testing.T) {
	assert.True(t, CanPrintString("Hello, world!"))
	assert.False(t, CanPrintString("\x01\x02\x03"))
}

func TestCanPrint(t *testing.T) {
	assert.True(t, CanPrint([]byte("Hello, world!")))
	assert.False(t, CanPrint([]byte("\x01\x02\x03")))
}
