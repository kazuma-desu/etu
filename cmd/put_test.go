package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveValue(t *testing.T) {
	t.Run("value from args", func(t *testing.T) {
		args := []string{"/key", "my-value"}
		value, err := resolveValue(args, nil)

		require.NoError(t, err)
		assert.Equal(t, "my-value", value)
	})

	t.Run("value from stdin when dash provided", func(t *testing.T) {
		args := []string{"/key", "-"}
		stdin := strings.NewReader("stdin-value")
		value, err := resolveValue(args, stdin)

		require.NoError(t, err)
		assert.Equal(t, "stdin-value", value)
	})

	t.Run("value from stdin when no value arg", func(t *testing.T) {
		args := []string{"/key"}
		stdin := strings.NewReader("stdin-value")
		value, err := resolveValue(args, stdin)

		require.NoError(t, err)
		assert.Equal(t, "stdin-value", value)
	})
}

func TestReadValueFromStdin(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		stdin := strings.NewReader("single-line-value")
		value, err := readValueFromStdin(stdin)

		require.NoError(t, err)
		assert.Equal(t, "single-line-value", value)
	})

	t.Run("multi-line preserves newlines", func(t *testing.T) {
		stdin := strings.NewReader("line1\nline2\nline3")
		value, err := readValueFromStdin(stdin)

		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\nline3", value)
	})

	t.Run("empty stdin returns error", func(t *testing.T) {
		stdin := strings.NewReader("")
		_, err := readValueFromStdin(stdin)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty value")
	})

	t.Run("whitespace only is valid", func(t *testing.T) {
		stdin := strings.NewReader("   ")
		value, err := readValueFromStdin(stdin)

		require.NoError(t, err)
		assert.Equal(t, "   ", value)
	})
}

func TestValidateKeyValue(t *testing.T) {
	t.Run("valid key and value", func(t *testing.T) {
		err := validateKeyValue("/app/config/host", "localhost")
		assert.NoError(t, err)
	})

	t.Run("invalid key without slash", func(t *testing.T) {
		err := validateKeyValue("invalid-key", "value")
		assert.Error(t, err)
	})

	t.Run("key with spaces fails", func(t *testing.T) {
		err := validateKeyValue("/app/config/my key", "value")
		assert.Error(t, err)
	})
}
