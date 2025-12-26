package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginCommand_ErrorReturns(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("Login with invalid endpoint returns error", func(t *testing.T) {
		loginEndpoints = []string{"http://invalid-host-that-does-not-exist:2379"}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = false

		err := runLogin(loginCmd, []string{"invalid-context"})
		assert.Error(t, err)
	})

	t.Run("Login saves config even with failed connection test when user declines", func(t *testing.T) {
		loginEndpoints = []string{"http://localhost:9999"}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = true

		err := runLogin(loginCmd, []string{"notest-saves-context"})
		assert.NoError(t, err)
	})
}
