package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".config/etu/config.yaml")
}

func TestLoadConfig_NonExistent(t *testing.T) {
	// Create temp directory and override config path for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Contexts)
	assert.Empty(t, cfg.Contexts)
	assert.Empty(t, cfg.CurrentContext)
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create test config
	cfg := &Config{
		CurrentContext: "dev",
		LogLevel:       "debug",
		DefaultFormat:  "etcdctl",
		Strict:         true,
		NoValidate:     false,
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://localhost:2379"},
				Username:  "admin",
				Password:  "secret",
			},
			"prod": {
				Endpoints: []string{"http://prod:2379"},
				Username:  "user",
			},
		},
	}

	// Save config
	err := SaveConfig(cfg)
	require.NoError(t, err)

	// Verify file exists
	configPath, _ := GetConfigPath()
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load config
	loaded, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, cfg.CurrentContext, loaded.CurrentContext)
	assert.Equal(t, cfg.LogLevel, loaded.LogLevel)
	assert.Equal(t, cfg.DefaultFormat, loaded.DefaultFormat)
	assert.Equal(t, cfg.Strict, loaded.Strict)
	assert.Equal(t, cfg.NoValidate, loaded.NoValidate)
	assert.Len(t, loaded.Contexts, 2)
	assert.Equal(t, cfg.Contexts["dev"], loaded.Contexts["dev"])
	assert.Equal(t, cfg.Contexts["prod"], loaded.Contexts["prod"])
}

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Contexts: map[string]*ContextConfig{},
	}

	err := SaveConfig(cfg)
	require.NoError(t, err)

	configPath, _ := GetConfigPath()
	configDir := filepath.Dir(configPath)

	// Verify directory exists
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify directory permissions (0700)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}

func TestSaveConfig_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Contexts: map[string]*ContextConfig{},
	}

	err := SaveConfig(cfg)
	require.NoError(t, err)

	configPath, _ := GetConfigPath()
	info, err := os.Stat(configPath)
	require.NoError(t, err)

	// Verify file permissions (0600)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestGetCurrentContext(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name          string
		config        *Config
		expectContext *ContextConfig
		expectName    string
		expectError   bool
	}{
		{
			name: "valid current context",
			config: &Config{
				CurrentContext: "dev",
				Contexts: map[string]*ContextConfig{
					"dev": {
						Endpoints: []string{"http://localhost:2379"},
					},
				},
			},
			expectContext: &ContextConfig{
				Endpoints: []string{"http://localhost:2379"},
			},
			expectName:  "dev",
			expectError: false,
		},
		{
			name: "no current context",
			config: &Config{
				Contexts: map[string]*ContextConfig{
					"dev": {
						Endpoints: []string{"http://localhost:2379"},
					},
				},
			},
			expectContext: nil,
			expectName:    "",
			expectError:   false,
		},
		{
			name: "current context not found",
			config: &Config{
				CurrentContext: "nonexistent",
				Contexts: map[string]*ContextConfig{
					"dev": {
						Endpoints: []string{"http://localhost:2379"},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save config
			err := SaveConfig(tt.config)
			require.NoError(t, err)

			// Get current context
			ctx, name, err := GetCurrentContext()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectContext, ctx)
				assert.Equal(t, tt.expectName, name)
			}
		})
	}
}

func TestSetContext(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test adding new context
	ctx := &ContextConfig{
		Endpoints: []string{"http://localhost:2379"},
		Username:  "admin",
	}

	err := SetContext("dev", ctx, true)
	require.NoError(t, err)

	// Verify it was saved
	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.CurrentContext)
	assert.Equal(t, ctx, cfg.Contexts["dev"])

	// Test adding another context without making it current
	ctx2 := &ContextConfig{
		Endpoints: []string{"http://prod:2379"},
	}

	err = SetContext("prod", ctx2, false)
	require.NoError(t, err)

	cfg, err = LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.CurrentContext) // Should still be dev
	assert.Equal(t, ctx2, cfg.Contexts["prod"])
}

func TestSetContext_FirstContextBecomesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	ctx := &ContextConfig{
		Endpoints: []string{"http://localhost:2379"},
	}

	// First context should become current even if makeCurrent is false
	err := SetContext("dev", ctx, false)
	require.NoError(t, err)

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.CurrentContext)
}

func TestDeleteContext(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Setup initial config with multiple contexts
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://localhost:2379"},
			},
			"prod": {
				Endpoints: []string{"http://prod:2379"},
			},
		},
	}
	err := SaveConfig(cfg)
	require.NoError(t, err)

	// Delete non-current context
	err = DeleteContext("prod")
	require.NoError(t, err)

	cfg, err = LoadConfig()
	require.NoError(t, err)
	assert.Len(t, cfg.Contexts, 1)
	assert.NotContains(t, cfg.Contexts, "prod")
	assert.Equal(t, "dev", cfg.CurrentContext) // Should still be dev

	// Delete current context
	err = DeleteContext("dev")
	require.NoError(t, err)

	cfg, err = LoadConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.Contexts)
	assert.Empty(t, cfg.CurrentContext)
}

func TestDeleteContext_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://localhost:2379"},
			},
		},
	}
	err := SaveConfig(cfg)
	require.NoError(t, err)

	// Try to delete non-existent context
	err = DeleteContext("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUseContext(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Setup initial config
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://localhost:2379"},
			},
			"prod": {
				Endpoints: []string{"http://prod:2379"},
			},
		},
	}
	err := SaveConfig(cfg)
	require.NoError(t, err)

	// Switch context
	err = UseContext("prod")
	require.NoError(t, err)

	cfg, err = LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "prod", cfg.CurrentContext)
}

func TestUseContext_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg := &Config{
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://localhost:2379"},
			},
		},
	}
	err := SaveConfig(cfg)
	require.NoError(t, err)

	// Try to use non-existent context
	err = UseContext("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
