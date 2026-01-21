//go:build integration

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CaptureStdout captures stdout functionality within a function
func captureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("captureStdout: failed to create pipe: %w", pipeErr)
	}
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

// TestDiffCommand_Integration tests the diff command against a real etcd instance.
// Note: This test mutates package-level diffOpts and cannot be run in parallel.
func TestDiffCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint, cleanup := setupEtcdContainerForCmd(t)
	defer cleanup()

	// Wait for etcd to be ready
	time.Sleep(2 * time.Second)

	// Setup etcd client
	cfg := &client.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	}
	etcdClient, err := client.NewClient(cfg)
	require.NoError(t, err)
	defer etcdClient.Close()
	ctx := context.Background()

	// 1. Pre-populate etcd with some data
	// /app/config/key1 = "old_value" (modified in file)
	// /app/config/key2 = "deleted_value" (deleted in file)
	// /app/config/key3 = "unchanged" (unchanged)
	err = etcdClient.Put(ctx, "/app/config/key1", "old_value")
	require.NoError(t, err)
	err = etcdClient.Put(ctx, "/app/config/key2", "deleted_value")
	require.NoError(t, err)
	err = etcdClient.Put(ctx, "/app/config/key3", "unchanged")
	require.NoError(t, err)

	t.Run("Diff with changes", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff.txt")

		// File content:
		// /app/config/key1 = "new_value" (modified)
		// /app/config/key3 = "unchanged" (unchanged)
		// /app/config/key4 = "added_value" (added)
		content := `/app/config/key1
new_value

/app/config/key3
unchanged

/app/config/key4
added_value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		// Isolate config to prevent loading user's config
		t.Setenv("HOME", tempDir)
		// Setup env
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		// Setup options
		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = ""

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		// Verify output contains expected changes
		assert.Contains(t, output, "+")
		assert.Contains(t, output, "/app/config/key4")
		assert.Contains(t, output, "~")
		assert.Contains(t, output, "/app/config/key1")
		assert.Contains(t, output, "-")
		assert.Contains(t, output, "/app/config/key2")
		assert.NotContains(t, output, "/app/config/key3") // Not shown by default
	})

	t.Run("Diff show unchanged", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_unchanged.txt")

		content := `/app/config/key3
unchanged
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		// Isolate config
		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = true
		diffOpts.Prefix = ""

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		assert.Contains(t, output, "=")
		assert.Contains(t, output, "/app/config/key3")
	})

	t.Run("Diff with JSON output", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_json.txt")

		content := `/app/config/key1
new_value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "json"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = ""

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		// Verify JSON structure
		assert.Contains(t, output, "\"added\"")
		assert.Contains(t, output, "\"modified\"")
		assert.Contains(t, output, "\"deleted\"")
	})

	t.Run("Diff with table output", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_table.txt")

		content := `/app/config/key1
new_value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "table"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = ""

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		// Verify table headers
		assert.Contains(t, output, "STATUS")
		assert.Contains(t, output, "KEY")
	})

	t.Run("Diff with prefix filter", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_prefix.txt")

		content := `/app/config/key1
new_value

/other/key
other_value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = "/app/config"

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		// Should only show /app/config keys
		assert.Contains(t, output, "/app/config/key1")
		assert.NotContains(t, output, "/other/key")
	})
}
