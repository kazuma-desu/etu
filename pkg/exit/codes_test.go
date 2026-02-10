package exit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDescription(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{Success, "Success"},
		{GeneralError, "General error"},
		{ValidationError, "Validation error"},
		{ConnectionError, "Connection error"},
		{KeyNotFound, "Key not found"},
		{999, "Unknown error"},
		{-1, "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := GetDescription(tt.code)
			assert.Equal(t, tt.expected, got)
		})
	}
}
