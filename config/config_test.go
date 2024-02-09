package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_SetLoggerLevel_Debug(t *testing.T) {
	cfg := &Config{
		Debug: true,
	}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	cfg.SetLoggerLevel()

	assert.Equal(t, 1, cfg.GetDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}

func TestConfig_SetLoggerLevel_Verbose(t *testing.T) {
	cfg := &Config{
		Verbose: true,
	}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	cfg.SetLoggerLevel()

	assert.Equal(t, 0, cfg.GetDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}

func TestConfig_SetLoggerLevel_Default(t *testing.T) {
	cfg := &Config{}

	assert.Equal(t, false, cfg.logLevelHasBeenSet)
	cfg.SetLoggerLevel()

	assert.Equal(t, 0, cfg.GetDebugLevel())
	assert.Equal(t, true, cfg.logLevelHasBeenSet)
}

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
