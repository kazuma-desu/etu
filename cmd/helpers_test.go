package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/models"
)

func TestResolveFormat(t *testing.T) {
	tests := []struct {
		name       string
		flagFormat models.FormatType
		appCfg     *config.Config
		want       models.FormatType
	}{
		{
			name:       "flag takes priority over config",
			flagFormat: models.FormatEtcdctl,
			appCfg:     &config.Config{DefaultFormat: "auto"},
			want:       models.FormatEtcdctl,
		},
		{
			name:       "config used when flag empty",
			flagFormat: "",
			appCfg:     &config.Config{DefaultFormat: "etcdctl"},
			want:       models.FormatEtcdctl,
		},
		{
			name:       "defaults to auto when both empty",
			flagFormat: "",
			appCfg:     &config.Config{},
			want:       models.FormatAuto,
		},
		{
			name:       "handles nil config",
			flagFormat: "",
			appCfg:     nil,
			want:       models.FormatAuto,
		},
		{
			name:       "flag priority with nil config",
			flagFormat: models.FormatEtcdctl,
			appCfg:     nil,
			want:       models.FormatEtcdctl,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFormat(tt.flagFormat, tt.appCfg)
			if got != tt.want {
				t.Errorf("resolveFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapTimeoutError(t *testing.T) {
	tests := []struct {
		name    string
		wantMsg string
		err     error
		wantNil bool
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:    "deadline exceeded gets wrapped",
			err:     context.DeadlineExceeded,
			wantNil: false,
			wantMsg: "operation timed out",
		},
		{
			name:    "other errors pass through",
			err:     errors.New("some other error"),
			wantNil: false,
			wantMsg: "some other error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapTimeoutError(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("wrapTimeoutError() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("wrapTimeoutError() = nil, want error containing %q", tt.wantMsg)
				return
			}
			if !strings.Contains(got.Error(), tt.wantMsg) {
				t.Errorf("wrapTimeoutError() = %v, want error containing %q", got, tt.wantMsg)
			}
		})
	}
}

func TestGetOperationContext(t *testing.T) {
	ctx, cancel := getOperationContext()
	defer cancel()

	if ctx == nil {
		t.Error("getOperationContext() returned nil context")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("getOperationContext() context has no deadline")
	}
	if deadline.IsZero() {
		t.Error("getOperationContext() deadline is zero")
	}
}

func TestResolveStrictOption(t *testing.T) {
	tests := []struct {
		name        string
		appCfg      *config.Config
		flagValue   bool
		flagChanged bool
		want        bool
	}{
		{
			name:        "flag true overrides config false",
			flagValue:   true,
			flagChanged: true,
			appCfg:      &config.Config{Strict: false},
			want:        true,
		},
		{
			name:        "flag false overrides config true",
			flagValue:   false,
			flagChanged: true,
			appCfg:      &config.Config{Strict: true},
			want:        false,
		},
		{
			name:        "uses config when flag unchanged",
			flagValue:   false,
			flagChanged: false,
			appCfg:      &config.Config{Strict: true},
			want:        true,
		},
		{
			name:        "defaults to false with nil config",
			flagValue:   false,
			flagChanged: false,
			appCfg:      nil,
			want:        false,
		},
		{
			name:        "flag overrides nil config",
			flagValue:   true,
			flagChanged: true,
			appCfg:      nil,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveStrictOption(tt.flagValue, tt.flagChanged, tt.appCfg)
			if got != tt.want {
				t.Errorf("resolveStrictOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveNoValidateOption(t *testing.T) {
	tests := []struct {
		name        string
		appCfg      *config.Config
		flagValue   bool
		flagChanged bool
		want        bool
	}{
		{
			name:        "flag true overrides config false",
			flagValue:   true,
			flagChanged: true,
			appCfg:      &config.Config{NoValidate: false},
			want:        true,
		},
		{
			name:        "flag false overrides config true",
			flagValue:   false,
			flagChanged: true,
			appCfg:      &config.Config{NoValidate: true},
			want:        false,
		},
		{
			name:        "uses config when flag unchanged",
			flagValue:   false,
			flagChanged: false,
			appCfg:      &config.Config{NoValidate: true},
			want:        true,
		},
		{
			name:        "defaults to false with nil config",
			flagValue:   false,
			flagChanged: false,
			appCfg:      nil,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveNoValidateOption(tt.flagValue, tt.flagChanged, tt.appCfg)
			if got != tt.want {
				t.Errorf("resolveNoValidateOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsQuietOutput(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   bool
	}{
		{
			name:   "json is quiet",
			format: "json",
			want:   true,
		},
		{
			name:   "simple is not quiet",
			format: "simple",
			want:   false,
		},
		{
			name:   "table is not quiet",
			format: "table",
			want:   false,
		},
		{
			name:   "empty is not quiet",
			format: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := outputFormat
			defer func() { outputFormat = original }()

			outputFormat = tt.format
			got := isQuietOutput()
			if got != tt.want {
				t.Errorf("isQuietOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeOutputFormat(t *testing.T) {
	tests := []struct {
		supportedFormats []string
		name             string
		format           string
		want             string
		wantErr          bool
	}{
		{
			name:             "valid format passes through",
			format:           "json",
			supportedFormats: []string{"simple", "json", "table"},
			want:             "json",
			wantErr:          false,
		},
		{
			name:             "tree with tree support",
			format:           "tree",
			supportedFormats: []string{"simple", "json", "table", "tree"},
			want:             "tree",
			wantErr:          false,
		},
		{
			name:             "tree without tree support falls back to table",
			format:           "tree",
			supportedFormats: []string{"simple", "json", "table"},
			want:             "table",
			wantErr:          false,
		},
		{
			name:             "invalid format errors",
			format:           "invalid",
			supportedFormats: []string{"simple", "json", "table"},
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := outputFormat
			defer func() { outputFormat = original }()

			outputFormat = tt.format
			got, err := normalizeOutputFormat(tt.supportedFormats)
			if tt.wantErr {
				if err == nil {
					t.Errorf("normalizeOutputFormat() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Errorf("normalizeOutputFormat() error = %v, want nil", err)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeOutputFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
