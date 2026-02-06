package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/testutil"
	"github.com/kazuma-desu/etu/pkg/validator"
)

func TestPrintConfigPairsWithFormat_YAML(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/db/host", Value: "localhost"},
		{Key: "/app/db/port", Value: 5432},
		{Key: "/app/name", Value: "test-app"},
	}

	output, err := testutil.CaptureStdout(func() error {
		return PrintConfigPairsWithFormat(pairs, FormatYAML.String())
	})
	require.NoError(t, err)

	assert.Contains(t, output, "app:")
	assert.Contains(t, output, "db:")
	assert.Contains(t, output, "host: localhost")
	assert.Contains(t, output, "port: 5432")
	assert.Contains(t, output, "name: test-app")
}

func TestPrintValidationWithFormat_YAML(t *testing.T) {
	result := &validator.ValidationResult{
		Valid: false,
		Issues: []validator.ValidationIssue{
			{
				Key:     "/invalid/key",
				Message: "Invalid format",
				Level:   "error",
			},
		},
	}

	output, err := testutil.CaptureStdout(func() error {
		return PrintValidationWithFormat(result, true, FormatYAML.String())
	})
	require.NoError(t, err)

	assert.Contains(t, output, "valid: false")
	assert.Contains(t, output, "strict: true")
	assert.Contains(t, output, "issues:")
	assert.Contains(t, output, "key: /invalid/key")
	assert.Contains(t, output, "message: Invalid format")
	assert.Contains(t, output, "level: error")
}

func TestPrintApplyResultsWithFormat_YAML(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/config", Value: "value"},
	}

	t.Run("Dry Run", func(t *testing.T) {
		output, err := testutil.CaptureStdout(func() error {
			return PrintApplyResultsWithFormat(pairs, FormatYAML.String(), true)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "applied: 1")
		assert.Contains(t, output, "dry_run: true")
		assert.Contains(t, output, "items:")
		assert.Contains(t, output, "key: /app/config")
	})

	t.Run("Applied", func(t *testing.T) {
		output, err := testutil.CaptureStdout(func() error {
			return PrintApplyResultsWithFormat(pairs, FormatYAML.String(), false)
		})
		require.NoError(t, err)

		assert.Contains(t, output, "applied: 1")
		assert.Contains(t, output, "dry_run: false")
		assert.Contains(t, output, "items:")
		assert.Contains(t, output, "key: /app/config")
	})
}

func TestPrintContextsWithFormat_YAML(t *testing.T) {
	contexts := map[string]*ContextView{
		"dev": {
			Endpoints: []string{"http://localhost:2379"},
			Username:  "admin",
		},
		"prod": {
			Endpoints: []string{"http://prod:2379"},
			Username:  "admin",
		},
	}

	output, err := testutil.CaptureStdout(func() error {
		return PrintContextsWithFormat(contexts, "dev", FormatYAML.String())
	})
	require.NoError(t, err)

	assert.Contains(t, output, "name: dev")
	assert.Contains(t, output, "current: true")
	assert.Contains(t, output, "name: prod")
	assert.Contains(t, output, "current: false")
	assert.Contains(t, output, "endpoints:")
	assert.Contains(t, output, "- http://localhost:2379")
}

func TestPrintConfigViewWithFormat_YAML(t *testing.T) {
	cfg := &ConfigView{
		CurrentContext: "dev",
		LogLevel:       "debug",
		DefaultFormat:  "yaml",
		Strict:         true,
		NoValidate:     false,
		Contexts: map[string]*ContextView{
			"dev": {
				Endpoints: []string{"http://localhost:2379"},
			},
		},
	}

	output, err := testutil.CaptureStdout(func() error {
		return PrintConfigViewWithFormat(cfg, FormatYAML.String())
	})
	require.NoError(t, err)

	assert.Contains(t, output, "current_context: dev")
	assert.Contains(t, output, "log_level: debug")
	assert.Contains(t, output, "default_format: yaml")
	assert.Contains(t, output, "strict: true")
	assert.Contains(t, output, "no_validate: false")
	assert.Contains(t, output, "contexts:")
}
