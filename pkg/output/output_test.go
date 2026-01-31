package output

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/testutil"
	"github.com/kazuma-desu/etu/pkg/validator"
)

func TestPrintConfigPairs(t *testing.T) {
	t.Run("Human readable output", func(t *testing.T) {
		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/port", Value: int64(8080)},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairs(pairs, false)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/app/name")
		assert.Contains(t, output, "myapp")
		assert.Contains(t, output, "/app/port")
		assert.Contains(t, output, "8080")
	})

	t.Run("JSON output", func(t *testing.T) {
		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/port", Value: int64(8080)},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairs(pairs, true)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, `"key"`)
		assert.Contains(t, output, `"value"`)
		assert.Contains(t, output, "/app/name")
		assert.Contains(t, output, "myapp")
	})

	t.Run("Map value formatting", func(t *testing.T) {
		pairs := []*models.ConfigPair{
			{
				Key: "/config/settings",
				Value: map[string]any{
					"timeout": 30,
					"retries": 3,
				},
			},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairs(pairs, false)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/config/settings")
		assert.Contains(t, output, "timeout")
		assert.Contains(t, output, "retries")
	})
}

func TestPrintValidationResult(t *testing.T) {
	t.Run("No issues", func(t *testing.T) {
		result := &validator.ValidationResult{
			Valid:  true,
			Issues: []validator.ValidationIssue{},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			PrintValidationResult(result, false)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "Validation passed")
		assert.Contains(t, output, "no issues")
	})

	t.Run("With errors", func(t *testing.T) {
		result := &validator.ValidationResult{
			Valid: false,
			Issues: []validator.ValidationIssue{
				{
					Key:     "/invalid/key",
					Message: "Invalid key format",
					Level:   "error",
				},
			},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			PrintValidationResult(result, false)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "error(s)")
		assert.Contains(t, output, "/invalid/key")
		assert.Contains(t, output, "Invalid key format")
		assert.Contains(t, output, "Validation failed")
	})

	t.Run("With warnings", func(t *testing.T) {
		result := &validator.ValidationResult{
			Valid: true,
			Issues: []validator.ValidationIssue{
				{
					Key:     "/config/deprecated",
					Message: "This key is deprecated",
					Level:   "warning",
				},
			},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			PrintValidationResult(result, false)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "warning(s)")
		assert.Contains(t, output, "/config/deprecated")
		assert.Contains(t, output, "deprecated")
	})

	t.Run("Strict mode with warnings", func(t *testing.T) {
		result := &validator.ValidationResult{
			Valid: false,
			Issues: []validator.ValidationIssue{
				{
					Key:     "/config/test",
					Message: "Warning message",
					Level:   "warning",
				},
			},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			PrintValidationResult(result, true)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "strict mode")
		assert.Contains(t, output, "warnings treated as errors")
	})

	t.Run("Mixed errors and warnings", func(t *testing.T) {
		result := &validator.ValidationResult{
			Valid: false,
			Issues: []validator.ValidationIssue{
				{
					Key:     "/error/key",
					Message: "Error message",
					Level:   "error",
				},
				{
					Key:     "/warning/key",
					Message: "Warning message",
					Level:   "warning",
				},
			},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			PrintValidationResult(result, false)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "error(s)")
		assert.Contains(t, output, "warning(s)")
	})
}

func TestPrintDryRun(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/name", Value: "testapp"},
		{Key: "/app/version", Value: "1.0.0"},
	}

	output, err := testutil.CaptureStdoutFunc(func() {
		PrintDryRun(pairs)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "DRY RUN")
	assert.Contains(t, output, "Would apply")
	assert.Contains(t, output, "/app/name")
	assert.Contains(t, output, "/app/version")
	assert.Contains(t, output, "testapp")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "PUT")
	assert.Contains(t, output, "no changes made")
}

func TestPrintApplyProgress(_ *testing.T) {
	// PrintApplyProgress uses log.Info which doesn't write to stdout
	// Just verify it doesn't panic
	PrintApplyProgress(1, 10, "/test/key")
	PrintApplyProgress(10, 10, "/test/another")
}

func TestPrintApplySuccess(t *testing.T) {
	output, err := testutil.CaptureStdoutFunc(func() {
		PrintApplySuccess(5)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "Successfully applied")
	assert.Contains(t, output, "5 items")
}

func TestPrintError(t *testing.T) {
	output, err := testutil.CaptureStdoutFunc(func() {
		PrintError(assert.AnError)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "Error")
	assert.Contains(t, output, assert.AnError.Error())
}

func TestFormatValue(t *testing.T) {
	t.Run("String value", func(t *testing.T) {
		result := formatValue("test")
		assert.Equal(t, "test", result)
	})

	t.Run("Map value", func(t *testing.T) {
		mapVal := map[string]any{
			"key1": "value1",
			"key2": "value2",
		}
		result := formatValue(mapVal)
		assert.Contains(t, result, "key1: value1")
		assert.Contains(t, result, "key2: value2")
	})

	t.Run("Other types", func(t *testing.T) {
		result := formatValue(42)
		assert.Equal(t, "42", result)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("Info", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			Info("test message")
		})
		require.NoError(t, err)
		assert.Contains(t, output, "test message")
	})

	t.Run("Success", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			Success("operation completed")
		})
		require.NoError(t, err)
		assert.Contains(t, output, "operation completed")
	})

	t.Run("Error", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			Error("error occurred")
		})
		require.NoError(t, err)
		assert.Contains(t, output, "error occurred")
	})

	t.Run("Warning", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			Warning("warning message")
		})
		require.NoError(t, err)
		assert.Contains(t, output, "warning message")
	})

	t.Run("Prompt", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			Prompt("Enter value: ")
		})
		require.NoError(t, err)
		assert.Contains(t, output, "Enter value:")
	})
}

