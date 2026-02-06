package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
	}{
		{
			name: "simple key value",
			input: map[string]any{
				"key": "value",
			},
			contains: []string{"key: value"},
		},
		{
			name: "multi-line string",
			input: map[string]any{
				"config": "line1\nline2",
			},
			contains: []string{
				"config: |",
				"  line1",
				"  line2",
			},
		},
		{
			name: "nested map",
			input: map[string]any{
				"parent": map[string]any{
					"child": "value",
				},
			},
			contains: []string{
				"parent:",
				"  child: value",
			},
		},
		{
			name: "nested multi-line",
			input: map[string]any{
				"scripts": map[string]any{
					"start": "echo hello\necho world",
				},
			},
			contains: []string{
				"scripts:",
				"  start: |",
				"    echo hello",
				"    echo world",
			},
		},
		{
			name: "mixed types",
			input: map[string]any{
				"int":    123,
				"bool":   true,
				"string": "text",
			},
			contains: []string{
				"int: 123",
				"bool: true",
				"string: text",
			},
		},
		{
			name: "slice support",
			input: map[string]any{
				"list": []any{"a", "b\nc"},
			},
			contains: []string{
				"list:",
				"- a",
				"- |",
				"  b",
				"  c",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := SerializeYAML(tt.input)
			require.NoError(t, err)

			strOutput := string(output)
			for _, substr := range tt.contains {
				assert.Contains(t, strOutput, substr)
			}
		})
	}
}

func TestSerializeYAML_Deterministic(t *testing.T) {
	input := map[string]any{
		"z": 1,
		"a": 2,
		"m": 3,
	}

	output, err := SerializeYAML(input)
	require.NoError(t, err)

	// Keys should be sorted
	expected := "a: 2\nm: 3\nz: 1\n"
	assert.Equal(t, expected, string(output))
}

func TestSerializeYAML_EmptyMap(t *testing.T) {
	input := map[string]any{}
	output, err := SerializeYAML(input)
	require.NoError(t, err)
	// Empty map produces "{}\n" in YAML
	assert.Equal(t, "{}\n", string(output))
}
