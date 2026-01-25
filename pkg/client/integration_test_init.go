//go:build integration

package client

import "os"

func init() {
	// Disable Ryuk for Podman compatibility - must be set before testcontainers import
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}
