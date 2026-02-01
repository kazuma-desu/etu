package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
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

func TestWrapContextError(t *testing.T) {
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
			name:    "context canceled gets wrapped",
			err:     context.Canceled,
			wantNil: false,
			wantMsg: "operation canceled by user",
		},
		{
			name:    "other errors pass through",
			err:     errors.New("some other error"),
			wantNil: false,
			wantMsg: "some other error",
		},
		{
			name:    "wrapped deadline exceeded is detected",
			err:     fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
			wantNil: false,
			wantMsg: "operation timed out",
		},
		{
			name:    "wrapped context canceled is detected",
			err:     fmt.Errorf("wrapped: %w", context.Canceled),
			wantNil: false,
			wantMsg: "operation canceled by user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapContextError(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("wrapContextError() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("wrapContextError() = nil, want error containing %q", tt.wantMsg)
				return
			}
			if !strings.Contains(got.Error(), tt.wantMsg) {
				t.Errorf("wrapContextError() = %v, want error containing %q", got, tt.wantMsg)
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

func TestApplyGlobalOverrides_MutuallyExclusivePasswordFlags(t *testing.T) {
	original := struct {
		password      string
		passwordStdin bool
	}{globalPassword, globalPasswordStdin}
	defer func() {
		globalPassword = original.password
		globalPasswordStdin = original.passwordStdin
	}()

	globalPassword = "secret"
	globalPasswordStdin = true

	cfg := &client.Config{}
	err := applyGlobalOverrides(cfg)

	if err == nil {
		t.Fatalf("applyGlobalOverrides() should error when both --password and --password-stdin are set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("applyGlobalOverrides() error = %v, want error containing 'mutually exclusive'", err)
	}
}

func TestApplyGlobalOverrides_PasswordFlag(t *testing.T) {
	original := struct {
		password      string
		passwordStdin bool
		username      string
	}{globalPassword, globalPasswordStdin, globalUsername}
	defer func() {
		globalPassword = original.password
		globalPasswordStdin = original.passwordStdin
		globalUsername = original.username
	}()

	globalPassword = "secret123"
	globalPasswordStdin = false
	globalUsername = "admin"

	cfg := &client.Config{}
	err := applyGlobalOverrides(cfg)

	if err != nil {
		t.Errorf("applyGlobalOverrides() error = %v, want nil", err)
	}
	if cfg.Password != "secret123" {
		t.Errorf("cfg.Password = %v, want 'secret123'", cfg.Password)
	}
	if cfg.Username != "admin" {
		t.Errorf("cfg.Username = %v, want 'admin'", cfg.Username)
	}
}

func TestApplyGlobalOverrides_TLSFlags(t *testing.T) {
	original := struct {
		cacert                string
		cert                  string
		key                   string
		insecureSkipTLSVerify bool
	}{globalCACert, globalCert, globalKey, globalInsecureSkipTLSVerify}
	defer func() {
		globalCACert = original.cacert
		globalCert = original.cert
		globalKey = original.key
		globalInsecureSkipTLSVerify = original.insecureSkipTLSVerify
	}()

	globalCACert = "/path/to/ca.crt"
	globalCert = "/path/to/client.crt"
	globalKey = "/path/to/client.key"
	globalInsecureSkipTLSVerify = true

	cfg := &client.Config{}
	err := applyGlobalOverrides(cfg)

	if err != nil {
		t.Errorf("applyGlobalOverrides() error = %v, want nil", err)
	}
	if cfg.CACert != "/path/to/ca.crt" {
		t.Errorf("cfg.CACert = %v, want '/path/to/ca.crt'", cfg.CACert)
	}
	if cfg.Cert != "/path/to/client.crt" {
		t.Errorf("cfg.Cert = %v, want '/path/to/client.crt'", cfg.Cert)
	}
	if cfg.Key != "/path/to/client.key" {
		t.Errorf("cfg.Key = %v, want '/path/to/client.key'", cfg.Key)
	}
	if !cfg.InsecureSkipTLSVerify {
		t.Error("cfg.InsecureSkipTLSVerify = false, want true")
	}
}

func TestLoadAppConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("ETUCONFIG", tmpDir+"/config.yaml")

	// Test when config doesn't exist - should return nil without error
	cfg := loadAppConfig()
	if cfg != nil && len(cfg.Contexts) > 0 {
		t.Error("loadAppConfig() returned non-empty config for missing file")
	}
}

func TestLoadAppConfig_WithValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"
	t.Setenv("ETUCONFIG", configPath)

	// Create a valid config
	testCfg := &config.Config{
		DefaultFormat: "etcdctl",
		Strict:        true,
		Contexts:      map[string]*config.ContextConfig{},
	}
	if err := config.SaveConfig(testCfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	cfg := loadAppConfig()
	if cfg == nil {
		t.Fatal("loadAppConfig() returned nil for valid config")
	}
	if cfg.DefaultFormat != "etcdctl" {
		t.Errorf("loadAppConfig().DefaultFormat = %v, want 'etcdctl'", cfg.DefaultFormat)
	}
	if !cfg.Strict {
		t.Error("loadAppConfig().Strict = false, want true")
	}
}

func TestGetParserForFile(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		format     models.FormatType
		wantFormat models.FormatType
		wantErr    bool
	}{
		{
			name:       "explicit yaml format",
			filePath:   "test.yaml",
			format:     models.FormatYAML,
			wantFormat: models.FormatYAML,
		},
		{
			name:       "explicit json format",
			filePath:   "test.json",
			format:     models.FormatJSON,
			wantFormat: models.FormatJSON,
		},
		{
			name:       "explicit etcdctl format",
			filePath:   "test.txt",
			format:     models.FormatEtcdctl,
			wantFormat: models.FormatEtcdctl,
		},
		{
			name:     "invalid format",
			filePath: "test.txt",
			format:   models.FormatType("invalid"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, format, err := getParserForFile(tt.filePath, tt.format)
			if tt.wantErr {
				if err == nil {
					t.Errorf("getParserForFile() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Errorf("getParserForFile() error = %v, want nil", err)
				return
			}
			if parser == nil {
				t.Error("getParserForFile() parser = nil, want non-nil")
			}
			if format != tt.wantFormat {
				t.Errorf("getParserForFile() format = %v, want %v", format, tt.wantFormat)
			}
		})
	}
}

func TestLogVerbose(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"quiet with json format", "json"},
		{"verbose with simple format", "simple"},
		{"verbose with table format", "table"},
		{"verbose with empty format", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			original := outputFormat
			defer func() { outputFormat = original }()

			outputFormat = tt.format
			// Should not panic
			logVerbose("test message", "key", "value")
		})
	}
}

