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

func TestSerializeYAML_NonSerializableType(t *testing.T) {
	// yaml.v3 panics on channels, so we need to recover
	data := map[string]any{
		"channel": make(chan int),
	}

	defer func() {
		if r := recover(); r != nil {
			// Expected panic for non-serializable type
			t.Log("Correctly panicked on channel type:", r)
		}
	}()

	_, err := SerializeYAML(data)
	if err != nil {
		// If it returns error instead of panic, that's also acceptable
		assert.Contains(t, err.Error(), "failed to convert to YAML node")
	}
}

func TestSerializeYAML_FuncType(t *testing.T) {
	// yaml.v3 panics on functions, so we need to recover
	data := map[string]any{
		"func": func() {},
	}

	defer func() {
		if r := recover(); r != nil {
			// Expected panic for non-serializable type
			t.Log("Correctly panicked on function type:", r)
		}
	}()

	_, err := SerializeYAML(data)
	if err != nil {
		// If it returns error instead of panic, that's also acceptable
		assert.Contains(t, err.Error(), "failed to convert to YAML node")
	}
}

func TestSerializeYAML_NumericBoolHeuristic(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name: "integer-looking string renders unquoted",
			input: map[string]any{
				"port": "8080",
			},
			expected: "port: 8080\n",
		},
		{
			name: "negative integer-looking string renders unquoted",
			input: map[string]any{
				"offset": "-42",
			},
			expected: "offset: -42\n",
		},
		{
			name: "float-looking string renders unquoted",
			input: map[string]any{
				"rate": "3.14",
			},
			expected: "rate: 3.14\n",
		},
		{
			name: "scientific notation float renders unquoted",
			input: map[string]any{
				"value": "1.5e10",
			},
			expected: "value: 1.5e10\n",
		},
		{
			name: "bool-looking string true renders unquoted",
			input: map[string]any{
				"enabled": "true",
			},
			expected: "enabled: true\n",
		},
		{
			name: "bool-looking string false renders unquoted",
			input: map[string]any{
				"disabled": "false",
			},
			expected: "disabled: false\n",
		},
		{
			name: "regular string renders normally",
			input: map[string]any{
				"name": "hello",
			},
			expected: "name: hello\n",
		},
		{
			name: "YAML special value null is quoted",
			input: map[string]any{
				"value": "null",
			},
			expected: "value: \"null\"\n",
		},
		{
			name: "YAML special value tilde is quoted",
			input: map[string]any{
				"value": "~",
			},
			expected: "value: \"~\"\n",
		},
		{
			name: "YAML special value yes is quoted",
			input: map[string]any{
				"value": "yes",
			},
			expected: "value: \"yes\"\n",
		},
		{
			name: "YAML special value no is quoted",
			input: map[string]any{
				"value": "no",
			},
			expected: "value: \"no\"\n",
		},
		{
			name: "YAML special value on is quoted",
			input: map[string]any{
				"value": "on",
			},
			expected: "value: \"on\"\n",
		},
		{
			name: "YAML special value off is quoted",
			input: map[string]any{
				"flag": "off",
			},
			expected: "flag: \"off\"\n",
		},
		{
			name: "YAML special value Yes (case-variant) is quoted",
			input: map[string]any{
				"answer": "Yes",
			},
			expected: "answer: \"Yes\"\n",
		},
		{
			name: "YAML special value YES (uppercase) is quoted",
			input: map[string]any{
				"answer": "YES",
			},
			expected: "answer: \"YES\"\n",
		},
		{
			name: "YAML special value No (case-variant) is quoted",
			input: map[string]any{
				"answer": "No",
			},
			expected: "answer: \"No\"\n",
		},
		{
			name: "YAML special value FALSE (uppercase) is quoted",
			input: map[string]any{
				"answer": "FALSE",
			},
			expected: "answer: \"FALSE\"\n",
		},
		{
			name: "YAML special value On (case-variant) is quoted",
			input: map[string]any{
				"switch": "On",
			},
			expected: "switch: \"On\"\n",
		},
		{
			name: "YAML special value OFF (uppercase) is quoted",
			input: map[string]any{
				"switch": "OFF",
			},
			expected: "switch: \"OFF\"\n",
		},
		{
			name: "YAML special value TRUE (uppercase) is quoted",
			input: map[string]any{
				"flag": "TRUE",
			},
			expected: "flag: \"TRUE\"\n",
		},
		{
			name: "YAML special value False (mixed case) is quoted",
			input: map[string]any{
				"flag": "False",
			},
			expected: "flag: \"False\"\n",
		},
		{
			name: "mixed numeric and regular strings",
			input: map[string]any{
				"port":    "8080",
				"host":    "localhost",
				"enabled": "true",
				"rate":    "3.14",
			},
			expected: "enabled: true\nhost: localhost\nport: 8080\nrate: 3.14\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := SerializeYAML(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(output))
		})
	}
}
