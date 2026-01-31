//go:build integration

package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestWatchCommand_Integration(t *testing.T) {
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

	t.Run("watch single key", func(t *testing.T) {
		// Create key first
		err := etcdClient.Put(ctx, "/watch/cmd/key", "value1")
		require.NoError(t, err)

		// Start watching with timeout
		watchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		opts := &client.WatchOptions{}
		watchChan := etcdClient.Watch(watchCtx, "/watch/cmd/key", opts)

		// Update key in background
		go func() {
			time.Sleep(100 * time.Millisecond)
			err := etcdClient.Put(ctx, "/watch/cmd/key", "value2")
			require.NoError(t, err)
		}()

		// Wait for event
		select {
		case resp := <-watchChan:
			require.Len(t, resp.Events, 1)
			require.Equal(t, "/watch/cmd/key", resp.Events[0].Key)
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for watch event")
		}
	})

	t.Run("watch with prefix", func(t *testing.T) {
		watchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		opts := &client.WatchOptions{Prefix: true}
		watchChan := etcdClient.Watch(watchCtx, "/watch/prefix/cmd/", opts)

		// Create keys in background
		go func() {
			time.Sleep(100 * time.Millisecond)
			err := etcdClient.Put(ctx, "/watch/prefix/cmd/key1", "value1")
			require.NoError(t, err)
		}()

		// Wait for event
		select {
		case resp := <-watchChan:
			require.Len(t, resp.Events, 1)
			require.Equal(t, "/watch/prefix/cmd/key1", resp.Events[0].Key)
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for watch event")
		}
	})
}
