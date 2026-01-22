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

// captureStdout captures stdout functionality within a function.
// It ensures proper cleanup even if f() panics by using deferred restoration
// of os.Stdout and closing of the pipe ends.
func captureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("captureStdout: failed to create pipe: %w", pipeErr)
	}

	defer func() {
		w.Close()
		os.Stdout = old
	}()

	os.Stdout = w

	var fErr error
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				fErr = fmt.Errorf("captureStdout: f() panicked: %v", rec)
			}
		}()
		fErr = f()
	}()

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	r.Close()

	return buf.String(), fErr
}

// TestDiffCommand_Integration tests the diff command against a real etcd instance.
func TestDiffCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	originalOpts := diffOpts
	defer func() { diffOpts = originalOpts }()

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

	t.Run("Diff file-scoped (default)", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff.txt")

		// File content:
		// /app/config/key1 = "new_value" (modified)
		// /app/config/key3 = "unchanged" (unchanged)
		// /app/config/key4 = "added_value" (added)
		// Note: /app/config/key2 exists in etcd but NOT in file
		// With file-scoped diff (default), key2 should NOT appear as deleted
		content := `/app/config/key1
new_value

/app/config/key3
unchanged

/app/config/key4
added_value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = ""
		diffOpts.Full = false

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		// key4: added (in file, not in etcd)
		assert.Contains(t, output, "+")
		assert.Contains(t, output, "/app/config/key4")
		// key1: modified (different value)
		assert.Contains(t, output, "~")
		assert.Contains(t, output, "/app/config/key1")
		// key2: NOT shown (not in file, file-scoped diff ignores etcd-only keys)
		assert.NotContains(t, output, "/app/config/key2")
		// key3: unchanged, not shown by default
		assert.NotContains(t, output, "/app/config/key3")
	})

	t.Run("Diff with --full shows etcd-only keys as deleted", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_full.txt")

		// File only has key1 and key3
		// key2 exists in etcd but not in file - should show as deleted with --full
		content := `/app/config/key1
new_value

/app/config/key3
unchanged
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = "/app/config"
		diffOpts.Full = true

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		// key1: modified
		assert.Contains(t, output, "~")
		assert.Contains(t, output, "/app/config/key1")
		// key2: deleted (in etcd under prefix, not in file)
		assert.Contains(t, output, "-")
		assert.Contains(t, output, "/app/config/key2")
	})

	t.Run("Diff show unchanged", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_unchanged.txt")

		content := `/app/config/key3
unchanged
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = true
		diffOpts.Prefix = ""
		diffOpts.Full = false

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
		diffOpts.Full = false

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

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
		diffOpts.Full = false

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

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
		diffOpts.Full = false

		output, err := captureStdout(func() error {
			return runDiff(diffCmd, []string{})
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/app/config/key1")
		assert.NotContains(t, output, "/other/key")
	})

	t.Run("Full flag requires prefix", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "diff_full_no_prefix.txt")

		content := `/app/config/key1
value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		t.Setenv("HOME", tempDir)
		t.Setenv("ETCD_ENDPOINTS", endpoint)

		diffOpts.FilePath = configFile
		diffOpts.Format = "simple"
		diffOpts.ShowUnchanged = false
		diffOpts.Prefix = ""
		diffOpts.Full = true

		err = runDiff(diffCmd, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--full requires --prefix")
	})
}
