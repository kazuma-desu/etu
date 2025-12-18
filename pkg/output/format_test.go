package output

import (
	"testing"
)

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		name             string
		expectedFormat   string
		requestedFormat  string
		supportedFormats []string
		expectError      bool
	}{
		{
			name:             "supported format simple",
			requestedFormat:  "simple",
			supportedFormats: []string{"simple", "json", "table"},
			expectedFormat:   "simple",
			expectError:      false,
		},
		{
			name:             "supported format json",
			requestedFormat:  "json",
			supportedFormats: []string{"simple", "json", "table"},
			expectedFormat:   "json",
			expectError:      false,
		},
		{
			name:             "supported format table",
			requestedFormat:  "table",
			supportedFormats: []string{"simple", "json", "table"},
			expectedFormat:   "table",
			expectError:      false,
		},
		{
			name:             "fallback tree to table",
			requestedFormat:  "tree",
			supportedFormats: []string{"simple", "json", "table"},
			expectedFormat:   "table",
			expectError:      false,
		},
		{
			name:             "fallback fields to table",
			requestedFormat:  "fields",
			supportedFormats: []string{"simple", "json", "table"},
			expectedFormat:   "table",
			expectError:      false,
		},
		{
			name:             "invalid format",
			requestedFormat:  "invalid",
			supportedFormats: []string{"simple", "json", "table"},
			expectedFormat:   "",
			expectError:      true,
		},
		{
			name:             "tree supported directly",
			requestedFormat:  "tree",
			supportedFormats: []string{"simple", "json", "table", "tree"},
			expectedFormat:   "tree",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeFormat(tt.requestedFormat, tt.supportedFormats)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expectedFormat {
					t.Errorf("Expected format %s but got %s", tt.expectedFormat, result)
				}
			}
		})
	}
}
