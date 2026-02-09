//go:build integration

package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupEtcdContainer(t *testing.T) string {
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

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err, "failed to get container endpoint")

	return "http://" + endpoint
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
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		assert.NotNil(t, client, "client should not be nil")
	})

	t.Run("Put and Get", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

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
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Put integer value (as string)
		err := client.Put(ctx, "/test/port", "8080")
		require.NoError(t, err, "Put operation should succeed")

		value, err := client.Get(ctx, "/test/port")
		require.NoError(t, err, "Get operation should succeed")
		assert.Equal(t, "8080", value, "string value should match")
	})

	t.Run("PutAll multiple pairs", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/version", Value: "1.0.0"},
			{Key: "/app/port", Value: "8080"},
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
		assert.Equal(t, "8080", port, "app port value should match")
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		_, err := client.Get(ctx, "/nonexistent/key")
		assert.Error(t, err, "Get should fail for non-existent key")
		assert.Contains(t, err.Error(), "not found", "error should indicate key not found")
	})

	t.Run("Status check", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		status, err := client.Status(ctx, endpoint)
		require.NoError(t, err, "Status check should succeed")
		assert.NotNil(t, status, "status response should not be nil")
	})

	t.Run("Put map value via PutAll", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Put map value using PutAll (which formats the value)
		mapValue := map[string]any{
			"en": "Hello",
			"es": "Hola",
			"fr": "Bonjour",
		}

		pairs := []*models.ConfigPair{
			{Key: "/app/greetings", Value: models.FormatValue(mapValue)},
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
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := []*models.ConfigPair{
			{Key: "/types/string", Value: "text"},
			{Key: "/types/int", Value: "42"},
			{Key: "/types/float", Value: "3.14159"},
			{Key: "/types/bool", Value: "true"},
		}

		err := client.PutAll(ctx, pairs)
		require.NoError(t, err, "PutAll with various types should succeed")

		val, err := client.Get(ctx, "/types/string")
		require.NoError(t, err, "Get string value should succeed")
		assert.Equal(t, "text", val, "string value should match")

		val, err = client.Get(ctx, "/types/int")
		require.NoError(t, err, "Get int value should succeed")
		assert.Equal(t, "42", val, "int value should be stored as string")

		val, err = client.Get(ctx, "/types/float")
		require.NoError(t, err, "Get float value should succeed")
		assert.Equal(t, "3.14159", val, "float value should be stored as string")

		val, err = client.Get(ctx, "/types/bool")
		require.NoError(t, err, "Get bool value should succeed")
		assert.Equal(t, "true", val, "bool value should be stored as string")
	})

	t.Run("Close client", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

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
		endpoint := setupEtcdContainer(t)

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

// TestGetWithOptions_Integration tests the GetWithOptions function with various options
// against a real etcd instance. These tests verify that buildClientOptions correctly
// translates GetOptions into clientv3.OpOption behaviors.
func TestGetWithOptions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Prefix returns matching keys", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create keys with different prefixes
		require.NoError(t, client.Put(ctx, "/app/name", "myapp"))
		require.NoError(t, client.Put(ctx, "/app/version", "1.0.0"))
		require.NoError(t, client.Put(ctx, "/app/port", "8080"))
		require.NoError(t, client.Put(ctx, "/other/key", "value"))

		// Test: get all keys with /app/ prefix
		resp, err := client.GetWithOptions(ctx, "/app/", &GetOptions{Prefix: true})
		require.NoError(t, err)

		assert.Len(t, resp.Kvs, 3, "should return exactly 3 keys with /app/ prefix")

		// Verify all returned keys have the correct prefix
		for _, kv := range resp.Kvs {
			assert.True(t, len(kv.Key) >= 5 && kv.Key[:5] == "/app/",
				"key %s should have /app/ prefix", kv.Key)
		}
	})

	t.Run("Limit restricts result count", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create multiple keys
		for i := 0; i < 10; i++ {
			key := "/limit/key" + string(rune('0'+i))
			require.NoError(t, client.Put(ctx, key, "value"))
		}

		// Test: get with limit of 3
		resp, err := client.GetWithOptions(ctx, "/limit/", &GetOptions{
			Prefix: true,
			Limit:  3,
		})
		require.NoError(t, err)

		assert.Len(t, resp.Kvs, 3, "should return exactly 3 keys due to limit")
		assert.True(t, resp.More, "More flag should be true when more keys exist")
	})

	t.Run("Sort order ASCEND by KEY", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create keys in random order
		require.NoError(t, client.Put(ctx, "/sort/zebra", "z"))
		require.NoError(t, client.Put(ctx, "/sort/apple", "a"))
		require.NoError(t, client.Put(ctx, "/sort/mango", "m"))

		// Test: get with ascending sort by key
		resp, err := client.GetWithOptions(ctx, "/sort/", &GetOptions{
			Prefix:     true,
			SortOrder:  "ASCEND",
			SortTarget: "KEY",
		})
		require.NoError(t, err)

		require.Len(t, resp.Kvs, 3, "should return all 3 keys")
		assert.Equal(t, "/sort/apple", resp.Kvs[0].Key, "first key should be apple")
		assert.Equal(t, "/sort/mango", resp.Kvs[1].Key, "second key should be mango")
		assert.Equal(t, "/sort/zebra", resp.Kvs[2].Key, "third key should be zebra")
	})

	t.Run("Sort order DESCEND by KEY", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create keys
		require.NoError(t, client.Put(ctx, "/desc/a", "1"))
		require.NoError(t, client.Put(ctx, "/desc/b", "2"))
		require.NoError(t, client.Put(ctx, "/desc/c", "3"))

		// Test: get with descending sort by key
		resp, err := client.GetWithOptions(ctx, "/desc/", &GetOptions{
			Prefix:     true,
			SortOrder:  "DESCEND",
			SortTarget: "KEY",
		})
		require.NoError(t, err)

		require.Len(t, resp.Kvs, 3, "should return all 3 keys")
		assert.Equal(t, "/desc/c", resp.Kvs[0].Key, "first key should be c (descending)")
		assert.Equal(t, "/desc/b", resp.Kvs[1].Key, "second key should be b")
		assert.Equal(t, "/desc/a", resp.Kvs[2].Key, "third key should be a")
	})

	t.Run("KeysOnly returns empty values", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create keys with values
		require.NoError(t, client.Put(ctx, "/keysonly/a", "value-a"))
		require.NoError(t, client.Put(ctx, "/keysonly/b", "value-b"))

		// Test: get with KeysOnly option
		resp, err := client.GetWithOptions(ctx, "/keysonly/", &GetOptions{
			Prefix:   true,
			KeysOnly: true,
		})
		require.NoError(t, err)

		assert.Len(t, resp.Kvs, 2, "should return 2 keys")
		for _, kv := range resp.Kvs {
			assert.Empty(t, kv.Value, "value should be empty when KeysOnly is true")
			assert.NotEmpty(t, kv.Key, "key should not be empty")
		}
	})

	t.Run("CountOnly returns count without values", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create 5 keys
		for i := 0; i < 5; i++ {
			key := "/count/key" + string(rune('0'+i))
			require.NoError(t, client.Put(ctx, key, "value"))
		}

		// Test: get with CountOnly option
		resp, err := client.GetWithOptions(ctx, "/count/", &GetOptions{
			Prefix:    true,
			CountOnly: true,
		})
		require.NoError(t, err)

		assert.Equal(t, int64(5), resp.Count, "count should be 5")
		assert.Empty(t, resp.Kvs, "Kvs should be empty when CountOnly is true")
	})

	t.Run("FromKey returns keys greater than or equal", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create keys
		require.NoError(t, client.Put(ctx, "/fromkey/a", "1"))
		require.NoError(t, client.Put(ctx, "/fromkey/b", "2"))
		require.NoError(t, client.Put(ctx, "/fromkey/c", "3"))
		require.NoError(t, client.Put(ctx, "/fromkey/d", "4"))

		// Test: get keys starting from /fromkey/b
		resp, err := client.GetWithOptions(ctx, "/fromkey/b", &GetOptions{
			FromKey: true,
		})
		require.NoError(t, err)

		// Should return /fromkey/b, /fromkey/c, /fromkey/d (and potentially more keys in etcd)
		assert.GreaterOrEqual(t, len(resp.Kvs), 3, "should return at least 3 keys >= /fromkey/b")

		// First key should be /fromkey/b
		found := false
		for _, kv := range resp.Kvs {
			if kv.Key == "/fromkey/b" {
				found = true
				break
			}
		}
		assert.True(t, found, "result should include /fromkey/b")
	})

	t.Run("RangeEnd limits key range", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create keys
		require.NoError(t, client.Put(ctx, "/range/a", "1"))
		require.NoError(t, client.Put(ctx, "/range/b", "2"))
		require.NoError(t, client.Put(ctx, "/range/c", "3"))
		require.NoError(t, client.Put(ctx, "/range/d", "4"))

		// Test: get keys in range [/range/a, /range/c) - excludes /range/c and beyond
		resp, err := client.GetWithOptions(ctx, "/range/a", &GetOptions{
			RangeEnd: "/range/c",
		})
		require.NoError(t, err)

		assert.Len(t, resp.Kvs, 2, "should return exactly 2 keys in range [a, c)")
		keys := make([]string, len(resp.Kvs))
		for i, kv := range resp.Kvs {
			keys[i] = kv.Key
		}
		assert.Contains(t, keys, "/range/a", "should contain /range/a")
		assert.Contains(t, keys, "/range/b", "should contain /range/b")
	})

	t.Run("Combined options: Prefix + Limit + Sort", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Setup: create many keys
		require.NoError(t, client.Put(ctx, "/combined/z", "last"))
		require.NoError(t, client.Put(ctx, "/combined/a", "first"))
		require.NoError(t, client.Put(ctx, "/combined/m", "middle"))
		require.NoError(t, client.Put(ctx, "/combined/b", "second"))
		require.NoError(t, client.Put(ctx, "/combined/y", "almost-last"))

		// Test: get top 3 keys sorted ascending
		resp, err := client.GetWithOptions(ctx, "/combined/", &GetOptions{
			Prefix:     true,
			Limit:      3,
			SortOrder:  "ASCEND",
			SortTarget: "KEY",
		})
		require.NoError(t, err)

		assert.Len(t, resp.Kvs, 3, "should return exactly 3 keys")
		assert.Equal(t, "/combined/a", resp.Kvs[0].Key, "first should be /combined/a")
		assert.Equal(t, "/combined/b", resp.Kvs[1].Key, "second should be /combined/b")
		assert.Equal(t, "/combined/m", resp.Kvs[2].Key, "third should be /combined/m")
		assert.True(t, resp.More, "More should be true")
	})
}

