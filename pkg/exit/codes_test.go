package exit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDescription(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"success", Success, "Success"},
		{"general error", GeneralError, "General error"},
		{"validation error", ValidationError, "Validation error"},
		{"connection error", ConnectionError, "Connection error"},
		{"key not found", KeyNotFound, "Key not found"},
		{"unknown error code 999", 999, "Unknown error"},
		{"unknown error code -1", -1, "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDescription(tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}
