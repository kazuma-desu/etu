package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunLs_InvalidKey(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()

	origContextName := contextName
	defer func() { contextName = origContextName }()
	contextName = ""

	err := runLs(nil, []string{"invalid-key"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must start with '/'")
}

func TestRunLs_InvalidOutputFormat(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()

	origFormat := outputFormat
	origContextName := contextName
	defer func() {
		outputFormat = origFormat
		contextName = origContextName
	}()

	outputFormat = "invalid-format"
	contextName = ""

	err := runLs(nil, []string{"/"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}
