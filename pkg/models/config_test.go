package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigPair_String(t *testing.T) {
	tests := []struct {
		name     string
		pair     *ConfigPair
		expected string
	}{
		{
			name:     "string value",
			pair:     &ConfigPair{Key: "/app/name", Value: "myapp"},
			expected: "/app/name: myapp",
		},
		{
			name:     "integer value",
			pair:     &ConfigPair{Key: "/app/port", Value: 8080},
			expected: "/app/port: 8080",
		},
		{
			name:     "map value",
			pair:     &ConfigPair{Key: "/app/config", Value: map[string]string{"key": "value"}},
			expected: "/app/config: map[key:value]",
		},
		{
			name:     "nil value",
			pair:     &ConfigPair{Key: "/app/nil", Value: nil},
			expected: "/app/nil: <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pair.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		format   FormatType
		expected bool
	}{
		{"auto format", FormatAuto, true},
		{"etcdctl format", FormatEtcdctl, true},
		{"invalid format", FormatType("invalid"), false},
		{"empty format", FormatType(""), false},
		{"random string", FormatType("yaml"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.format.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatType_Constants(t *testing.T) {
	assert.Equal(t, FormatType("auto"), FormatAuto)
	assert.Equal(t, FormatType("etcdctl"), FormatEtcdctl)
}

func TestApplyOptions(t *testing.T) {
	opts := ApplyOptions{
		FilePath:   "/path/to/config",
		Format:     FormatEtcdctl,
		DryRun:     true,
		NoValidate: false,
		Strict:     true,
	}

	assert.Equal(t, "/path/to/config", opts.FilePath)
	assert.Equal(t, FormatEtcdctl, opts.Format)
	assert.True(t, opts.DryRun)
	assert.False(t, opts.NoValidate)
	assert.True(t, opts.Strict)
}

func TestValidateOptions(t *testing.T) {
	opts := ValidateOptions{
		FilePath: "/path/to/config",
		Format:   FormatEtcdctl,
		Strict:   true,
	}

	assert.Equal(t, "/path/to/config", opts.FilePath)
	assert.Equal(t, FormatEtcdctl, opts.Format)
	assert.True(t, opts.Strict)
}

func TestParseOptions(t *testing.T) {
	opts := ParseOptions{
		FilePath:   "/path/to/config",
		Format:     FormatAuto,
		JSONOutput: true,
	}

	assert.Equal(t, "/path/to/config", opts.FilePath)
	assert.Equal(t, FormatAuto, opts.Format)
	assert.True(t, opts.JSONOutput)
}
