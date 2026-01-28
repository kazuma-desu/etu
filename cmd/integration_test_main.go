//go:build integration

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kazuma-desu/etu/pkg/config"
)

func init() {
	// Disable Ryuk for Podman compatibility - must be set before testcontainers import
	// Ryuk causes "container not found" errors with Podman due to database sync issues
	// Containers are cleaned up via t.Cleanup() instead
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}

func setupTestContext(t *testing.T, endpoint string) string {
	t.Helper()

	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "etu")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.ContextConfig{
			"test": {
				Endpoints: []string{endpoint},
			},
		},
	}

	t.Setenv("ETUCONFIG", filepath.Join(configDir, "config.yaml"))

	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	return tempDir
}
