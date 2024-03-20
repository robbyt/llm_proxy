package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var loremBytes = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod
tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in
reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint
occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
`)

var loremStr = string(loremBytes)

func TestCanPrint(t *testing.T) {
	assert.True(t, CanPrint(loremBytes))
	assert.False(t, CanPrint([]byte("\x01\x02\x03")))
}

func BenchmarkCanPrint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CanPrint(loremBytes)
		CanPrint([]byte("\x01\x02\x03"))
		_ = string(loremBytes) // simulate conversion to string, because Fast returns the converted string
	}
}

func TestCanPrintFast(t *testing.T) {
	str, isStr := CanPrintFast(loremBytes)
	require.True(t, isStr)
	assert.Equal(t, loremStr, str)

	str, isStr = CanPrintFast([]byte("\x01\x02\x03"))
	require.False(t, isStr)
	assert.Equal(t, "", str)
}

func BenchmarkCanPrintFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CanPrintFast(loremBytes)
		CanPrintFast([]byte("\x01\x02\x03"))
	}
}
