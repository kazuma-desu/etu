package parsers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.NotNil(t, r.parsers)

	parser, err := r.GetParser(models.FormatEtcdctl)
	assert.NoError(t, err)
	assert.NotNil(t, parser)
	assert.Equal(t, "etcdctl", parser.FormatName())

	yamlParser, err := r.GetParser(models.FormatYAML)
	assert.NoError(t, err)
	assert.NotNil(t, yamlParser)
	assert.Equal(t, "yaml", yamlParser.FormatName())

	jsonParser, err := r.GetParser(models.FormatJSON)
	assert.NoError(t, err)
	assert.NotNil(t, jsonParser)
	assert.Equal(t, "json", jsonParser.FormatName())
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	// Create a mock parser
	mock := &mockParser{name: "mock"}
	r.Register(models.FormatType("mock"), mock)

	parser, err := r.GetParser(models.FormatType("mock"))
	assert.NoError(t, err)
	assert.Equal(t, "mock", parser.FormatName())
}

func TestRegistry_GetParser(t *testing.T) {
	r := NewRegistry()

	t.Run("registered parser", func(t *testing.T) {
		parser, err := r.GetParser(models.FormatEtcdctl)
		assert.NoError(t, err)
		assert.NotNil(t, parser)
	})

	t.Run("unregistered parser", func(t *testing.T) {
		_, err := r.GetParser(models.FormatType("toml"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no parser registered for format")
	})
}

func TestDetectFormat_ByExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		expected models.FormatType
	}{
		{
			name:     "yaml extension",
			filename: "config.yaml",
			content:  "key: value",
			expected: models.FormatYAML,
		},
		{
			name:     "yml extension",
			filename: "config.yml",
			content:  "key: value",
			expected: models.FormatYAML,
		},
		{
			name:     "json extension",
			filename: "config.json",
			content:  `{"key": "value"}`,
			expected: models.FormatJSON,
		},
		{
			name:     "txt extension",
			filename: "config.txt",
			content:  "/app/key\nvalue",
			expected: models.FormatEtcdctl,
		},
		{
			name:     "no extension with etcdctl content",
			filename: "config",
			content:  "/app/key\nvalue",
			expected: models.FormatEtcdctl,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, format)
		})
	}
}

func TestDetectFormat_ByExtensionWithRegisteredParsers(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		content        string
		registerFormat models.FormatType
		expected       models.FormatType
	}{
		{
			name:           "yaml extension with yaml parser",
			filename:       "config.yaml",
			content:        "key: value",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
		{
			name:           "yml extension with yaml parser",
			filename:       "config.yml",
			content:        "---\nkey: value",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
		{
			name:           "json extension with json parser",
			filename:       "config.json",
			content:        `{"key": "value"}`,
			registerFormat: models.FormatJSON,
			expected:       models.FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			r.Register(tt.registerFormat, &mockParser{name: string(tt.registerFormat)})

			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, format)
		})
	}
}

func TestDetectFormat_ByContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected models.FormatType
	}{
		{
			name:     "json object",
			content:  `{"key": "value"}`,
			expected: models.FormatJSON,
		},
		{
			name:     "json array",
			content:  `[{"key": "value"}]`,
			expected: models.FormatJSON,
		},
		{
			name:     "json with leading whitespace",
			content:  `   {"key": "value"}`,
			expected: models.FormatJSON,
		},
		{
			name:     "yaml document separator",
			content:  "---\nkey: value",
			expected: models.FormatYAML,
		},
		{
			name:     "yaml key-value",
			content:  "key: value",
			expected: models.FormatYAML,
		},
		{
			name:     "yaml nested structure",
			content:  "database: config\n  host: localhost",
			expected: models.FormatYAML,
		},
		{
			name:     "etcdctl format",
			content:  "/app/key\nvalue",
			expected: models.FormatEtcdctl,
		},
		{
			name:     "etcdctl with comment",
			content:  "# comment\n/app/key\nvalue",
			expected: models.FormatEtcdctl,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, format)
		})
	}
}

func TestDetectFormat_FallbackToYAML(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "unknown content",
			content: "some random text",
		},
		{
			name:    "unrecognized format",
			content: "not-a-key = value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()

			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, models.FormatYAML, format, "Should fallback to YAML when no matching parser")
		})
	}
}

func TestDetectFormat_DefaultsToYAML(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "no extension, unknown content",
			filename: "config",
			content:  "some random text",
		},
		{
			name:     "unknown extension",
			filename: "config.unknown",
			content:  "some plain text without any format markers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, models.FormatYAML, format, "Should default to YAML for unknown extensions or content")
		})
	}
}

