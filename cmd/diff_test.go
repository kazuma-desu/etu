package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string",
			input:    "test",
			expected: "test",
		},
		{
			name:     "int",
			input:    123,
			expected: "123",
		},
		{
			name:     "nil",
			input:    nil,
			expected: "",
		},
		{
			name:     "bool",
			input:    true,
			expected: "true",
		},
		{
			name:     "float",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDiffFullFlagRequiresPrefix(t *testing.T) {
	originalOpts := diffOpts
	defer func() { diffOpts = originalOpts }()

	diffOpts.FilePath = "test.txt"
	diffOpts.Full = true
	diffOpts.Prefix = ""

	err := runDiff(nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--full requires --prefix")
}
