package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsCommand_Output(t *testing.T) {
	// Save current rootCmd output to avoid test leakage (cross-test isolation)
	oldOut := rootCmd.OutOrStderr()
	defer rootCmd.SetOut(oldOut)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"options"})

	err := rootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "--cacert")
	assert.Contains(t, output, "--cert")
	assert.Contains(t, output, "--key")
	assert.Contains(t, output, "--context")
	assert.Contains(t, output, "--insecure-skip-tls-verify")
	assert.Contains(t, output, "--username")
	assert.Contains(t, output, "--password")
	assert.Contains(t, output, "--password-stdin")
	assert.Contains(t, output, "--timeout")
	assert.Contains(t, output, "--output")
	assert.Contains(t, output, "--log-level")
}

func TestOptionsCommand_HiddenFromMainHelp(t *testing.T) {
	hiddenFlags := []string{"cacert", "cert", "key", "insecure-skip-tls-verify", "username", "password", "password-stdin"}

	for _, flagName := range hiddenFlags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		assert.NotNil(t, flag, "flag %s should exist", flagName)
		assert.True(t, flag.Hidden, "flag %s should be hidden", flagName)
	}
}

func TestOptionsCommand_VisibleFlags(t *testing.T) {
	visibleFlags := []string{"context", "output", "timeout", "log-level"}

	for _, flagName := range visibleFlags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		assert.NotNil(t, flag, "flag %s should exist", flagName)
		assert.False(t, flag.Hidden, "flag %s should be visible", flagName)
	}
}