func TestPutAllWithProgress_BatchOperations_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("empty batch", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := []*models.ConfigPair{}
		result, err := client.PutAllWithProgress(ctx, pairs, nil)

		require.NoError(t, err)
		assert.Equal(t, 0, result.Total)
		assert.Equal(t, 0, result.Succeeded)
		assert.Equal(t, 0, result.Failed)
	})

	t.Run("small batch within single transaction", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := generateTestPairs("/small", 10)
		result, err := client.PutAllWithProgress(ctx, pairs, nil)

		require.NoError(t, err)
		assert.Equal(t, 10, result.Total)
		assert.Equal(t, 10, result.Succeeded)
		assert.Equal(t, 0, result.Failed)

		for _, pair := range pairs {
			val, err := client.Get(ctx, pair.Key)
			require.NoError(t, err)
			assert.Equal(t, pair.Value, val)
		}
	})

	t.Run("exact limit batch (128 items)", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := generateTestPairs("/exact", DefaultMaxOpsPerTxn)
		result, err := client.PutAllWithProgress(ctx, pairs, nil)

		require.NoError(t, err)
		assert.Equal(t, DefaultMaxOpsPerTxn, result.Total)
		assert.Equal(t, DefaultMaxOpsPerTxn, result.Succeeded)
		assert.Equal(t, 0, result.Failed)

		val, err := client.Get(ctx, pairs[0].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[0].Value, val)

		val, err = client.Get(ctx, pairs[DefaultMaxOpsPerTxn-1].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[DefaultMaxOpsPerTxn-1].Value, val)
	})

	t.Run("over limit batch (129 items - 2 transactions)", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := generateTestPairs("/over", DefaultMaxOpsPerTxn+1)
		result, err := client.PutAllWithProgress(ctx, pairs, nil)

		require.NoError(t, err)
		assert.Equal(t, DefaultMaxOpsPerTxn+1, result.Total)
		assert.Equal(t, DefaultMaxOpsPerTxn+1, result.Succeeded)
		assert.Equal(t, 0, result.Failed)

		val, err := client.Get(ctx, pairs[0].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[0].Value, val)

		val, err = client.Get(ctx, pairs[DefaultMaxOpsPerTxn].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[DefaultMaxOpsPerTxn].Value, val)
	})

	t.Run("large batch (300 items - 3 transactions)", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := generateTestPairs("/large", 300)
		result, err := client.PutAllWithProgress(ctx, pairs, nil)

		require.NoError(t, err)
		assert.Equal(t, 300, result.Total)
		assert.Equal(t, 300, result.Succeeded)
		assert.Equal(t, 0, result.Failed)

		val, err := client.Get(ctx, pairs[0].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[0].Value, val)

		val, err = client.Get(ctx, pairs[150].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[150].Value, val)

		val, err = client.Get(ctx, pairs[299].Key)
		require.NoError(t, err)
		assert.Equal(t, pairs[299].Value, val)
	})

	t.Run("progress callback is called correctly", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := generateTestPairs("/progress", 10)
		var progressCalls []struct {
			current int
			total   int
			key     string
		}

		onProgress := func(current, total int, key string) {
			progressCalls = append(progressCalls, struct {
				current int
				total   int
				key     string
			}{current, total, key})
		}

		result, err := client.PutAllWithProgress(ctx, pairs, onProgress)

		require.NoError(t, err)
		assert.Equal(t, 10, result.Succeeded)
		assert.Len(t, progressCalls, 10)

		for i, call := range progressCalls {
			assert.Equal(t, i+1, call.current)
			assert.Equal(t, 10, call.total)
			assert.Equal(t, pairs[i].Key, call.key)
		}
	})

	t.Run("progress callback across multiple chunks", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)

		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		pairs := generateTestPairs("/multiprogress", 150)
		callCount := 0

		onProgress := func(current, total int, key string) {
			callCount++
			assert.Equal(t, callCount, current)
			assert.Equal(t, 150, total)
		}

		result, err := client.PutAllWithProgress(ctx, pairs, onProgress)

		require.NoError(t, err)
		assert.Equal(t, 150, result.Succeeded)
		assert.Equal(t, 150, callCount)
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
			expected: "3.14",
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
		value any
		name  string
	}{
		{name: "string", value: "hello world"},
		{name: "int", value: int64(42)},
		{name: "float", value: 3.14159},
		{name: "map", value: map[string]any{"key1": "value1", "key2": "value2"}},
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

func setupTLSEtcdContainer(t *testing.T, requireClientCert bool) (endpoint string, certDir string) {
	t.Helper()

	certDir = generateTLSTestCerts(t)

	ctx := context.Background()

	clientCertAuth := "false"
	if requireClientCert {
		clientCertAuth = "true"
	}

	req := testcontainers.ContainerRequest{
		Image:        "quay.io/coreos/etcd:v3.5.9",
		ExposedPorts: []string{"2379/tcp"},
		Env: map[string]string{
			"ETCD_NAME":                        "test-etcd-tls",
			"ETCD_ADVERTISE_CLIENT_URLS":       "https://0.0.0.0:2379",
			"ETCD_LISTEN_CLIENT_URLS":          "https://0.0.0.0:2379",
			"ETCD_CERT_FILE":                   "/certs/server.crt",
			"ETCD_KEY_FILE":                    "/certs/server.key",
			"ETCD_TRUSTED_CA_FILE":             "/certs/ca.crt",
			"ETCD_CLIENT_CERT_AUTH":            clientCertAuth,
			"ETCD_INITIAL_ADVERTISE_PEER_URLS": "http://0.0.0.0:2380",
			"ETCD_LISTEN_PEER_URLS":            "http://0.0.0.0:2380",
			"ETCD_INITIAL_CLUSTER":             "test-etcd-tls=http://0.0.0.0:2380",
		},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(certDir, "/certs"),
		),
		WaitingFor: wait.ForLog("ready to serve client requests").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start TLS etcd container")

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	ep, err := container.Endpoint(ctx, "")
	require.NoError(t, err, "failed to get container endpoint")

	return "https://" + ep, certDir
}

func generateTLSTestCerts(t *testing.T) string {
	t.Helper()
	certDir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caCert, err := x509.ParseCertificate(caCertDER)
	require.NoError(t, err)

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "ca.crt"), caCertPEM, 0644))

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "etcd-server"},
		NotBefore:    time.Now().Add(-5 * time.Minute),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "etcd-server"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("0.0.0.0")},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	require.NoError(t, err)

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "server.crt"), serverCertPEM, 0644))

	serverKeyDER, err := x509.MarshalECPrivateKey(serverKey)
	require.NoError(t, err)
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "server.key"), serverKeyPEM, 0600))

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "etcd-client"},
		NotBefore:    time.Now().Add(-5 * time.Minute),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	require.NoError(t, err)

	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "client.crt"), clientCertPEM, 0644))

	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	require.NoError(t, err)
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "client.key"), clientKeyPEM, 0600))

	return certDir
}

