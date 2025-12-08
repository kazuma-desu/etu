package cmd

import (
	"os"
	"testing"

	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) func() {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Set the config path via environment variable or test setup
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	cleanup := func() {
		os.Setenv("HOME", oldHome)
	}

	return cleanup
}

func TestConfigCommands(t *testing.T) {
	t.Run("Get contexts with empty config", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// This should not panic even with no config file
		runGetContexts(getContextsCmd, []string{})
	})

	t.Run("Current context with no active context", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Should handle gracefully
		runCurrentContext(currentContextCmd, []string{})
	})

	t.Run("Use context", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// First create a context
		ctxConfig := &config.ContextConfig{
			Endpoints: []string{"localhost:2379"},
			Username:  "test",
			Password:  "pass",
		}

		err := config.SetContext("test-context", ctxConfig, true)
		require.NoError(t, err)

		// Now switch to it
		runUseContext(useContextCmd, []string{"test-context"})

		// Verify it's current
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-context", cfg.CurrentContext)
	})

	t.Run("Delete context", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Create a context first
		ctxConfig := &config.ContextConfig{
			Endpoints: []string{"localhost:2379"},
			Username:  "",
			Password:  "",
		}

		err := config.SetContext("delete-me", ctxConfig, false)
		require.NoError(t, err)

		// Delete it
		runDeleteContext(deleteContextCmd, []string{"delete-me"})

		// Verify it's gone
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Nil(t, cfg.Contexts["delete-me"])
	})

	t.Run("Set config value", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Set log level
		runSetConfig(setConfigCmd, []string{"log-level", "debug"})

		// Verify it was set
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("Set config default format", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Set default format
		runSetConfig(setConfigCmd, []string{"default-format", "etcdctl"})

		// Verify it was set
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "etcdctl", cfg.DefaultFormat)
	})

	t.Run("Set config strict mode", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Enable strict mode
		runSetConfig(setConfigCmd, []string{"strict", "true"})

		// Verify it was set
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.True(t, cfg.Strict)
	})

	t.Run("Set config no-validate", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Enable no-validate
		runSetConfig(setConfigCmd, []string{"no-validate", "true"})

		// Verify it was set
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.True(t, cfg.NoValidate)
	})

	t.Run("View config", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Create some config
		ctxConfig := &config.ContextConfig{
			Endpoints: []string{"localhost:2379"},
			Username:  "testuser",
			Password:  "",
		}

		err := config.SetContext("view-test", ctxConfig, true)
		require.NoError(t, err)

		// View should not panic
		runViewConfig(viewConfigCmd, []string{})
	})

	t.Run("Get contexts with multiple contexts", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Create multiple contexts
		ctx1 := &config.ContextConfig{
			Endpoints: []string{"localhost:2379"},
			Username:  "user1",
			Password:  "",
		}
		ctx2 := &config.ContextConfig{
			Endpoints: []string{"remote:2379"},
			Username:  "user2",
			Password:  "",
		}

		err := config.SetContext("ctx1", ctx1, true)
		require.NoError(t, err)

		err = config.SetContext("ctx2", ctx2, false)
		require.NoError(t, err)

		// List contexts should not panic
		runGetContexts(getContextsCmd, []string{})
	})
}
