//go:build integration

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	fullEndpoint := setupEtcdContainerForCmd(t)

	t.Run("Login with valid endpoint", func(t *testing.T) {
		loginContextName = "test-context"
		loginEndpoints = []string{fullEndpoint}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = false

		err := runLogin(loginCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Login with no-test flag", func(t *testing.T) {
		loginContextName = "notest-context"
		loginEndpoints = []string{fullEndpoint}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = true

		err := runLogin(loginCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Login with multiple endpoints", func(t *testing.T) {
		loginContextName = "multi-context"
		loginEndpoints = []string{fullEndpoint, fullEndpoint}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = true

		err := runLogin(loginCmd, []string{})
		assert.NoError(t, err)
	})
}
