package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEtcdctlParser_FormatName(t *testing.T) {
	parser := &EtcdctlParser{}
	assert.Equal(t, "etcdctl", parser.FormatName())
}

func TestEtcdctlParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []*models.ConfigPair
		wantErr  bool
	}{
		{
			name: "simple key-value",
			content: `/app/name
myapp`,
			expected: []*models.ConfigPair{
				{Key: "/app/name", Value: "myapp"},
			},
		},
		{
			name: "multiple key-values",
			content: `/app/name
myapp

/app/version
1.0.0`,
			expected: []*models.ConfigPair{
				{Key: "/app/name", Value: "myapp"},
				{Key: "/app/version", Value: "1.0.0"},
			},
		},
		{
			name: "integer value",
			content: `/app/port
8080`,
			expected: []*models.ConfigPair{
				{Key: "/app/port", Value: int64(8080)},
			},
		},
		{
			name: "negative integer",
			content: `/app/offset
-100`,
			expected: []*models.ConfigPair{
				{Key: "/app/offset", Value: int64(-100)},
			},
		},
		{
			name: "float value",
			content: `/app/threshold
0.95`,
			expected: []*models.ConfigPair{
				{Key: "/app/threshold", Value: 0.95},
			},
		},
		{
			name: "negative float",
			content: `/app/temperature
-3.14`,
			expected: []*models.ConfigPair{
				{Key: "/app/temperature", Value: -3.14},
			},
		},
		{
			name: "quoted string",
			content: `/app/message
"hello world"`,
			expected: []*models.ConfigPair{
				{Key: "/app/message", Value: "hello world"},
			},
		},
		{
			name: "single quoted string",
			content: `/app/message
'hello world'`,
			expected: []*models.ConfigPair{
				{Key: "/app/message", Value: "hello world"},
			},
		},
		{
			name: "language map format",
			content: `/app/title
en: English Title
es: Spanish Title
fr: French Title`,
			expected: []*models.ConfigPair{
				{
					Key: "/app/title",
					Value: map[string]any{
						"en": "English Title",
						"es": "Spanish Title",
						"fr": "French Title",
					},
				},
			},
		},
		{
			name: "language map with quotes",
			content: `/app/greeting
en: "Hello"
es: "Hola"`,
			expected: []*models.ConfigPair{
				{
					Key: "/app/greeting",
					Value: map[string]any{
						"en": "Hello",
						"es": "Hola",
					},
				},
			},
		},
		{
			name: "multiline value",
			content: `/app/config
line 1
line 2
line 3`,
			expected: []*models.ConfigPair{
				{Key: "/app/config", Value: "line 1\nline 2\nline 3"},
			},
		},
		{
			name:     "empty file",
			content:  "",
			expected: []*models.ConfigPair{},
		},
		{
			name: "trailing empty lines",
			content: `/app/name
myapp


`,
			expected: []*models.ConfigPair{
				{Key: "/app/name", Value: "myapp"},
			},
		},
		{
			name: "key with no value",
			content: `/app/empty
`,
			expected: []*models.ConfigPair{
				{Key: "/app/empty", Value: ""},
			},
		},
		{
			name: "complex nested path",
			content: `/app/services/database/primary/host
localhost`,
			expected: []*models.ConfigPair{
				{Key: "/app/services/database/primary/host", Value: "localhost"},
			},
		},
		{
			name: "value with spaces",
			content: `/app/description
This is a long description with multiple words`,
			expected: []*models.ConfigPair{
				{Key: "/app/description", Value: "This is a long description with multiple words"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err, "Failed to create temp file")

			parser := &EtcdctlParser{}
			got, err := parser.Parse(tmpFile)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, len(tt.expected), "Number of parsed pairs doesn't match")

			for i := range got {
				assert.Equal(t, tt.expected[i].Key, got[i].Key, "Key mismatch at index %d", i)

				// Compare values based on type
				switch expectedVal := tt.expected[i].Value.(type) {
				case map[string]any:
					gotMap, ok := got[i].Value.(map[string]any)
					require.True(t, ok, "Expected map[string]any at index %d, got %T", i, got[i].Value)
					assert.Equal(t, expectedVal, gotMap, "Map value mismatch at index %d", i)
				default:
					assert.Equal(t, tt.expected[i].Value, got[i].Value, "Value mismatch at index %d", i)
				}
			}
		})
	}
}

