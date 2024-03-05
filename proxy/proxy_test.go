package proxy

import (
	"testing"

	"github.com/robbyt/llm_proxy/config"
	"github.com/stretchr/testify/assert"
)

// Testing imperative code is tough
func TestNewProxy(t *testing.T) {
	tempDir := t.TempDir()

	ca, err := newCA(tempDir)
	assert.NoError(t, err)

	p, err := newProxy(1, "localhost:8080", false, ca)
	assert.NoError(t, err)
	assert.NotNil(t, p)
}

func TestNewCA(t *testing.T) {
	tempDir := t.TempDir()

	ca, err := newCA(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, ca)
}

func TestConfigProxy(t *testing.T) {
	// Create a mock configuration
	cfg := config.NewDefaultConfig()
	cfg.CertDir = t.TempDir()
	cfg.AppMode = config.SimpleMode

	// Call the function with the mock configuration
	p, err := configProxy(cfg)

	// Assert that no error was returned
	assert.NoError(t, err)

	// Assert that a proxy was returned
	assert.NotNil(t, p)

	assert.Equal(t, 1, len(p.Addons))
}
