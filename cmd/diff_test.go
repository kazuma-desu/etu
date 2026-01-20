package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kazuma-desu/etu/pkg/models"
)

func TestExtractPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []*models.ConfigPair
		expected []string
	}{
		{
			name: "single prefix",
			pairs: []*models.ConfigPair{
				{Key: "/app/config/key1", Value: "val1"},
				{Key: "/app/config/key2", Value: "val2"},
			},
			expected: []string{"/app"},
		},
		{
			name: "multiple prefixes",
			pairs: []*models.ConfigPair{
				{Key: "/app/service1/key1", Value: "val1"},
				{Key: "/app/service2/key1", Value: "val1"},
			},
			expected: []string{"/app"},
		},
		{
			name: "deeply nested",
			pairs: []*models.ConfigPair{
				{Key: "/a/b/c/d", Value: "val"},
			},
			expected: []string{"/a"},
		},
		{
			name: "root keys",
			pairs: []*models.ConfigPair{
				{Key: "/key1", Value: "val1"},
			},
			expected: []string{"/key1"},
		},
		{
			name:     "empty input",
			pairs:    []*models.ConfigPair{},
			expected: []string{},
		},
		{
			name: "mixed root and nested",
			pairs: []*models.ConfigPair{
				{Key: "/root", Value: "val1"},
				{Key: "/app/nested", Value: "val2"},
			},
			expected: []string{"/root", "/app"},
		},
		{
			name: "different top-level prefixes",
			pairs: []*models.ConfigPair{
				{Key: "/app/config", Value: "val1"},
				{Key: "/db/config", Value: "val2"},
				{Key: "/cache/config", Value: "val3"},
			},
			expected: []string{"/app", "/db", "/cache"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPrefixes(tt.pairs)
			assert.ElementsMatch(t, tt.expected, got)
		})
	}
}

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
			got := formatValue(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