func TestPrintSecurityWarning(t *testing.T) {
	output, err := testutil.CaptureStdoutFunc(func() {
		PrintSecurityWarning()
	})
	require.NoError(t, err)

	assert.Contains(t, output, "Security Warning")
	assert.Contains(t, output, "plain text")
	assert.Contains(t, output, "password")
	assert.Contains(t, strings.ToLower(output), "security")
}

func TestColorConstantsAreDefined(t *testing.T) {
	// This test documents the expected color format
	// All colors should be in hex format (e.g., "#7C3AED") not ANSI codes (e.g., "252")
	colors := map[string]lipgloss.Color{
		"colorPrimary":   colorPrimary,
		"colorSuccess":   colorSuccess,
		"colorWarning":   colorWarning,
		"colorError":     colorError,
		"colorInfo":      colorInfo,
		"colorMuted":     colorMuted,
		"colorHighlight": colorHighlight,
		"colorTableOdd":  colorTableOdd,
		"colorTableEven": colorTableEven,
	}

	// Note: lipgloss.Color is an opaque type, so we can't directly inspect the value
	// This test serves as documentation that all colors should be hex
	for name := range colors {
		// Just verify the color is defined (not nil/zero)
		assert.NotNil(t, colors[name], "Color %s should be defined", name)
	}
}

// TestIsTerminal does not mock terminal detection but documents the behavior.
func TestIsTerminal(t *testing.T) {
	// IsTerminal returns true when stdout is a TTY
	// In test environment, this is typically false (redirected)
	// We just verify the function doesn't panic
	result := IsTerminal()
	assert.IsType(t, true, result)
}

// TestStyleIfTerminal applies styling only in terminal mode.
func TestStyleIfTerminal(t *testing.T) {
	testStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	content := "test content"

	// When not in terminal, should return unstyled content
	if !IsTerminal() {
		result := StyleIfTerminal(testStyle, content)
		assert.Equal(t, content, result)
	}
}

// TestFormatConstants verifies all format constants are defined.
func TestFormatConstants(t *testing.T) {
	assert.Equal(t, Format("simple"), FormatSimple)
	assert.Equal(t, Format("json"), FormatJSON)
	assert.Equal(t, Format("table"), FormatTable)
	assert.Equal(t, Format("tree"), FormatTree)
	assert.Equal(t, Format("fields"), FormatFields)
}

// TestFormatIsValid validates format detection.
func TestFormatIsValid(t *testing.T) {
	tests := []struct {
		format  Format
		isValid bool
	}{
		{FormatSimple, true},
		{FormatJSON, true},
		{FormatTable, true},
		{FormatTree, true},
		{FormatFields, true},
		{Format("invalid"), false},
		{Format(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.format.IsValid())
		})
	}
}

// TestParseFormat validates format parsing with error cases.
func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		wantErr  bool
		expected Format
	}{
		{"simple", false, FormatSimple},
		{"json", false, FormatJSON},
		{"table", false, FormatTable},
		{"tree", false, FormatTree},
		{"fields", false, FormatFields},
		{"invalid", true, ""},
		{"", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseFormat(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// BenchmarkParseFormat measures format validation performance.
func BenchmarkParseFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseFormat("json")
	}
}

// TestAllFormats verifies AllFormats returns a copy of all supported formats.
func TestAllFormats(t *testing.T) {
	formats := AllFormats()

	assert.Len(t, formats, 5)
	assert.Contains(t, formats, FormatSimple)
	assert.Contains(t, formats, FormatJSON)
	assert.Contains(t, formats, FormatTable)
	assert.Contains(t, formats, FormatTree)
	assert.Contains(t, formats, FormatFields)

	// Verify it's a copy by modifying the returned slice
	originalLen := len(formats)
	_ = append(formats, Format("extra"))
	assert.Len(t, AllFormats(), originalLen)
}

func TestPrintTree(t *testing.T) {
	t.Run("Simple tree output", func(t *testing.T) {
		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/port", Value: int64(8080)},
			{Key: "/db/host", Value: "localhost"},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintTree(pairs)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/")
		assert.Contains(t, output, "app")
		assert.Contains(t, output, "db")
		assert.Contains(t, output, "myapp")
		assert.Contains(t, output, "8080")
		assert.Contains(t, output, "localhost")
	})

	t.Run("Empty pairs", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintTree([]*models.ConfigPair{})
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/")
	})

	t.Run("Nested paths", func(t *testing.T) {
		pairs := []*models.ConfigPair{
			{Key: "/a/b/c/d", Value: "deep"},
		}

		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintTree(pairs)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "a")
		assert.Contains(t, output, "b")
		assert.Contains(t, output, "c")
		assert.Contains(t, output, "d")
		assert.Contains(t, output, "deep")
	})
}

