package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/testutil"
)

func TestPrintDryRunOperations(t *testing.T) {
	ops := []DryRunOperation{
		{Type: "PUT", Key: "/key1", Value: "value1"},
		{Type: "DELETE", Key: "/key2"},
	}

	t.Run("json format", func(t *testing.T) {
		output, err := testutil.CaptureStdout(func() error {
			return PrintDryRunOperations(ops, "json")
		})

		assert.NoError(t, err)

		var decoded []DryRunOperation
		err = json.Unmarshal([]byte(output), &decoded)
		require.NoError(t, err)

		assert.Equal(t, ops, decoded)
	})

	t.Run("simple format", func(t *testing.T) {
		output, err := testutil.CaptureStdout(func() error {
			return PrintDryRunOperations(ops, "simple")
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "DRY RUN - Would perform 2 operations")
		assert.Contains(t, output, "PUT → /key1")
		assert.Contains(t, output, "value1")
		assert.Contains(t, output, "DELETE → /key2")
	})

	t.Run("table format (alias for simple)", func(t *testing.T) {
		output, err := testutil.CaptureStdout(func() error {
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
		output, err := testutil.CaptureStdout(func() error {
			return PrintDryRunOperations([]DryRunOperation{}, "simple")
		})

		assert.NoError(t, err)
		assert.Contains(t, output, "Would perform 0 operations")
	})
}
