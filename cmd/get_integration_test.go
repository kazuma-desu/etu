//go:build integration

package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/stretchr/testify/require"
)

// captureStdout captures stdout from f() with panic recovery.
func captureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("captureStdout: failed to create pipe: %w", pipeErr)
	}

	// Read from pipe in goroutine to avoid blocking
	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		outCh <- buf.String()
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

	// Close writer to signal EOF to goroutine, then restore stdout
	w.Close()
	os.Stdout = old

	// Wait for goroutine to finish reading
	output := <-outCh

	return output, fErr
}

func TestGetCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	fullEndpoint := setupEtcdContainerForCmd(t)
	ctx := context.Background()

	cfg := &client.Config{
		Endpoints: []string{fullEndpoint},
	}
	etcdClient, err := client.NewClient(cfg)
	require.NoError(t, err)
	defer etcdClient.Close()

	// Setup context in config
	appCfg := &config.Config{
		Contexts: map[string]*config.ContextConfig{
			"test": {
				Endpoints: []string{fullEndpoint},
			},
		},
		CurrentContext: "test",
	}
	err = config.SaveConfig(appCfg)
	require.NoError(t, err)

	// Populate test data
	testData := map[string]string{
		"/config/app/host":     "localhost",
		"/config/app/port":     "8080",
		"/config/app/database": "postgres",
		"/config/db/host":      "db.example.com",
		"/config/db/port":      "5432",
		"/other/key":           "value",
	}

	for k, v := range testData {
		err := etcdClient.Put(ctx, k, v)
		require.NoError(t, err)
	}

	t.Run("Get single key", func(t *testing.T) {
		resetGetFlags()

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/host"})
		})
		require.NoError(t, err)

		require.Contains(t, output, "/config/app/host")
		require.Contains(t, output, "localhost")
	})

	t.Run("Get with prefix", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.Contains(t, output, "/config/app/host")
		require.Contains(t, output, "/config/app/port")
		require.Contains(t, output, "/config/app/database")
		require.NotContains(t, output, "/config/db/")
	})

	t.Run("Get with prefix and limit", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.limit = 2

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		// Should only have 2 keys
		lines := strings.Split(strings.TrimSpace(output), "\n")
		// Each key-value pair takes 2 lines
		require.LessOrEqual(t, len(lines), 4)
	})

	t.Run("Get keys only", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.keysOnly = true

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.Contains(t, output, "/config/app/host")
		require.NotContains(t, output, "localhost")
		require.NotContains(t, output, "8080")
	})

	t.Run("Get count only", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.countOnly = true

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.Equal(t, "3", strings.TrimSpace(output))
	})

	t.Run("Get with print-value-only", func(t *testing.T) {
		resetGetFlags()
		getOpts.printValue = true

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/host"})
		})
		require.NoError(t, err)

		require.Equal(t, "localhost", strings.TrimSpace(output))
		require.NotContains(t, output, "/config/app/host")
	})

	t.Run("Get with JSON output", func(t *testing.T) {
		resetGetFlags()
		outputFormat = "json"

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/host"})
		})
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		kvs := result["kvs"].([]any)
		require.Len(t, kvs, 1)

		kv := kvs[0].(map[string]any)
		// Keys and values are base64 encoded in JSON output
		keyBytes, _ := base64.StdEncoding.DecodeString(kv["key"].(string))
		valueBytes, _ := base64.StdEncoding.DecodeString(kv["value"].(string))
		require.Equal(t, "/config/app/host", string(keyBytes))
		require.Equal(t, "localhost", string(valueBytes))
	})

	t.Run("Get with range", func(t *testing.T) {
		resetGetFlags()

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/database", "/config/app/port"})
		})
		require.NoError(t, err)

		// Range is [start, end) so should include database and host but not port
		require.Contains(t, output, "/config/app/database")
		require.Contains(t, output, "/config/app/host")
	})

	t.Run("Get from-key", func(t *testing.T) {
		resetGetFlags()
		getOpts.fromKey = true
		getOpts.limit = 10

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/db/"})
		})
		require.NoError(t, err)

		require.Contains(t, output, "/config/db/host")
		require.Contains(t, output, "/other/key")
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		resetGetFlags()

		err := runGet(getCmd, []string{"/nonexistent/key"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "key not found")
	})

	t.Run("Get with sort order", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortOrder = "DESCEND"
		getOpts.sortTarget = "KEY"

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		// Check that keys appear in descending order
		portIdx := strings.Index(output, "/config/app/port")
		hostIdx := strings.Index(output, "/config/app/host")
		dbIdx := strings.Index(output, "/config/app/database")

		require.True(t, portIdx < hostIdx)
		require.True(t, hostIdx < dbIdx)
	})

	t.Run("Get with invalid sort order", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortOrder = "INVALID"

		err := runGet(getCmd, []string{"/config/app/"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid sort order")
	})

	t.Run("Get with invalid sort target", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortTarget = "INVALID"

		err := runGet(getCmd, []string{"/config/app/"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid sort target")
	})

	t.Run("Get with VERSION sort target", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortTarget = "VERSION"
		getOpts.sortOrder = "ASCEND"

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.NotEmpty(t, output)
	})

	t.Run("Get with CREATE sort target", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortTarget = "CREATE"

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.NotEmpty(t, output)
	})

	t.Run("Get with MODIFY sort target", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortTarget = "MODIFY"

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.NotEmpty(t, output)
	})

	t.Run("Get with VALUE sort target", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.sortTarget = "VALUE"

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.NotEmpty(t, output)
	})

	t.Run("Get with revision filters", func(t *testing.T) {
		resetGetFlags()
		getOpts.prefix = true
		getOpts.minModRev = 1
		getOpts.maxModRev = 1000
		getOpts.minCreateRev = 1
		getOpts.maxCreateRev = 1000

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/"})
		})
		require.NoError(t, err)

		require.NotEmpty(t, output)
	})

	t.Run("Get with revision option", func(t *testing.T) {
		// First, get the current key to know it exists
		resp, err := etcdClient.GetWithOptions(ctx, "/config/app/host", &client.GetOptions{})
		require.NoError(t, err)
		require.NotEmpty(t, resp.Kvs)

		resetGetFlags()
		// Use the current revision to ensure the key exists at that revision
		getOpts.revision = resp.Kvs[0].CreateRevision

		output, err := captureStdout(func() error {
			return runGet(getCmd, []string{"/config/app/host"})
		})
		require.NoError(t, err)

		require.NotEmpty(t, output)
		require.Contains(t, output, "/config/app/host")
	})
}

func resetGetFlags() {
	getOpts.prefix = false
	getOpts.fromKey = false
	getOpts.limit = 0
	getOpts.revision = 0
	getOpts.sortOrder = ""
	getOpts.sortTarget = ""
	getOpts.keysOnly = false
	getOpts.countOnly = false
	getOpts.printValue = false
	outputFormat = "simple"
	getOpts.consistency = "l"
	getOpts.minModRev = 0
	getOpts.maxModRev = 0
	getOpts.minCreateRev = 0
	getOpts.maxCreateRev = 0
	getOpts.showMetadata = false
	getOpts.rangeEnd = ""
}
