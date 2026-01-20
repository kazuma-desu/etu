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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