func TestDetectFormat_FileNotFound(t *testing.T) {
	r := NewRegistry()

	// Non-existent file should fallback to YAML (default)
	format, err := r.DetectFormat("/nonexistent/path/config.txt")
	assert.NoError(t, err)
	assert.Equal(t, models.FormatYAML, format)
}

func TestDetectFormat_ScannerError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "longline")

	// Create a line longer than bufio.Scanner's default 64KB buffer
	longLine := make([]byte, 65*1024)
	for i := range longLine {
		longLine[i] = 'x'
	}
	err := os.WriteFile(tmpFile, longLine, 0644)
	require.NoError(t, err)

	r := NewRegistry()
	format, err := r.DetectFormat(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, models.FormatYAML, format, "Scanner error should fallback to YAML via FormatAuto")
}

func TestDetectFormat_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty")
	err := os.WriteFile(tmpFile, []byte(""), 0644)
	require.NoError(t, err)

	r := NewRegistry()
	format, err := r.DetectFormat(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, models.FormatYAML, format, "Empty file should default to YAML")
}

func TestDetectFormat_CommentsOnly(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "single comment",
			content: "# This is a comment",
		},
		{
			name:    "multiple comments",
			content: "# Comment 1\n# Comment 2\n# Comment 3",
		},
		{
			name:    "comments with blank lines",
			content: "# Header\n\n# Another comment\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "comments")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, models.FormatYAML, format, "File with only comments should default to YAML")
		})
	}
}

func TestDetectFormat_MixedContentWithComments(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		registerFormat models.FormatType
		expected       models.FormatType
	}{
		{
			name:           "comment then json",
			content:        "# Config file\n{\"key\": \"value\"}",
			registerFormat: models.FormatJSON,
			expected:       models.FormatJSON,
		},
		{
			name:           "comment then yaml",
			content:        "# YAML config\nkey: value",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
		{
			name:           "comment then etcdctl",
			content:        "# etcdctl format\n/app/key\nvalue",
			registerFormat: models.FormatEtcdctl,
			expected:       models.FormatEtcdctl,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			if tt.registerFormat != models.FormatEtcdctl {
				r.Register(tt.registerFormat, &mockParser{name: string(tt.registerFormat)})
			}

			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, format)
		})
	}
}

func TestDetectFormat_CaseInsensitiveExtension(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		registerFormat models.FormatType
		expected       models.FormatType
	}{
		{
			name:           "uppercase YAML",
			filename:       "config.YAML",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
		{
			name:           "mixed case YaML",
			filename:       "config.YaML",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
		{
			name:           "uppercase JSON",
			filename:       "config.JSON",
			registerFormat: models.FormatJSON,
			expected:       models.FormatJSON,
		},
		{
			name:           "uppercase TXT",
			filename:       "config.TXT",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(tmpFile, []byte("content"), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			if tt.registerFormat != models.FormatEtcdctl {
				r.Register(tt.registerFormat, &mockParser{name: string(tt.registerFormat)})
			}

			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, format)
		})
	}
}

func TestDetectFormat_YAMLKeyValueNotStartingWithSlash(t *testing.T) {
	// Ensure YAML detection doesn't trigger for etcdctl key-value lines
	tests := []struct {
		name           string
		content        string
		registerFormat models.FormatType
		expected       models.FormatType
	}{
		{
			name:           "etcdctl path not confused with yaml",
			content:        "/config/db: value",
			registerFormat: models.FormatYAML,
			expected:       models.FormatEtcdctl, // Starts with /, so etcdctl
		},
		{
			name:           "yaml key detected",
			content:        "config: value",
			registerFormat: models.FormatYAML,
			expected:       models.FormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			r := NewRegistry()
			if tt.registerFormat != models.FormatEtcdctl {
				r.Register(tt.registerFormat, &mockParser{name: string(tt.registerFormat)})
			}

			format, err := r.DetectFormat(tmpFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, format)
		})
	}
}

type mockParser struct {
	name string
}

func (m *mockParser) Parse(_ context.Context, _ string) ([]*models.ConfigPair, error) {
	return nil, nil
}

func (m *mockParser) FormatName() string {
	return m.name
}

func BenchmarkDetectFormat_ByExtension(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.txt")
	err := os.WriteFile(tmpFile, []byte("/app/key\nvalue"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	r := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.DetectFormat(tmpFile)
	}
}

func BenchmarkDetectFormat_ByContent(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "config")
	err := os.WriteFile(tmpFile, []byte("/app/key\nvalue"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	r := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.DetectFormat(tmpFile)
	}
}
