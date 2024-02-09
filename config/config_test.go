package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_SetLoggerLevel_ImplicitCallToSetLoggerLevel(t *testing.T) {
	cfg := &Config{}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	assert.Equal(t, 0, cfg.GetDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	assert.IsType(t, &Config{}, cfg)
}
