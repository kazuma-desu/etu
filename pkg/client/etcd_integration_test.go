package client

import (
	"context"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupEtcdContainer creates an isolated etcd container for testing
func setupEtcdContainer(t *testing.T) (string, func()) {
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
	require.NoError(t, err, "failed to start etcd container")

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err, "failed to get container endpoint")

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	// Wait for etcd to be fully ready
	time.Sleep(2 * time.Second)

	return "http://" + endpoint, cleanup
}

// newTestClient creates a test client with proper cleanup
func newTestClient(t *testing.T, endpoint string) *Client {
	t.Helper()

	cfg := &Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err, "failed to create client")

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("failed to close client: %v", err)
		}
	})

	return client
}

// testContext creates a context with timeout for tests
func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("NewClient", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		assert.NotNil(t, client, "client should not be nil")
	})

	t.Run("Put and Get", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Put a key-value pair
		err := client.Put(ctx, "/test/key", "test-value")
		require.NoError(t, err, "Put operation should succeed")

		// Get the value back
		value, err := client.Get(ctx, "/test/key")
		require.NoError(t, err, "Get operation should succeed")
		assert.Equal(t, "test-value", value, "retrieved value should match")
	})

	t.Run("Put integer value", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Put integer value (as string)
		err := client.Put(ctx, "/test/port", "8080")
		require.NoError(t, err, "Put operation should succeed")

		// Get the value back
		value, err := client.Get(ctx, "/test/port")
		require.NoError(t, err, "Get operation should succeed")
		assert.Equal(t, "8080", value, "integer value should be stored as string")
	})

	t.Run("PutAll multiple pairs", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/version", Value: "1.0.0"},
			{Key: "/app/port", Value: int64(8080)},
		}

		err := client.PutAll(ctx, pairs)
		require.NoError(t, err, "PutAll operation should succeed")

		// Verify all values were set
		name, err := client.Get(ctx, "/app/name")
		require.NoError(t, err, "Get name should succeed")
		assert.Equal(t, "myapp", name, "app name should match")

		version, err := client.Get(ctx, "/app/version")
		require.NoError(t, err, "Get version should succeed")
		assert.Equal(t, "1.0.0", version, "app version should match")

		port, err := client.Get(ctx, "/app/port")
		require.NoError(t, err, "Get port should succeed")
		assert.Equal(t, "8080", port, "app port should match")
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		_, err := client.Get(ctx, "/nonexistent/key")
		assert.Error(t, err, "Get should fail for non-existent key")
		assert.Contains(t, err.Error(), "not found", "error should indicate key not found")
	})

	t.Run("Status check", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		status, err := client.Status(ctx, endpoint)
		require.NoError(t, err, "Status check should succeed")
		assert.NotNil(t, status, "status response should not be nil")
	})

	t.Run("Put map value via PutAll", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Put map value using PutAll (which formats the value)
		mapValue := map[string]any{
			"en": "Hello",
			"es": "Hola",
			"fr": "Bonjour",
		}

		pairs := []*models.ConfigPair{
			{Key: "/app/greetings", Value: mapValue},
		}

		err := client.PutAll(ctx, pairs)
		require.NoError(t, err, "PutAll with map value should succeed")

		// Get it back
		value, err := client.Get(ctx, "/app/greetings")
		require.NoError(t, err, "Get map value should succeed")
		// Should contain the formatted values
		assert.Contains(t, value, "Hello", "value should contain 'Hello'")
		assert.Contains(t, value, "Hola", "value should contain 'Hola'")
	})

	t.Run("PutAll with different value types", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := []*models.ConfigPair{
			{Key: "/types/string", Value: "text"},
			{Key: "/types/int", Value: int64(42)},
			{Key: "/types/float", Value: 3.14159},
			{Key: "/types/bool", Value: true},
		}

		err := client.PutAll(ctx, pairs)
		require.NoError(t, err, "PutAll with various types should succeed")

		// Verify all were set correctly
		val, err := client.Get(ctx, "/types/string")
		require.NoError(t, err, "Get string value should succeed")
		assert.Equal(t, "text", val, "string value should match")

		val, err = client.Get(ctx, "/types/int")
		require.NoError(t, err, "Get int value should succeed")
		assert.Equal(t, "42", val, "int value should be formatted as string")

		val, err = client.Get(ctx, "/types/float")
		require.NoError(t, err, "Get float value should succeed")
		assert.Contains(t, val, "3.14", "float value should be present")

		val, err = client.Get(ctx, "/types/bool")
		require.NoError(t, err, "Get bool value should succeed")
		assert.Equal(t, "true", val, "bool value should be formatted as string")
	})

	t.Run("Close client", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		cfg := &Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err, "NewClient should succeed")

		// Close should succeed
		err = client.Close()
		require.NoError(t, err, "Close should succeed")
	})

	t.Run("Client with authentication (no auth server)", func(t *testing.T) {
		endpoint, cleanup := setupEtcdContainer(t)
		defer cleanup()

		cfg := &Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
			Username:    "testuser",
			Password:    "testpass",
		}

		client, err := NewClient(cfg)
		require.NoError(t, err, "NewClient with auth should succeed")
		defer client.Close()

		ctx := testContext(t)

		// Should connect (auth will fail on operations if server requires it)
		err = client.Put(ctx, "/auth/test", "value")
		// May succeed or fail depending on etcd auth setup
		// Just verify we can create the client
		_ = err
	})
}

