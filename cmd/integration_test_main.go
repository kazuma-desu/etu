//go:build integration

package cmd

import "os"

func init() {
	// Disable Ryuk for Podman compatibility - must be set before testcontainers import
	// Ryuk causes "container not found" errors with Podman due to database sync issues
	// Containers are cleaned up via t.Cleanup() instead
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}