func TestLogVerboseInfo(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"quiet with json format", "json"},
		{"verbose with simple format", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			original := outputFormat
			defer func() { outputFormat = original }()

			outputFormat = tt.format
			// Should not panic
			logVerboseInfo("test message")
		})
	}
}

func TestNewEtcdClientOrDryRun_DryRun(t *testing.T) {
	client, cleanup, err := newEtcdClientOrDryRun(true)
	if err != nil {
		t.Fatalf("newEtcdClientOrDryRun(true) error = %v, want nil", err)
	}
	if client == nil {
		t.Error("newEtcdClientOrDryRun(true) client = nil, want non-nil")
	}
	if cleanup == nil {
		t.Error("newEtcdClientOrDryRun(true) cleanup = nil, want non-nil")
	}
	// Should not panic
	cleanup()
}

func TestOperationTimeoutBehavior(t *testing.T) {
	originalTimeout := operationTimeout
	defer func() {
		operationTimeout = originalTimeout
	}()

	operationTimeout = 100 * time.Millisecond

	ctx, cancel := getOperationContext()
	defer cancel()

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		t.Fatal("context should have a deadline")
	}

	expectedDeadline := time.Now().Add(operationTimeout)
	deadlineDiff := deadline.Sub(expectedDeadline)
	if deadlineDiff > time.Second || deadlineDiff < -time.Second {
		t.Errorf("deadline should be approximately %v from now, got diff %v", operationTimeout, deadlineDiff)
	}
}

func BenchmarkWrapContextError(b *testing.B) {
	errs := []error{
		nil,
		context.DeadlineExceeded,
		context.Canceled,
		errors.New("some error"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapContextError(errs[i%len(errs)])
	}
}
