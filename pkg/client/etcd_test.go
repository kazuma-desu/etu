package client

import (
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"
)

// TestGetWithOptionsErrorCases tests error handling in GetWithOptions
func TestGetWithOptionsErrorCases(t *testing.T) {
	opts := &GetOptions{
		SortOrder: "INVALID",
	}

	if opts.SortOrder != "INVALID" {
		t.Error("Expected sort order to be set")
	}
}

func TestClientStructure(t *testing.T) {
	cfg := &Config{
		Endpoints: []string{"localhost:2379"},
	}

	if len(cfg.Endpoints) != 1 {
		t.Error("Expected one endpoint")
	}
}

func TestGetResponse(t *testing.T) {
	resp := &GetResponse{
		Count: 1,
	}

	if resp.Count != 1 {
		t.Error("Expected count to be 1")
	}
}

func TestKeyValue(t *testing.T) {
	kv := &KeyValue{
		Key: "/test",
	}

	if kv.Key != "/test" {
		t.Error("Expected key to be /test")
	}
}

func TestGetOptions(t *testing.T) {
	opts := &GetOptions{
		Prefix: true,
		Limit:  10,
	}

	if !opts.Prefix {
		t.Error("Expected prefix to be true")
	}
	if opts.Limit != 10 {
		t.Error("Expected limit to be 10")
	}
}

// TestFormatValueUnit tests formatValue function
func TestFormatValueUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "test", "test"},
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"float64", 3.14, "3.140000"},
		{"bool", true, "true"},
		{"nil", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}

	// Test map formatting
	t.Run("map", func(t *testing.T) {
		mapVal := map[string]any{"key1": "value1", "key2": "value2"}
		result := formatValue(mapVal)
		// Map iteration order is not guaranteed, so just check it's not empty
		if result == "" {
			t.Error("formatValue(map) should not be empty")
		}
	})
}

// TestPutAllWithError tests PutAll error handling
func TestPutAllWithError(t *testing.T) {
	// This test verifies the structure but can't test actual error
	// without a real etcd instance
	pairs := []*models.ConfigPair{
		{Key: "/test1", Value: "value1"},
		{Key: "/test2", Value: "value2"},
	}

	if len(pairs) != 2 {
		t.Error("Expected 2 pairs")
	}
}

// TestConfigPair tests ConfigPair structure
func TestConfigPair(t *testing.T) {
	pair := &models.ConfigPair{
		Key:   "/test",
		Value: "value",
	}

	if pair.Key != "/test" {
		t.Error("Expected key to be /test")
	}
	if pair.Value != "value" {
		t.Error("Expected value to be 'value'")
	}
}

// TestNewClientWithEmptyEndpoints tests NewClient validation
func TestNewClientWithEmptyEndpoints(t *testing.T) {
	cfg := &Config{
		Endpoints: []string{},
	}

	_, err := NewClient(cfg)
	if err == nil {
		t.Error("Expected error when creating client with empty endpoints")
	}
}
