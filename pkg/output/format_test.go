package output

import (
	"strings"
	"testing"
)

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name        string
		requested   string
		allowed     []string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid format json",
			requested: "json",
			allowed:   []string{"simple", "json", "table"},
			wantErr:   false,
		},
		{
			name:      "valid format simple",
			requested: "simple",
			allowed:   []string{"simple", "json", "table"},
			wantErr:   false,
		},
		{
			name:      "valid format table",
			requested: "table",
			allowed:   []string{"simple", "json", "table"},
			wantErr:   false,
		},
		{
			name:      "valid format tree",
			requested: "tree",
			allowed:   []string{"simple", "json", "table", "tree"},
			wantErr:   false,
		},
		{
			name:        "invalid format fields",
			requested:   "fields",
			allowed:     []string{"simple", "json", "table"},
			wantErr:     true,
			errContains: "invalid format: fields",
		},
		{
			name:        "invalid format xml",
			requested:   "xml",
			allowed:     []string{"simple", "json", "table"},
			wantErr:     true,
			errContains: "invalid format: xml",
		},
		{
			name:        "error includes valid formats list",
			requested:   "invalid",
			allowed:     []string{"simple", "json", "yaml"},
			wantErr:     true,
			errContains: "valid: simple, json, yaml",
		},
		{
			name:        "tree not allowed",
			requested:   "tree",
			allowed:     []string{"simple", "json", "table"},
			wantErr:     true,
			errContains: "invalid format: tree",
		},
		{
			name:        "empty allowed list",
			requested:   "json",
			allowed:     []string{},
			wantErr:     true,
			errContains: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.requested, tt.allowed)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFormat() error = nil, want error containing %q", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateFormat() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFormat() error = %v, want nil", err)
				}
			}
		})
	}
}