func TestClient_ConnectionFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Invalid endpoint", func(t *testing.T) {
		cfg := &Config{
			Endpoints:   []string{"http://invalid-host:2379"},
			DialTimeout: 1 * time.Second,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err, "Client creation should succeed") // Client creation succeeds
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Operations should fail
		err = client.Put(ctx, "/test/key", "value")
		assert.Error(t, err, "Put should fail with invalid endpoint")

		// Get should also fail
		_, err = client.Get(ctx, "/test/key")
		assert.Error(t, err, "Get should fail with invalid endpoint")
	})

	t.Run("Empty endpoints", func(t *testing.T) {
		cfg := &Config{
			Endpoints:   []string{},
			DialTimeout: 1 * time.Second,
		}

		_, err := NewClient(cfg)
		assert.Error(t, err, "NewClient should fail with empty endpoints")
		assert.Contains(t, err.Error(), "at least one endpoint is required", "error message should be clear")
	})
}

func TestFormatValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "string value",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "integer value",
			value:    int64(42),
			expected: "42",
		},
		{
			name:     "float value",
			value:    3.14,
			expected: "3.140000",
		},
		{
			name:     "bool value",
			value:    true,
			expected: "true",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatValue(tt.value)
			assert.Equal(t, tt.expected, result, "formatted value should match expected")
		})
	}

	// Test map value separately as order is not guaranteed
	t.Run("map value", func(t *testing.T) {
		t.Parallel()
		mapValue := map[string]any{
			"key": "value",
		}
		result := formatValue(mapValue)
		assert.Contains(t, result, "key: value", "map should be formatted correctly")
	})

	// Test slice value
	t.Run("slice value", func(t *testing.T) {
		t.Parallel()
		sliceValue := []string{"item1", "item2"}
		result := formatValue(sliceValue)
		assert.Contains(t, result, "item1", "slice should contain item1")
		assert.Contains(t, result, "item2", "slice should contain item2")
	})
}

// Benchmark for formatValue function
func BenchmarkFormatValue(b *testing.B) {
	testCases := []struct {
		name  string
		value any
	}{
		{"string", "hello world"},
		{"int", int64(42)},
		{"float", 3.14159},
		{"map", map[string]any{"key1": "value1", "key2": "value2"}},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				formatValue(tc.value)
			}
		})
	}
}