func TestPrintConfigPairsWithFormat(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/name", Value: "myapp"},
		{Key: "/app/port", Value: int64(8080)},
	}

	t.Run("Simple format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairsWithFormat(pairs, "simple")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/app/name")
		assert.Contains(t, output, "myapp")
	})

	t.Run("JSON format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairsWithFormat(pairs, "json")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, `"key"`)
		assert.Contains(t, output, `"value"`)
	})

	t.Run("Table format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairsWithFormat(pairs, "table")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "KEY")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "/app/name")
	})

	t.Run("Tree format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigPairsWithFormat(pairs, "tree")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "/")
		assert.Contains(t, output, "app")
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		err := PrintConfigPairsWithFormat(pairs, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestPrintValidationWithFormat(t *testing.T) {
	result := &validator.ValidationResult{
		Valid: true,
		Issues: []validator.ValidationIssue{
			{Key: "/test", Message: "Test warning", Level: "warning"},
		},
	}

	t.Run("Simple format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintValidationWithFormat(result, false, "simple")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "warning")
	})

	t.Run("JSON format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintValidationWithFormat(result, false, "json")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, `"valid"`)
		assert.Contains(t, output, `"issues"`)
	})

	t.Run("Table format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintValidationWithFormat(result, false, "table")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "LEVEL")
		assert.Contains(t, output, "KEY")
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		err := PrintValidationWithFormat(result, false, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestPrintApplyResultsWithFormat(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/name", Value: "myapp"},
	}

	t.Run("Simple format dry run", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintApplyResultsWithFormat(pairs, "simple", true)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "DRY RUN")
	})

	t.Run("Simple format applied", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintApplyResultsWithFormat(pairs, "simple", false)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "Successfully applied")
	})

	t.Run("JSON format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintApplyResultsWithFormat(pairs, "json", false)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, `"applied"`)
		assert.Contains(t, output, `"dry_run"`)
	})

	t.Run("Table format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintApplyResultsWithFormat(pairs, "table", false)
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "KEY")
		assert.Contains(t, output, "VALUE")
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		err := PrintApplyResultsWithFormat(pairs, "invalid", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestPrintContextsWithFormat(t *testing.T) {
	contexts := map[string]*ContextView{
		"dev":  {Username: "admin", Endpoints: []string{"http://localhost:2379"}},
		"prod": {Username: "admin", Endpoints: []string{"http://prod:2379", "http://prod2:2379"}},
	}

	t.Run("Simple format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintContextsWithFormat(contexts, "dev", "simple")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "dev")
		assert.Contains(t, output, "prod")
	})

	t.Run("Simple format with current context", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintContextsWithFormat(contexts, "prod", "simple")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "dev")
		assert.Contains(t, output, "prod")
	})

	t.Run("Simple format empty contexts", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintContextsWithFormat(map[string]*ContextView{}, "", "simple")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "No contexts")
	})

	t.Run("JSON format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintContextsWithFormat(contexts, "dev", "json")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, `"name"`)
		assert.Contains(t, output, `"endpoints"`)
		assert.Contains(t, output, "dev")
	})

	t.Run("Table format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintContextsWithFormat(contexts, "dev", "table")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "NAME")
		assert.Contains(t, output, "ENDPOINTS")
		assert.Contains(t, output, "dev")
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		err := PrintContextsWithFormat(contexts, "dev", "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

func TestPrintConfigViewWithFormat(t *testing.T) {
	cfg := &ConfigView{
		CurrentContext: "prod",
		LogLevel:       "info",
		DefaultFormat:  "simple",
		Strict:         true,
		NoValidate:     false,
		Contexts: map[string]*ContextView{
			"dev":  {Username: "admin", Endpoints: []string{"http://localhost:2379"}},
			"prod": {Username: "admin", Endpoints: []string{"http://prod:2379"}},
		},
	}

	t.Run("Simple format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigViewWithFormat(cfg, "simple")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "current-context")
		assert.Contains(t, output, "prod")
		assert.Contains(t, output, "log-level")
		assert.Contains(t, output, "info")
	})

	t.Run("JSON format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigViewWithFormat(cfg, "json")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, `"current_context"`)
		assert.Contains(t, output, `"log_level"`)
		assert.Contains(t, output, "prod")
	})

	t.Run("Table format", func(t *testing.T) {
		output, err := testutil.CaptureStdoutFunc(func() {
			err := PrintConfigViewWithFormat(cfg, "table")
			require.NoError(t, err)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "SETTING")
		assert.Contains(t, output, "VALUE")
		assert.Contains(t, output, "current-context")
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		err := PrintConfigViewWithFormat(cfg, "invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}