func TestClient_TLS_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TLS integration test in short mode")
	}

	t.Run("TLS with CA verification", func(t *testing.T) {
		endpoint, certDir := setupTLSEtcdContainer(t, false)

		cfg := &Config{
			Endpoints:             []string{endpoint},
			DialTimeout:           10 * time.Second,
			CACert:                filepath.Join(certDir, "ca.crt"),
			Cert:                  filepath.Join(certDir, "client.crt"),
			Key:                   filepath.Join(certDir, "client.key"),
			InsecureSkipTLSVerify: false,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		ctx := testContext(t)
		err = client.Put(ctx, "/tls-test/key1", "value1")
		require.NoError(t, err)

		value, err := client.Get(ctx, "/tls-test/key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("TLS with insecure skip verify", func(t *testing.T) {
		endpoint, certDir := setupTLSEtcdContainer(t, false)

		cfg := &Config{
			Endpoints:             []string{endpoint},
			DialTimeout:           10 * time.Second,
			Cert:                  filepath.Join(certDir, "client.crt"),
			Key:                   filepath.Join(certDir, "client.key"),
			InsecureSkipTLSVerify: true,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		ctx := testContext(t)
		err = client.Put(ctx, "/tls-insecure-test/key1", "value1")
		require.NoError(t, err)

		value, err := client.Get(ctx, "/tls-insecure-test/key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("mTLS with client cert", func(t *testing.T) {
		endpoint, certDir := setupTLSEtcdContainer(t, true)

		cfg := &Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 10 * time.Second,
			CACert:      filepath.Join(certDir, "ca.crt"),
			Cert:        filepath.Join(certDir, "client.crt"),
			Key:         filepath.Join(certDir, "client.key"),
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		ctx := testContext(t)
		err = client.Put(ctx, "/mtls-test/key1", "value1")
		require.NoError(t, err)

		value, err := client.Get(ctx, "/mtls-test/key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("TLS fails with wrong CA cert", func(t *testing.T) {
		endpoint, _ := setupTLSEtcdContainer(t, false)

		wrongCertDir := generateTLSTestCerts(t)

		cfg := &Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
			CACert:      filepath.Join(wrongCertDir, "ca.crt"),
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = client.Put(ctx, "/wrong-ca-test/key1", "value1")
		assert.Error(t, err)
	})

	t.Run("mTLS fails without client cert when required", func(t *testing.T) {
		endpoint, certDir := setupTLSEtcdContainer(t, true)

		cfg := &Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
			CACert:      filepath.Join(certDir, "ca.crt"),
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = client.Put(ctx, "/no-client-cert-test/key1", "value1")
		assert.Error(t, err)
	})
}
