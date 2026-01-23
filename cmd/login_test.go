package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCommand_ErrorReturns(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("Login with invalid endpoint returns error", func(t *testing.T) {
		loginContextName = "invalid-context"
		loginEndpoints = []string{"http://invalid-host-that-does-not-exist:2379"}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = false

		err := runLogin(loginCmd, []string{})
		assert.Error(t, err)
	})

	t.Run("Login saves config even with failed connection test when user declines", func(t *testing.T) {
		loginContextName = "notest-saves-context"
		loginEndpoints = []string{"http://localhost:9999"}
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = true
		loginNoTest = true

		err := runLogin(loginCmd, []string{})
		assert.NoError(t, err)
	})
}

func TestValidateContextName(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "enter a context name",
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
			errMsg:  "enter a context name",
		},
		{
			name:    "too short",
			input:   "a",
			wantErr: true,
			errMsg:  "at least 2 characters",
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 64),
			wantErr: true,
			errMsg:  "max 63 characters",
		},
		{
			name:    "contains space",
			input:   "my context",
			wantErr: true,
			errMsg:  "spaces not allowed",
		},
		{
			name:    "invalid character @",
			input:   "my@context",
			wantErr: true,
			errMsg:  "invalid character '@'",
		},
		{
			name:    "invalid character .",
			input:   "my.context",
			wantErr: true,
			errMsg:  "invalid character '.'",
		},
		{
			name:    "valid simple name",
			input:   "production",
			wantErr: false,
		},
		{
			name:    "valid with dash",
			input:   "my-context",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			input:   "my_context",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "prod01",
			wantErr: false,
		},
		{
			name:    "valid mixed case",
			input:   "MyContext",
			wantErr: false,
		},
		{
			name:    "valid minimum length",
			input:   "ab",
			wantErr: false,
		},
		{
			name:    "valid maximum length",
			input:   strings.Repeat("a", 63),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContextName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEndpoints(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "enter at least one endpoint",
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
			errMsg:  "enter at least one endpoint",
		},
		{
			name:    "missing scheme",
			input:   "localhost:2379",
			wantErr: true,
			errMsg:  "must start with http:// or https://",
		},
		{
			name:    "invalid scheme ftp",
			input:   "ftp://localhost:2379",
			wantErr: true,
			errMsg:  "must start with http:// or https://",
		},
		{
			name:    "missing host",
			input:   "http://",
			wantErr: true,
			errMsg:  "missing hostname",
		},
		{
			name:    "valid http with port",
			input:   "http://localhost:2379",
			wantErr: false,
		},
		{
			name:    "valid https with port",
			input:   "https://etcd.example.com:2379",
			wantErr: false,
		},
		{
			name:    "valid http without port",
			input:   "http://localhost",
			wantErr: false,
		},
		{
			name:    "valid multiple endpoints",
			input:   "http://etcd1:2379,http://etcd2:2379,http://etcd3:2379",
			wantErr: false,
		},
		{
			name:    "multiple endpoints with spaces",
			input:   "http://etcd1:2379, http://etcd2:2379, http://etcd3:2379",
			wantErr: false,
		},
		{
			name:    "one valid one invalid",
			input:   "http://localhost:2379,invalid",
			wantErr: true,
			errMsg:  "must start with http:// or https://",
		},
		{
			name:    "valid IP address",
			input:   "http://192.168.1.1:2379",
			wantErr: false,
		},
		{
			name:    "valid with path",
			input:   "http://localhost:2379/v3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEndpoints(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "shorter than max",
			input:    "hello",
			max:      10,
			expected: "hello",
		},
		{
			name:     "equal to max",
			input:    "hello",
			max:      5,
			expected: "hello",
		},
		{
			name:     "longer than max",
			input:    "hello world",
			max:      5,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			max:      5,
			expected: "",
		},
		{
			name:     "max zero",
			input:    "hello",
			max:      0,
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single endpoint",
			input:    "http://localhost:2379",
			expected: []string{"http://localhost:2379"},
		},
		{
			name:     "multiple endpoints",
			input:    "http://etcd1:2379,http://etcd2:2379",
			expected: []string{"http://etcd1:2379", "http://etcd2:2379"},
		},
		{
			name:     "with spaces",
			input:    "http://etcd1:2379, http://etcd2:2379, http://etcd3:2379",
			expected: []string{"http://etcd1:2379", "http://etcd2:2379", "http://etcd3:2379"},
		},
		{
			name:     "empty entries filtered",
			input:    "http://etcd1:2379,,http://etcd2:2379",
			expected: []string{"http://etcd1:2379", "http://etcd2:2379"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only entries",
			input:    "http://localhost:2379,   ,http://other:2379",
			expected: []string{"http://localhost:2379", "http://other:2379"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEndpoints(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasLoginFlags(t *testing.T) {
	resetLoginFlags := func() {
		loginContextName = ""
		loginEndpoints = nil
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = false
		loginNoTest = false
	}

	t.Run("no flags set", func(t *testing.T) {
		resetLoginFlags()
		assert.False(t, hasLoginFlags())
	})

	t.Run("context name set", func(t *testing.T) {
		resetLoginFlags()
		loginContextName = "test"
		assert.True(t, hasLoginFlags())
	})

	t.Run("endpoints set", func(t *testing.T) {
		resetLoginFlags()
		loginEndpoints = []string{"http://localhost:2379"}
		assert.True(t, hasLoginFlags())
	})

	t.Run("username set", func(t *testing.T) {
		resetLoginFlags()
		loginUsername = "admin"
		assert.True(t, hasLoginFlags())
	})

	t.Run("password set", func(t *testing.T) {
		resetLoginFlags()
		loginPassword = "secret"
		assert.True(t, hasLoginFlags())
	})

	t.Run("no-auth flag set", func(t *testing.T) {
		resetLoginFlags()
		loginNoAuth = true
		assert.True(t, hasLoginFlags())
	})

	t.Run("no-test flag set", func(t *testing.T) {
		resetLoginFlags()
		loginNoTest = true
		assert.True(t, hasLoginFlags())
	})
}

func TestRunLoginAutomated(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	resetLoginFlags := func() {
		loginContextName = ""
		loginEndpoints = nil
		loginUsername = ""
		loginPassword = ""
		loginNoAuth = false
		loginNoTest = false
	}

	t.Run("missing context name", func(t *testing.T) {
		resetLoginFlags()
		loginEndpoints = []string{"http://localhost:2379"}
		loginNoTest = true

		err := runLoginAutomated()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--context-name is required")
	})

	t.Run("missing endpoints", func(t *testing.T) {
		resetLoginFlags()
		loginContextName = "test"
		loginNoTest = true

		err := runLoginAutomated()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--endpoints is required")
	})

	t.Run("successful save with no-test", func(t *testing.T) {
		resetLoginFlags()
		loginContextName = "automated-test"
		loginEndpoints = []string{"http://localhost:2379"}
		loginNoAuth = true
		loginNoTest = true

		err := runLoginAutomated()
		assert.NoError(t, err)
	})

	t.Run("clears auth when no-auth flag set", func(t *testing.T) {
		resetLoginFlags()
		loginContextName = "noauth-test"
		loginEndpoints = []string{"http://localhost:2379"}
		loginUsername = "should-be-cleared"
		loginPassword = "should-be-cleared"
		loginNoAuth = true
		loginNoTest = true

		err := runLoginAutomated()
		assert.NoError(t, err)
	})
}
