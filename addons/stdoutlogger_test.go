package addons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStdOutLogger(t *testing.T) {
	logger := NewStdOutLogger()
	assert.NotNil(t, logger)
}
