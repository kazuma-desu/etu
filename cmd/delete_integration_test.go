//go:build integration

package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCommand_Integration(t *testing.T) {
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

	t.Run("Delete single key", func(t *testing.T) {
		err := etcdClient.Put(ctx, "/delete/test/single", "to-be-deleted")
		require.NoError(t, err)

		resetDeleteFlags()

		err = runDelete(deleteCmd, []string{"/delete/test/single"})
		require.NoError(t, err)

		_, err = etcdClient.Get(ctx, "/delete/test/single")
		assert.Error(t, err)
	})

	t.Run("Delete non-existent key shows warning", func(t *testing.T) {
		resetDeleteFlags()

		old := os.Stdout
		defer func() { os.Stdout = old }()

		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runDelete(deleteCmd, []string{"/delete/nonexistent/key"})

		w.Close()

		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "not found")
	})

	t.Run("Delete with prefix and force", func(t *testing.T) {
		err := etcdClient.Put(ctx, "/delete/prefix/key1", "value1")
		require.NoError(t, err)
		err = etcdClient.Put(ctx, "/delete/prefix/key2", "value2")
		require.NoError(t, err)
		err = etcdClient.Put(ctx, "/delete/prefix/key3", "value3")
		require.NoError(t, err)

		resetDeleteFlags()
		deleteOpts.prefix = true
		deleteOpts.force = true

		err = runDelete(deleteCmd, []string{"/delete/prefix/"})
		require.NoError(t, err)

		_, err = etcdClient.Get(ctx, "/delete/prefix/key1")
		assert.Error(t, err)
		_, err = etcdClient.Get(ctx, "/delete/prefix/key2")
		assert.Error(t, err)
		_, err = etcdClient.Get(ctx, "/delete/prefix/key3")
		assert.Error(t, err)
	})

	t.Run("Delete prefix dry-run shows keys", func(t *testing.T) {
		err := etcdClient.Put(ctx, "/delete/dryrun/a", "1")
		require.NoError(t, err)
		err = etcdClient.Put(ctx, "/delete/dryrun/b", "2")
		require.NoError(t, err)

		resetDeleteFlags()
		deleteOpts.prefix = true
		deleteOpts.dryRun = true

		old := os.Stdout
		defer func() { os.Stdout = old }()

		r, w, _ := os.Pipe()
		os.Stdout = w

		err = runDelete(deleteCmd, []string{"/delete/dryrun/"})

		w.Close()

		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "/delete/dryrun/a")
		assert.Contains(t, output, "/delete/dryrun/b")
		assert.Contains(t, output, "Would delete")

		value, err := etcdClient.Get(ctx, "/delete/dryrun/a")
		require.NoError(t, err)
		assert.Equal(t, "1", value)
	})

	t.Run("Delete error without leading slash", func(t *testing.T) {
		resetDeleteFlags()

		err := runDelete(deleteCmd, []string{"invalid-key"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key must start with '/'")
	})

	t.Run("Delete prefix with no matching keys", func(t *testing.T) {
		resetDeleteFlags()
		deleteOpts.prefix = true
		deleteOpts.force = true

		old := os.Stdout
		defer func() { os.Stdout = old }()

		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runDelete(deleteCmd, []string{"/nonexistent/prefix/"})

		w.Close()

		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "No keys found")
	})

	t.Run("Delete prefix with confirmation accepted", func(t *testing.T) {
		err := etcdClient.Put(ctx, "/delete/confirm/key1", "val1")
		require.NoError(t, err)

		resetDeleteFlags()
		deleteOpts.prefix = true

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		defer r.Close()
		os.Stdin = r

		go func() {
			w.WriteString("y\n")
			w.Close()
		}()

		oldStdout := os.Stdout
		defer func() { os.Stdout = oldStdout }()

		rOut, wOut, _ := os.Pipe()
		os.Stdout = wOut

		err = runDelete(deleteCmd, []string{"/delete/confirm/"})

		wOut.Close()

		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		rOut.Close()

		require.NoError(t, err)

		_, err = etcdClient.Get(ctx, "/delete/confirm/key1")
		assert.Error(t, err)
	})

	t.Run("Delete prefix with confirmation rejected", func(t *testing.T) {
		err := etcdClient.Put(ctx, "/delete/reject/key1", "val1")
		require.NoError(t, err)

		resetDeleteFlags()
		deleteOpts.prefix = true

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		defer r.Close()
		os.Stdin = r

		go func() {
			w.WriteString("n\n")
			w.Close()
		}()

		oldStdout := os.Stdout
		defer func() { os.Stdout = oldStdout }()

		rOut, wOut, _ := os.Pipe()
		os.Stdout = wOut

		err = runDelete(deleteCmd, []string{"/delete/reject/"})

		wOut.Close()

		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		rOut.Close()
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "canceled")

		value, err := etcdClient.Get(ctx, "/delete/reject/key1")
		require.NoError(t, err)
		assert.Equal(t, "val1", value)
	})
}

func resetDeleteFlags() {
	deleteOpts.prefix = false
	deleteOpts.force = false
	deleteOpts.dryRun = false
}

func TestConfirmDeletionIntegration(t *testing.T) {
	keys := []string{"/a", "/b", "/c"}

	t.Run("Accepts y", func(t *testing.T) {
		in := strings.NewReader("y\n")
		out := &bytes.Buffer{}
		result := confirmDeletion(keys, "/prefix", in, out)
		assert.True(t, result)
		assert.Contains(t, out.String(), "/a")
		assert.Contains(t, out.String(), "/b")
		assert.Contains(t, out.String(), "/c")
	})

	t.Run("Accepts yes", func(t *testing.T) {
		in := strings.NewReader("yes\n")
		out := &bytes.Buffer{}
		result := confirmDeletion(keys, "/prefix", in, out)
		assert.True(t, result)
	})

	t.Run("Rejects n", func(t *testing.T) {
		in := strings.NewReader("n\n")
		out := &bytes.Buffer{}
		result := confirmDeletion(keys, "/prefix", in, out)
		assert.False(t, result)
	})

	t.Run("Rejects empty", func(t *testing.T) {
		in := strings.NewReader("\n")
		out := &bytes.Buffer{}
		result := confirmDeletion(keys, "/prefix", in, out)
		assert.False(t, result)
	})

	t.Run("Rejects on EOF", func(t *testing.T) {
		in := strings.NewReader("")
		out := &bytes.Buffer{}
		result := confirmDeletion(keys, "/prefix", in, out)
		assert.False(t, result)
	})
}
