package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// Nil
		{"nil", nil, ""},

		// Strings
		{"string", "hello", "hello"},
		{"empty string", "", ""},

		// Signed integers
		{"int", int(42), "42"},
		{"int8", int8(8), "8"},
		{"int16", int16(16), "16"},
		{"int32", int32(32), "32"},
		{"int64", int64(64), "64"},
		{"negative int", int(-42), "-42"},
		{"zero", int(0), "0"},

		// Unsigned integers
		{"uint", uint(42), "42"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},

		// Floats
		{"float32", float32(3.14), "3.14"},
		{"float64", float64(3.14), "3.14"},
		{"float64 zero", float64(0.0), "0"},

		// Bool
		{"bool true", true, "true"},
		{"bool false", false, "false"},

		// Maps
		{"empty map", map[string]any{}, ""},
		{"map with values", map[string]any{"a": "1", "b": "2"}, "a: 1\nb: 2"},
		{"map with mixed types", map[string]any{"x": 1, "y": "hello"}, "x: 1\ny: hello"},

		// fmt.Stringer
		{"Stringer", testStringer("custom"), "custom"},

		// Default fallback
		{"slice (fallback)", []int{1, 2, 3}, "[1 2 3]"},
		{"struct (fallback)", struct{ Name string }{"test"}, "{test}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatValueMapExact(t *testing.T) {
	// Verify map formatting produces exactly the expected lines (sorted by key)
	m := map[string]any{"z": "last", "a": "first", "m": "middle"}
	got := FormatValue(m)

	// Since keys are sorted, output should be deterministic
	expected := "a: first\nm: middle\nz: last"
	assert.Equal(t, expected, got)

	// Verify no extra lines by splitting and counting
	lines := strings.Split(got, "\n")
	assert.Len(t, lines, 3, "should have exactly 3 lines for 3 map entries")
}

// testStringer implements fmt.Stringer for testing
type testStringer string

func (s testStringer) String() string {
	return string(s)
}

// BenchmarkFormatValue benchmarks the FormatValue function
func BenchmarkFormatValue(b *testing.B) {
	values := []any{
		"string",
		42,
		int64(64),
		3.14,
		true,
		map[string]any{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			FormatValue(v)
		}
	}
}

func TestInferType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"true string", "true", true},
		{"false string", "false", false},
		{"positive integer", "123", int64(123)},
		{"negative integer", "-456", int64(-456)},
		{"zero", "0", int64(0)},
		{"simple float", "1.5", float64(1.5)},
		{"scientific notation uppercase", "1e10", float64(1e10)},
		{"scientific notation lowercase", "1E10", float64(1e10)},
		{"negative scientific notation", "-1.5e-3", float64(-1.5e-3)},
		{"float with positive exponent", "2.5e+3", float64(2.5e3)},
		{"simple string", "hello", "hello"},
		{"alphanumeric string", "123abc", "123abc"},
		{"empty string", "", ""},
		{"multiple decimal points", "123.456.789", "123.456.789"},
		{"string with spaces", "hello world", "hello world"},
		{"int64 overflow", "99999999999999999999", "99999999999999999999"},
		{"case sensitivity True", "True", "True"},
		{"case sensitivity TRUE", "TRUE", "TRUE"},
		{"case sensitivity FALSE", "FALSE", "FALSE"},
		{"leading zeros", "007", int64(7)},
		{"whitespace prefix", " 123", " 123"},
		{"whitespace suffix", "123 ", "123 "},
		{"whitespace both", " 123 ", " 123 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferType(tt.input)

			if resultType(result) != resultType(tt.expected) {
				t.Errorf("InferType(%q) = %T, want %T", tt.input, result, tt.expected)
				return
			}

			switch v := result.(type) {
			case bool:
				assert.Equal(t, tt.expected.(bool), v)
			case int64:
				assert.Equal(t, tt.expected.(int64), v)
			case float64:
				assert.Equal(t, tt.expected.(float64), v)
			case string:
				assert.Equal(t, tt.expected.(string), v)
			}
		})
	}
}

func resultType(v any) string {
	switch v.(type) {
	case bool:
		return "bool"
	case int64:
		return "int64"
	case float64:
		return "float64"
	case string:
		return "string"
	default:
		return "unknown"
	}
}
