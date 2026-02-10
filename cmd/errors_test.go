package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapNotConnectedError(t *testing.T) {
	originalErr := errors.New("no current context set")
	wrappedErr := wrapNotConnectedError(originalErr)

	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "✗ not connected")
	assert.Contains(t, wrappedErr.Error(), "no current context set")
	assert.Contains(t, wrappedErr.Error(), "Use 'etu login' to configure a context")
	assert.ErrorIs(t, wrappedErr, originalErr)
}
