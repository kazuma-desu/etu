package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/testutil"
)

func resetLsOpts() {
	lsOpts.prefix = true
}

func TestPrintLsSimple(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/config/host"},
			{Key: "/app/config/port"},
			{Key: "/app/name"},
		},
		Count: 3,
	}

	output, err := testutil.CaptureStdout(func() error {
		printLsSimple(resp)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "/app/config/host\n/app/config/port\n/app/name\n", output)
}

func TestPrintLsJSON(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/config/host"},
			{Key: "/app/config/port"},
		},
		Count: 2,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printLsJSON(resp)
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, float64(2), result["count"])
	keys := result["keys"].([]any)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "/app/config/host")
	assert.Contains(t, keys, "/app/config/port")
}

func TestPrintLsYAML(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/name"},
			{Key: "/app/version"},
		},
		Count: 2,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printLsYAML(resp)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "keys:")
	assert.Contains(t, output, "/app/name")
	assert.Contains(t, output, "/app/version")
	assert.Contains(t, output, "count: 2")
}

func TestPrintLsTable(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/config/host"},
			{Key: "/app/config/port"},
		},
		Count: 2,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printLsTable(resp)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "KEY")
	assert.Contains(t, output, "/app/config/host")
	assert.Contains(t, output, "/app/config/port")
}

func TestPrintLsSimple_Empty(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs:   []*client.KeyValue{},
		Count: 0,
	}

	output, err := testutil.CaptureStdout(func() error {
		printLsSimple(resp)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestPrintLsJSON_Empty(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs:   []*client.KeyValue{},
		Count: 0,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printLsJSON(resp)
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, float64(0), result["count"])
	keys := result["keys"].([]any)
	assert.Len(t, keys, 0)
}

func TestPrintLsSimple_ManyKeys(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs: []*client.KeyValue{
			{Key: "/app/key1"},
			{Key: "/app/key2"},
			{Key: "/app/key3"},
			{Key: "/app/key4"},
			{Key: "/app/key5"},
		},
		Count: 5,
	}

	output, err := testutil.CaptureStdout(func() error {
		printLsSimple(resp)
		return nil
	})
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 5)
	for _, key := range resp.Kvs {
		assert.Contains(t, output, key.Key)
	}
}

func TestPrintLsTable_Empty(t *testing.T) {
	t.Cleanup(resetLsOpts)
	resetLsOpts()
	resp := &client.GetResponse{
		Kvs:   []*client.KeyValue{},
		Count: 0,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printLsTable(resp)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "KEY")
}
