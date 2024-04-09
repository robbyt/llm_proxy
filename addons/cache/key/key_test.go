package key

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewKey(t *testing.T) {
	key := []byte("test")
	result := NewKey(key)
	require.NotNil(t, result, "NewKey should return a non-nil Key object")

	expectedHash := NewBLAKE2Key([]byte(key)).Get()
	require.Equal(t, expectedHash, result.Get())
}

func TestNewKeyStr(t *testing.T) {
	key := "test"
	result := NewKeyStr(key)
	require.NotNil(t, result, "NewKeyStr should return a non-nil Key object")

	expectedHash := NewBLAKE2Key([]byte(key)).Get()
	require.Equal(t, expectedHash, result.Get())
}
