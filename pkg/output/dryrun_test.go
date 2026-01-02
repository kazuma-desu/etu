package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
)

func captureOutputWithError(f func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

func TestPrintDryRunOperations(t *testing.T) {
	ops := []client.Operation{
		{Type: "PUT", Key: "/key1", Value: "value1"},
		{Type: "DELETE", Key: "/key2"},
	}

	t.Run("json format", func(t *testing.T) {
		output, err := captureOutputWithError(func() error {
			return PrintDryRunOperations(ops, "json")
		})

		assert.NoError(t, err)

		var decoded []client.Operation
		err = json.Unmarshal([]byte(output), &decoded)
		require.NoError(t, err)

		assert.Equal(t, ops, decoded)
	})

	t.Run("simple format", func(t *testing.T) {
		output, err := captureOutputWithError(func() error {
			return PrintDryRunOperations(ops, "simple")
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "DRY RUN - Would perform 2 operations")
		assert.Contains(t, output, "PUT → /key1")
		assert.Contains(t, output, "value1")
		assert.Contains(t, output, "DELETE → /key2")
	})

	t.Run("table format (alias for simple)", func(t *testing.T) {
		output, err := captureOutputWithError(func() error {
			return PrintDryRunOperations(ops, "table")
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "DRY RUN - Would perform 2 operations")
	})

	t.Run("invalid format", func(t *testing.T) {
		err := PrintDryRunOperations(ops, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})

	t.Run("empty operations", func(t *testing.T) {
		output, err := captureOutputWithError(func() error {
			return PrintDryRunOperations([]client.Operation{}, "simple")
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Would perform 0 operations")
	})
}
