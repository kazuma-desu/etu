package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string unchanged",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length unchanged",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string truncated",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "maxLen zero returns empty",
			input:    "hello",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "negative maxLen returns empty",
			input:    "hello",
			maxLen:   -5,
			expected: "",
		},
		{
			name:     "maxLen 1 returns first rune",
			input:    "hello",
			maxLen:   1,
			expected: "h",
		},
		{
			name:     "maxLen 2 returns first 2 runes",
			input:    "hello",
			maxLen:   2,
			expected: "he",
		},
		{
			name:     "maxLen 3 returns first 3 runes no ellipsis",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
		{
			name:     "maxLen 4 truncates with ellipsis",
			input:    "hello",
			maxLen:   4,
			expected: "h...",
		},
		{
			name:     "unicode string truncated correctly",
			input:    "Hello 世界 🌍!",
			maxLen:   10,
			expected: "Hello 世...",
		},
		{
			name:     "unicode preserved when under limit",
			input:    "世界",
			maxLen:   50,
			expected: "世界",
		},
		{
			name:     "unicode exact length preserved",
			input:    "Hello 世界 🌍",
			maxLen:   10,
			expected: "Hello 世界 🌍",
		},
		{
			name:     "emoji truncated correctly",
			input:    "🎉🎊🎁🎄🎅🎆",
			maxLen:   5,
			expected: "🎉🎊...",
		},
		{
			name:     "newlines escaped",
			input:    "line1\nline2",
			maxLen:   50,
			expected: "line1\\nline2",
		},
		{
			name:     "mixed content with newlines and unicode",
			input:    "Hello\n世界",
			maxLen:   15,
			expected: "Hello\\n世界",
		},
		{
			name:     "newline escaped then truncated",
			input:    "a\nb\nc\nd\ne",
			maxLen:   8,
			expected: "a\\nb\\...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			if tt.maxLen > 0 {
				// Verify rune length constraint is respected
				assert.LessOrEqual(t, len([]rune(result)), tt.maxLen)
			}
		})
	}
}

func TestTruncate_LongString(t *testing.T) {
	long := strings.Repeat("a", 100)
	result := Truncate(long, 50)

	assert.Equal(t, 50, len([]rune(result)))
	assert.True(t, strings.HasSuffix(result, "..."))
}
