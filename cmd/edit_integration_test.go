//go:build integration

package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEditCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test config
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Start etcd container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "quay.io/coreos/etcd:v3.5.9",
		ExposedPorts: []string{"2379/tcp"},
		Env: map[string]string{
			"ETCD_NAME":                        "test-etcd",
			"ETCD_ADVERTISE_CLIENT_URLS":       "http://0.0.0.0:2379",
			"ETCD_LISTEN_CLIENT_URLS":          "http://0.0.0.0:2379",
			"ETCD_INITIAL_ADVERTISE_PEER_URLS": "http://0.0.0.0:2380",
			"ETCD_LISTEN_PEER_URLS":            "http://0.0.0.0:2380",
			"ETCD_INITIAL_CLUSTER":             "test-etcd=http://0.0.0.0:2380",
		},
		WaitingFor: wait.ForLog("ready to serve client requests").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)
	fullEndpoint := "http://" + endpoint

	// Wait for etcd to be ready
	time.Sleep(2 * time.Second)

	// Setup etcd client
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

	t.Run("Edit existing key with fake editor", func(t *testing.T) {
		// Put initial value
		testKey := "/config/test/edit"
		initialValue := "initial value"
		putErr := etcdClient.Put(ctx, testKey, initialValue)
		require.NoError(t, putErr)

		// Create a fake editor script
		fakeEditorPath := filepath.Join(tempDir, "fake-editor.sh")
		editorScript := `#!/bin/bash
echo "modified value" > "$1"
`
		err = os.WriteFile(fakeEditorPath, []byte(editorScript), 0755)
		require.NoError(t, err)

		// Set EDITOR to fake editor
		oldEditor := os.Getenv("EDITOR")
		os.Setenv("EDITOR", fakeEditorPath)
		defer os.Setenv("EDITOR", oldEditor)

		// Run edit command
		err = runEdit(editCmd, []string{testKey})
		require.NoError(t, err)

		// Verify value was updated
		newValue, getErr := etcdClient.Get(ctx, testKey)
		require.NoError(t, getErr)
		require.Equal(t, "modified value\n", newValue)
	})

	t.Run("Edit non-existent key returns error", func(t *testing.T) {
		// Create a no-op fake editor
		fakeEditorPath := filepath.Join(tempDir, "noop-editor.sh")
		editorScript := `#!/bin/bash
exit 0
`
		err = os.WriteFile(fakeEditorPath, []byte(editorScript), 0755)
		require.NoError(t, err)

		oldEditor := os.Getenv("EDITOR")
		os.Setenv("EDITOR", fakeEditorPath)
		defer os.Setenv("EDITOR", oldEditor)

		// Try to edit non-existent key
		testKey := "/config/test/nonexistent"
		err = runEdit(editCmd, []string{testKey})
		require.Error(t, err)
		require.Contains(t, err.Error(), "key not found")
	})

	t.Run("No editor available returns error", func(t *testing.T) {
		// Unset all editor environment variables
		oldEditor := os.Getenv("EDITOR")
		oldVisual := os.Getenv("VISUAL")
		oldPath := os.Getenv("PATH")
		os.Setenv("EDITOR", "")
		os.Setenv("VISUAL", "")
		os.Setenv("PATH", "") // Prevent fallback editors from being found
		defer func() {
			os.Setenv("EDITOR", oldEditor)
			os.Setenv("VISUAL", oldVisual)
			os.Setenv("PATH", oldPath)
		}()

		testKey := "/config/test/edit"
		err := runEdit(editCmd, []string{testKey})
		require.Error(t, err)
		require.Contains(t, err.Error(), "no editor found")
	})

	t.Run("Editor with no changes does not update", func(t *testing.T) {
		// Put initial value
		testKey := "/config/test/nochange"
		initialValue := "unchanged value"
		err := etcdClient.Put(ctx, testKey, initialValue)
		require.NoError(t, err)

		// Create a fake editor that doesn't modify the file
		fakeEditorPath := filepath.Join(tempDir, "nochange-editor.sh")
		editorScript := `#!/bin/bash
# Just open and close without modifying
exit 0
`
		err = os.WriteFile(fakeEditorPath, []byte(editorScript), 0755)
		require.NoError(t, err)

		oldEditor := os.Getenv("EDITOR")
		os.Setenv("EDITOR", fakeEditorPath)
		defer os.Setenv("EDITOR", oldEditor)

		// Run edit command
		err = runEdit(editCmd, []string{testKey})
		require.NoError(t, err)

		// Verify value was not changed
		value, err := etcdClient.Get(ctx, testKey)
		require.NoError(t, err)
		require.Equal(t, initialValue, value)
	})
}

func TestEditCommand_EditorFallback(t *testing.T) {
	// Test that the editor fallback logic works
	oldEditor := os.Getenv("EDITOR")
	oldVisual := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", oldEditor)
		os.Setenv("VISUAL", oldVisual)
	}()

	os.Setenv("EDITOR", "")
	os.Setenv("VISUAL", "")

	// Check if any fallback editor is available
	fallbackEditors := []string{"vi", "vim", "nano", "emacs"}
	var foundEditor string
	for _, editor := range fallbackEditors {
		if _, err := exec.LookPath(editor); err == nil {
			foundEditor = editor
			break
		}
	}

	if foundEditor == "" {
		t.Skip("No fallback editors available on this system")
	}

	t.Logf("Found fallback editor: %s", foundEditor)
}
