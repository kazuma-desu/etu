//go:build integration

package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint := setupEtcdContainerForCmd(t)

	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	appCfg := &config.Config{
		Contexts: map[string]*config.ContextConfig{
			"test": {
				Endpoints: []string{endpoint},
			},
		},
		CurrentContext: "test",
	}
	err := config.SaveConfig(appCfg)
	require.NoError(t, err)

	cfg := &client.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	}
	etcdClient, err := client.NewClient(cfg)
	require.NoError(t, err)
	defer etcdClient.Close()

	ctx := context.Background()

	t.Run("Put with inline value", func(t *testing.T) {
		resetPutFlags()

		err := runPut(putCmd, []string{"/put/test/inline", "hello-world"})
		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/inline")
		require.NoError(t, err)
		assert.Equal(t, "hello-world", value)
	})

	t.Run("Put from stdin", func(t *testing.T) {
		resetPutFlags()

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		defer r.Close()
		os.Stdin = r

		go func() {
			w.WriteString("value-from-stdin")
			w.Close()
		}()

		err := runPut(putCmd, []string{"/put/test/stdin", "-"})

		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/stdin")
		require.NoError(t, err)
		assert.Equal(t, "value-from-stdin", value)
	})

	t.Run("Put multiline from stdin", func(t *testing.T) {
		resetPutFlags()

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		defer r.Close()
		os.Stdin = r

		go func() {
			w.WriteString("line1\nline2\nline3")
			w.Close()
		}()

		err := runPut(putCmd, []string{"/put/test/multiline", "-"})

		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/multiline")
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\nline3", value)
	})

	t.Run("Put with dry-run does not write", func(t *testing.T) {
		resetPutFlags()
		putOpts.dryRun = true

		old := os.Stdout
		defer func() { os.Stdout = old }()

		r, w, pipeErr := os.Pipe()
		require.NoError(t, pipeErr)
		defer r.Close()
		os.Stdout = w

		err := runPut(putCmd, []string{"/put/test/dryrun", "should-not-exist"})

		require.NoError(t, w.Close())

		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, r)
		require.NoError(t, copyErr)
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "Would put")

		_, err = etcdClient.Get(ctx, "/put/test/dryrun")
		assert.Error(t, err)
	})

	t.Run("Put with validate flag", func(t *testing.T) {
		resetPutFlags()
		putOpts.validate = true

		err := runPut(putCmd, []string{"/put/test/validated", "valid-value"})
		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/validated")
		require.NoError(t, err)
		assert.Equal(t, "valid-value", value)
	})

	t.Run("Put error without leading slash", func(t *testing.T) {
		resetPutFlags()

		err := runPut(putCmd, []string{"invalid-key", "value"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key must start with '/'")
	})

	t.Run("Put overwrites existing value", func(t *testing.T) {
		resetPutFlags()

		err := runPut(putCmd, []string{"/put/test/overwrite", "original"})
		require.NoError(t, err)

		err = runPut(putCmd, []string{"/put/test/overwrite", "updated"})
		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/overwrite")
		require.NoError(t, err)
		assert.Equal(t, "updated", value)
	})

	t.Run("Put with special characters in value", func(t *testing.T) {
		resetPutFlags()

		specialValue := `{"key": "value", "nested": {"a": 1}}`
		err := runPut(putCmd, []string{"/put/test/json", specialValue})
		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/json")
		require.NoError(t, err)
		assert.Equal(t, specialValue, value)
	})

	t.Run("Put with unicode value", func(t *testing.T) {
		resetPutFlags()

		unicodeValue := "Hello 世界 🌍"
		err := runPut(putCmd, []string{"/put/test/unicode", unicodeValue})
		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/unicode")
		require.NoError(t, err)
		assert.Equal(t, unicodeValue, value)
	})

	t.Run("Put with empty value from args", func(t *testing.T) {
		resetPutFlags()

		err := runPut(putCmd, []string{"/put/test/empty", ""})
		require.NoError(t, err)

		value, err := etcdClient.Get(ctx, "/put/test/empty")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})
}

func resetPutFlags() {
	putOpts.dryRun = false
	putOpts.validate = false
}
