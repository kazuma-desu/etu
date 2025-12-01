package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/validator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintConfigPairs(t *testing.T) {
	t.Run("Human readable output", func(t *testing.T) {
		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/port", Value: int64(8080)},
		}

		output := captureOutput(func() {
			err := PrintConfigPairs(pairs, false)
			require.NoError(t, err)
		})

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

		output := captureOutput(func() {
			err := PrintConfigPairs(pairs, true)
			require.NoError(t, err)
		})

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

		output := captureOutput(func() {
			err := PrintConfigPairs(pairs, false)
			require.NoError(t, err)
		})

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

		output := captureOutput(func() {
			PrintValidationResult(result, false)
		})

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

		output := captureOutput(func() {
			PrintValidationResult(result, false)
		})

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

		output := captureOutput(func() {
			PrintValidationResult(result, false)
		})

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

		output := captureOutput(func() {
			PrintValidationResult(result, true)
		})

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

		output := captureOutput(func() {
			PrintValidationResult(result, false)
		})

		assert.Contains(t, output, "error(s)")
		assert.Contains(t, output, "warning(s)")
	})
}

func TestPrintDryRun(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/name", Value: "testapp"},
		{Key: "/app/version", Value: "1.0.0"},
	}

	output := captureOutput(func() {
		PrintDryRun(pairs)
	})

	assert.Contains(t, output, "DRY RUN")
	assert.Contains(t, output, "Would apply")
	assert.Contains(t, output, "/app/name")
	assert.Contains(t, output, "/app/version")
	assert.Contains(t, output, "testapp")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "PUT")
	assert.Contains(t, output, "no changes made")
}

func TestPrintApplyProgress(t *testing.T) {
	// PrintApplyProgress uses log.Info which doesn't write to stdout
	// Just verify it doesn't panic
	PrintApplyProgress(1, 10, "/test/key")
	PrintApplyProgress(10, 10, "/test/another")
}

func TestPrintApplySuccess(t *testing.T) {
	output := captureOutput(func() {
		PrintApplySuccess(5)
	})

	assert.Contains(t, output, "Successfully applied")
	assert.Contains(t, output, "5 items")
}

func TestPrintError(t *testing.T) {
	output := captureOutput(func() {
		PrintError(assert.AnError)
	})

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
		output := captureOutput(func() {
			Info("test message")
		})
		assert.Contains(t, output, "test message")
	})

	t.Run("Success", func(t *testing.T) {
		output := captureOutput(func() {
			Success("operation completed")
		})
		assert.Contains(t, output, "operation completed")
	})

	t.Run("Error", func(t *testing.T) {
		output := captureOutput(func() {
			Error("error occurred")
		})
		assert.Contains(t, output, "error occurred")
	})

	t.Run("Warning", func(t *testing.T) {
		output := captureOutput(func() {
			Warning("warning message")
		})
		assert.Contains(t, output, "warning message")
	})

	t.Run("Prompt", func(t *testing.T) {
		output := captureOutput(func() {
			Prompt("Enter value: ")
		})
		assert.Contains(t, output, "Enter value:")
	})
}

func TestPrintSecurityWarning(t *testing.T) {
	output := captureOutput(func() {
		PrintSecurityWarning()
	})

	assert.Contains(t, output, "Security Warning")
	assert.Contains(t, output, "plain text")
	assert.Contains(t, output, "password")
	assert.Contains(t, strings.ToLower(output), "security")
}
