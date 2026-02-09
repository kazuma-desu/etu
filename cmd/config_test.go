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

		err := runGetContexts(getContextsCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Current context with no active context", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runCurrentContext(currentContextCmd, []string{})
		assert.NoError(t, err)
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

		err = runUseContext(useContextCmd, []string{"test-context"})
		assert.NoError(t, err)

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

		err = runDeleteContext(deleteContextCmd, []string{"delete-me"})
		assert.NoError(t, err)

		// Verify it's gone
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Nil(t, cfg.Contexts["delete-me"])
	})

	t.Run("Set config value", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"log-level", "debug"})
		assert.NoError(t, err)

		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("Set config default format", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"default-format", "etcdctl"})
		assert.NoError(t, err)

		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "etcdctl", cfg.DefaultFormat)
	})

	t.Run("Set config strict mode", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"strict", "true"})
		assert.NoError(t, err)

		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.True(t, cfg.Strict)
	})

	t.Run("Set config no-validate", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"no-validate", "true"})
		assert.NoError(t, err)

		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		assert.True(t, cfg.NoValidate)
	})

	t.Run("View config", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		// Set outputFormat to a valid format for config view
		originalFormat := outputFormat
		defer func() { outputFormat = originalFormat }()
		outputFormat = "json"

		ctxConfig := &config.ContextConfig{
			Endpoints: []string{"localhost:2379"},
			Username:  "testuser",
			Password:  "",
		}

		err := config.SetContext("view-test", ctxConfig, true)
		require.NoError(t, err)

		err = runViewConfig(viewConfigCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Get contexts with multiple contexts", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

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

		err = runGetContexts(getContextsCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Use nonexistent context returns error", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runUseContext(useContextCmd, []string{"nonexistent-context"})
		assert.Error(t, err)
	})

	t.Run("Delete nonexistent context returns error", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runDeleteContext(deleteContextCmd, []string{"nonexistent-context"})
		assert.Error(t, err)
	})

	t.Run("Set invalid log level returns error", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"log-level", "invalid-level"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log level")
	})

	t.Run("Set invalid format returns error", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"default-format", "invalid-format"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})

	t.Run("Set unknown config key returns error", func(t *testing.T) {
		cleanup := setupTestConfig(t)
		defer cleanup()

		err := runSetConfig(setConfigCmd, []string{"unknown-key", "value"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown configuration key")
	})
}
