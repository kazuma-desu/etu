package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestLoginCommand_Integration(t *testing.T) {
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

	t.Run("Login with valid endpoint", func(_ *testing.T) {
		loginEndpoints = []string{fullEndpoint}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = false

		runLogin(loginCmd, []string{"test-context"})

		// Verify context was created
		// Note: This will print output, which is expected
	})

	t.Run("Login with no-test flag", func(_ *testing.T) {
		loginEndpoints = []string{fullEndpoint}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = true

		runLogin(loginCmd, []string{"notest-context"})

		// Should succeed even without testing connection
	})

	t.Run("Login with multiple endpoints", func(_ *testing.T) {
		loginEndpoints = []string{fullEndpoint, fullEndpoint}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = true

		runLogin(loginCmd, []string{"multi-context"})

		// Should handle multiple endpoints
	})
}
