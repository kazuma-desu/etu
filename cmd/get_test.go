package cmd

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/testutil"
)

func resetGetOpts() {
	getOpts.sortOrder = ""
	getOpts.sortTarget = ""
	getOpts.consistency = "l"
	getOpts.rangeEnd = ""
	getOpts.limit = 0
	getOpts.revision = 0
	getOpts.minModRev = 0
	getOpts.maxModRev = 0
	getOpts.minCreateRev = 0
	getOpts.maxCreateRev = 0
	getOpts.prefix = false
	getOpts.fromKey = false
	getOpts.keysOnly = false
	getOpts.countOnly = false
	getOpts.printValue = false
	getOpts.showMetadata = false
}

func TestPrintSimple(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "key1", Value: "val1"},
			{Key: "key2", Value: "val2"},
		},
		Count: 2,
	}

	t.Run("default output", func(t *testing.T) {
		getOpts.printValue = false
		getOpts.keysOnly = false
		getOpts.showMetadata = false

		output, err := testutil.CaptureStdout(func() error {
			printSimple(resp)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, "key1\nval1\nkey2\nval2\n", output)
	})

	t.Run("print value only", func(t *testing.T) {
		getOpts.printValue = true
		getOpts.keysOnly = false
		getOpts.showMetadata = false

		output, err := testutil.CaptureStdout(func() error {
			printSimple(resp)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, "val1\nval2\n", output)
	})

	t.Run("keys only", func(t *testing.T) {
		getOpts.printValue = false
		getOpts.keysOnly = true
		getOpts.showMetadata = false

		output, err := testutil.CaptureStdout(func() error {
			printSimple(resp)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, "key1\nkey2\n", output)
	})

	t.Run("with metadata and lease", func(t *testing.T) {
		getOpts.printValue = false
		getOpts.keysOnly = false
		getOpts.showMetadata = true
		respWithLease := &client.GetResponse{
			Kvs: []*client.KeyValue{
				{Key: "key1", Value: "val1", Lease: 123},
			},
			Count: 1,
		}

		output, err := testutil.CaptureStdout(func() error {
			printSimple(respWithLease)
			return nil
		})
		require.NoError(t, err)
		assert.Contains(t, output, "key1")
		assert.Contains(t, output, "val1")
		assert.Contains(t, output, "Lease")
		assert.Contains(t, output, "123")
	})
}

func TestPrintJSON(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{
				Key:            "key1",
				Value:          "val1",
				CreateRevision: 1,
				ModRevision:    2,
				Version:        1,
				Lease:          123,
			},
		},
		Count: 1,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printJSON(resp)
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, float64(1), result["count"])
	kvs := result["kvs"].([]any)
	assert.Len(t, kvs, 1)
	kv := kvs[0].(map[string]any)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("key1")), kv["key"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("val1")), kv["value"])
	assert.Equal(t, float64(123), kv["lease"])
}

func TestPrintTable(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "key1", Value: "val1"},
			{Key: "long-key", Value: strings.Repeat("a", 60)},
		},
		Count: 2,
	}

	t.Run("default table", func(t *testing.T) {
		getOpts.keysOnly = false
		getOpts.showMetadata = false

		output, err := testutil.CaptureStdout(func() error {
			return printTable(resp)
		})
		require.NoError(t, err)
		assert.Contains(t, output, "KEY")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "key1")
		assert.Contains(t, output, "val1")
		assert.Contains(t, output, "aaaaa...")
	})

	t.Run("keys only", func(t *testing.T) {
		getOpts.keysOnly = true
		getOpts.showMetadata = false

		output, err := testutil.CaptureStdout(func() error {
			return printTable(resp)
		})
		require.NoError(t, err)
		assert.Contains(t, output, "KEY")
		assert.NotContains(t, output, "VALUE")
		assert.Contains(t, output, "key1")
	})

	t.Run("with metadata", func(t *testing.T) {
		getOpts.keysOnly = false
		getOpts.showMetadata = true

		output, err := testutil.CaptureStdout(func() error {
			return printTable(resp)
		})
		require.NoError(t, err)
		assert.Contains(t, output, "KEY")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "CREATE_REV")
		assert.Contains(t, output, "MOD_REV")
	})
}

func TestPrintFields(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{
				Key:            "key1",
				Value:          "val1",
				CreateRevision: 10,
				ModRevision:    20,
				Version:        2,
				Lease:          123,
			},
		},
		Count: 1,
	}

	t.Run("default fields", func(t *testing.T) {
		getOpts.keysOnly = false
		output, err := testutil.CaptureStdout(func() error {
			printFields(resp)
			return nil
		})
		require.NoError(t, err)
		assert.Contains(t, output, "key1")
		assert.Contains(t, output, "val1")
		assert.Contains(t, output, "CreateRevision: 10")
		assert.Contains(t, output, "ModRevision: 20")
		assert.Contains(t, output, "Version: 2")
		assert.Contains(t, output, "Lease: 123")
	})

	t.Run("keys only", func(t *testing.T) {
		getOpts.keysOnly = true
		output, err := testutil.CaptureStdout(func() error {
			printFields(resp)
			return nil
		})
		require.NoError(t, err)
		assert.Contains(t, output, "key1")
		assert.NotContains(t, output, "val1")
	})
}

func TestPrintYAML(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/version", Value: "1.0.0"},
			{Key: "/app/empty", Value: ""},
		},
		Count: 3,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printYAML(resp)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "app:")
	assert.Contains(t, output, "name: myapp")
	assert.Contains(t, output, "version:")
	// Keys with empty values should be omitted from YAML output
	assert.NotContains(t, output, "empty:", "empty-value key should be excluded from YAML output")
}

func TestPrintYAML_Error(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app", Value: "val1"},
			{Key: "/app/config", Value: "val2"},
		},
		Count: 2,
	}

	err := printYAML(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unflatten keys")
}

func TestPrintTree(t *testing.T) {
	t.Cleanup(resetGetOpts)
	resetGetOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/config/host", Value: "localhost"},
			{Key: "/app/config/port", Value: "5432"},
		},
		Count: 2,
	}

	t.Run("tree output", func(t *testing.T) {
		getOpts.prefix = true
		getOpts.fromKey = false
		output, err := testutil.CaptureStdout(func() error {
			return printTree(resp)
		})
		require.NoError(t, err)
		assert.Contains(t, output, "app")
		assert.Contains(t, output, "config")
		assert.Contains(t, output, "host")
		assert.Contains(t, output, "localhost")
	})

	t.Run("fallback to table when no prefix/fromKey", func(t *testing.T) {
		getOpts.prefix = false
		getOpts.fromKey = false
		output, err := testutil.CaptureStdout(func() error {
			return printTree(resp)
		})
		require.NoError(t, err)
		assert.Contains(t, output, "KEY")
		assert.Contains(t, output, "VALUE")
	})
}