func TestEtcdctlParser_ParseNonExistentFile(t *testing.T) {
	parser := &EtcdctlParser{}
	_, err := parser.Parse("/nonexistent/file/that/does/not/exist.txt")
	assert.Error(t, err, "Expected error for non-existent file")
}

func TestEtcdctlParser_ParseInvalidPath(t *testing.T) {
	parser := &EtcdctlParser{}
	_, err := parser.Parse("")
	assert.Error(t, err, "Expected error for empty path")
}

func TestStripWrappingQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `"hello"`, "hello"},
		{"single quotes", `'hello'`, "hello"},
		{"no quotes", `hello`, "hello"},
		{"double quotes with spaces", `"hello world"`, "hello world"},
		{"single quotes with spaces", `'hello world'`, "hello world"},
		{"empty double quotes", `""`, ""},
		{"empty single quotes", `''`, ""},
		{"single char", `"a"`, "a"},
		{"only opening quote", `"hello`, `"hello`},
		{"only closing quote", `hello"`, `hello"`},
		{"mismatched quotes", `"hello'`, `"hello'`},
		{"with whitespace", `  "hello"  `, "hello"},
		{"with tabs", "	\"hello\"	", "hello"},
		{"nested quotes", `"'inner'"`, `'inner'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripWrappingQuotes(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEtcdctlParser_ParseScalar(t *testing.T) {
	parser := &EtcdctlParser{}
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"positive integer", "42", int64(42)},
		{"negative integer", "-42", int64(-42)},
		{"zero", "0", int64(0)},
		{"large integer", "999999999", int64(999999999)},
		{"float", "3.14", 3.14},
		{"negative float", "-3.14", -3.14},
		{"decimal less than one", "0.5", 0.5},
		{"decimal without leading zero", ".5", 0.5},
		{"simple string", "hello", "hello"},
		{"double quoted", `"quoted"`, "quoted"},
		{"single quoted", `'quoted'`, "quoted"},
		{"empty string", "", ""},
		{"string with spaces", "hello world", "hello world"},
		{"string with numbers", "version123", "version123"},
		{"url", "http://example.com", "http://example.com"},
		{"boolean-like string", "true", "true"},
		{"scientific notation lookalike", "1e5", "1e5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.parseScalar(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEtcdctlParser_ParseValueLines(t *testing.T) {
	parser := &EtcdctlParser{}
	tests := []struct {
		name     string
		lines    []string
		expected any
	}{
		{
			name:     "empty lines",
			lines:    []string{},
			expected: "",
		},
		{
			name:     "single line",
			lines:    []string{"hello"},
			expected: "hello",
		},
		{
			name:     "single integer",
			lines:    []string{"42"},
			expected: int64(42),
		},
		{
			name:     "language map",
			lines:    []string{"en: Hello", "es: Hola"},
			expected: map[string]any{"en": "Hello", "es": "Hola"},
		},
		{
			name:     "multiline text",
			lines:    []string{"line 1", "line 2", "line 3"},
			expected: "line 1\nline 2\nline 3",
		},
		{
			name:     "mixed format falls back to multiline",
			lines:    []string{"en: Hello", "not a tag"},
			expected: "en: Hello\nnot a tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.parseValueLines(tt.lines)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// Benchmark for etcdctl parser
func BenchmarkEtcdctlParser_Parse(b *testing.B) {
	// Create a temporary file with test content
	content := `/app/name
myapp

/app/version
1.0.0

/app/port
8080`

	tmpfile, err := os.CreateTemp("", "bench-*.txt")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		b.Fatal(err)
	}
	tmpfile.Close()

	parser := &EtcdctlParser{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(tmpfile.Name())
	}
}
