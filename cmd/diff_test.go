package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kazuma-desu/etu/pkg/models"
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
			got := models.FormatValue(tt.input)
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

	err := runDiff(diffCmd, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--full requires --prefix")
}

func TestDiffDeprecatedFormatFlag(t *testing.T) {
	originalOpts := diffOpts
	defer func() { diffOpts = originalOpts }()

	diffOpts.FilePath = "test.txt"
	diffOpts.Format = "simple"
	diffOpts.DeprecatedFormat = "json"

	var stderr strings.Builder
	diffCmd.SetErr(&stderr)

	_ = runDiff(diffCmd, nil)

	assert.Contains(t, stderr.String(), "deprecated")
}
