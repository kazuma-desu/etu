//go:build integration

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupEtcdContainerForCmd(t *testing.T) string {
	t.Helper()
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

	// Use t.Cleanup for proper test lifecycle integration
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	return "http://" + endpoint
}

func TestApplyCommand_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint := setupEtcdContainerForCmd(t)

	t.Run("Apply with valid etcdctl format", func(t *testing.T) {
		// Create a temporary config file
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.txt")

		content := `/test/app/name
integration-test

/test/app/version
1.0.0

/test/app/port
8080
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		// Isolate config to prevent loading user's config
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", oldHome)

		// Use environment variable for etcd endpoint
		oldEndpoints := os.Getenv("ETCD_ENDPOINTS")
		os.Setenv("ETCD_ENDPOINTS", endpoint)
		defer os.Setenv("ETCD_ENDPOINTS", oldEndpoints)

		// Run apply command
		applyOpts.FilePath = configFile
		applyOpts.Format = "etcdctl"
		applyOpts.DryRun = false
		applyOpts.NoValidate = false
		applyOpts.Strict = false

		err = runApply(applyCmd, []string{})
		require.NoError(t, err)

		// Verify data was written to etcd
		cfg := &client.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
		}
		etcdClient, err := client.NewClient(cfg)
		require.NoError(t, err)
		defer etcdClient.Close()

		ctx := context.Background()
		value, err := etcdClient.Get(ctx, "/test/app/name")
		require.NoError(t, err)
		assert.Equal(t, "integration-test", value)

		value, err = etcdClient.Get(ctx, "/test/app/version")
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", value)

		value, err = etcdClient.Get(ctx, "/test/app/port")
		require.NoError(t, err)
		assert.Equal(t, "8080", value)
	})

	t.Run("Apply with dry-run", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "dryrun.txt")

		content := `/dryrun/key
should-not-be-written
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		oldEndpoints := os.Getenv("ETCD_ENDPOINTS")
		os.Setenv("ETCD_ENDPOINTS", endpoint)
		defer os.Setenv("ETCD_ENDPOINTS", oldEndpoints)

		// Run with dry-run
		applyOpts.FilePath = configFile
		applyOpts.Format = "etcdctl"
		applyOpts.DryRun = true
		applyOpts.NoValidate = false
		applyOpts.Strict = false

		err = runApply(applyCmd, []string{})
		require.NoError(t, err)

		// Verify data was NOT written to etcd
		cfg := &client.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
		}
		etcdClient, err := client.NewClient(cfg)
		require.NoError(t, err)
		defer etcdClient.Close()

		ctx := context.Background()
		_, err = etcdClient.Get(ctx, "/dryrun/key")
		assert.Error(t, err) // Should not exist
	})

	t.Run("Apply with invalid file", func(t *testing.T) {
		oldEndpoints := os.Getenv("ETCD_ENDPOINTS")
		os.Setenv("ETCD_ENDPOINTS", endpoint)
		defer os.Setenv("ETCD_ENDPOINTS", oldEndpoints)

		applyOpts.FilePath = "/nonexistent/file.txt"
		applyOpts.Format = "etcdctl"
		applyOpts.DryRun = false
		applyOpts.NoValidate = false
		applyOpts.Strict = false

		err := runApply(applyCmd, []string{})
		assert.Error(t, err)
	})

	t.Run("Apply with auto-detect format", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "auto.txt")

		content := `/auto/detect/key
value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		// Isolate config to prevent loading user's config
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", oldHome)

		oldEndpoints := os.Getenv("ETCD_ENDPOINTS")
		os.Setenv("ETCD_ENDPOINTS", endpoint)
		defer os.Setenv("ETCD_ENDPOINTS", oldEndpoints)

		// Run with auto format detection
		applyOpts.FilePath = configFile
		applyOpts.Format = "" // Auto-detect
		applyOpts.DryRun = false
		applyOpts.NoValidate = false
		applyOpts.Strict = false

		err = runApply(applyCmd, []string{})
		require.NoError(t, err)

		// Verify it was applied
		cfg := &client.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
		}
		etcdClient, err := client.NewClient(cfg)
		require.NoError(t, err)
		defer etcdClient.Close()

		ctx := context.Background()
		value, err := etcdClient.Get(ctx, "/auto/detect/key")
		require.NoError(t, err)
		assert.Equal(t, "value", value)
	})

	t.Run("Apply with no-validate flag", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "novalidate.txt")

		content := `/novalidate/key
value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		// Isolate config to prevent loading user's config
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", oldHome)

		oldEndpoints := os.Getenv("ETCD_ENDPOINTS")
		os.Setenv("ETCD_ENDPOINTS", endpoint)
		defer os.Setenv("ETCD_ENDPOINTS", oldEndpoints)

		applyOpts.FilePath = configFile
		applyOpts.Format = "etcdctl"
		applyOpts.DryRun = false
		applyOpts.NoValidate = true // Skip validation
		applyOpts.Strict = false

		err = runApply(applyCmd, []string{})
		require.NoError(t, err)

		// Verify it was applied
		cfg := &client.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
		}
		etcdClient, err := client.NewClient(cfg)
		require.NoError(t, err)
		defer etcdClient.Close()

		ctx := context.Background()
		value, err := etcdClient.Get(ctx, "/novalidate/key")
		require.NoError(t, err)
		assert.Equal(t, "value", value)
	})
}
