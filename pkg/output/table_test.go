package output

import (
	"strings"
	"testing"
)

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name     string
		config   TableConfig
		contains []string // Strings that should be in the output
	}{
		{
			name: "simple table",
			config: TableConfig{
				Headers: []string{"Name", "Value"},
				Rows: [][]string{
					{"key1", "value1"},
					{"key2", "value2"},
				},
			},
			contains: []string{"Name", "Value", "key1", "value1", "key2", "value2"},
		},
		{
			name: "table with metadata",
			config: TableConfig{
				Headers: []string{"KEY", "VALUE", "CREATE_REV", "MOD_REV"},
				Rows: [][]string{
					{"/config/app/host", "localhost", "1", "1"},
					{"/config/app/port", "8080", "2", "2"},
				},
			},
			contains: []string{"KEY", "VALUE", "CREATE_REV", "MOD_REV", "/config/app/host", "localhost", "8080"},
		},
		{
			name: "empty table",
			config: TableConfig{
				Headers: []string{"KEY", "VALUE"},
				Rows:    [][]string{},
			},
			contains: []string{"KEY", "VALUE"},
		},
		{
			name: "single row",
			config: TableConfig{
				Headers: []string{"KEY", "VALUE"},
				Rows: [][]string{
					{"/mykey", "myvalue"},
				},
			},
			contains: []string{"KEY", "VALUE", "/mykey", "myvalue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTable(tt.config)

			// Check that output is not empty
			if result == "" {
				t.Error("Expected non-empty output")
			}

			// Check that all expected strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Output does not contain expected string: %s\nOutput:\n%s", expected, result)
				}
			}
		})
	}
}

func TestRenderTableRowCount(t *testing.T) {
	config := TableConfig{
		Headers: []string{"KEY", "VALUE"},
		Rows: [][]string{
			{"key1", "val1"},
			{"key2", "val2"},
			{"key3", "val3"},
		},
	}

	result := RenderTable(config)

	// The output should have all the keys
	if !strings.Contains(result, "key1") {
		t.Error("Missing key1 in output")
	}
	if !strings.Contains(result, "key2") {
		t.Error("Missing key2 in output")
	}
	if !strings.Contains(result, "key3") {
		t.Error("Missing key3 in output")
	}
}
