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

func TestTruncateForDisplay(t *testing.T) {
	t.Run("short string unchanged", func(t *testing.T) {
		result := truncateForDisplay("short", 50)
		assert.Equal(t, "short", result)
	})

	t.Run("long string truncated", func(t *testing.T) {
		long := strings.Repeat("a", 100)
		result := truncateForDisplay(long, 50)

		assert.Len(t, result, 50)
		assert.True(t, strings.HasSuffix(result, "..."))
	})

	t.Run("newlines escaped", func(t *testing.T) {
		result := truncateForDisplay("line1\nline2", 50)
		assert.Equal(t, "line1\\nline2", result)
	})

	t.Run("exact length unchanged", func(t *testing.T) {
		result := truncateForDisplay("exact", 5)
		assert.Equal(t, "exact", result)
	})

	t.Run("maxLen zero returns empty", func(t *testing.T) {
		result := truncateForDisplay("hello", 0)
		assert.Equal(t, "", result)
	})

	t.Run("negative maxLen returns empty", func(t *testing.T) {
		result := truncateForDisplay("hello", -5)
		assert.Equal(t, "", result)
	})

	t.Run("maxLen 1 returns first rune", func(t *testing.T) {
		result := truncateForDisplay("hello", 1)
		assert.Equal(t, "h", result)
	})

	t.Run("maxLen 2 returns first 2 runes", func(t *testing.T) {
		result := truncateForDisplay("hello", 2)
		assert.Equal(t, "he", result)
	})

	t.Run("maxLen 3 returns first 3 runes no ellipsis", func(t *testing.T) {
		result := truncateForDisplay("hello", 3)
		assert.Equal(t, "hel", result)
	})

	t.Run("maxLen 4 truncates with ellipsis", func(t *testing.T) {
		result := truncateForDisplay("hello", 4)
		assert.Equal(t, "h...", result)
	})

	t.Run("unicode string truncated correctly", func(t *testing.T) {
		result := truncateForDisplay("Hello 世界 🌍!", 10)
		assert.Equal(t, "Hello 世...", result)
		assert.Equal(t, 10, len([]rune(result)))
	})

	t.Run("unicode preserved when under limit", func(t *testing.T) {
		result := truncateForDisplay("世界", 50)
		assert.Equal(t, "世界", result)
	})

	t.Run("unicode exact length preserved", func(t *testing.T) {
		result := truncateForDisplay("Hello 世界 🌍", 10)
		assert.Equal(t, "Hello 世界 🌍", result)
	})

	t.Run("emoji truncated correctly", func(t *testing.T) {
		result := truncateForDisplay("🎉🎊🎁🎄🎅🎆", 5)
		assert.Equal(t, "🎉🎊...", result)
		assert.Equal(t, 5, len([]rune(result)))
	})

	t.Run("mixed content with newlines and unicode", func(t *testing.T) {
		result := truncateForDisplay("Hello\n世界", 15)
		assert.Equal(t, "Hello\\n世界", result)
	})

	t.Run("newline escaped then truncated", func(t *testing.T) {
		result := truncateForDisplay("a\nb\nc\nd\ne", 8)
		assert.Equal(t, "a\\nb\\...", result)
	})
}
