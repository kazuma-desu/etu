package validator

import (
	"strings"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator(false)
	assert.NotNil(t, v)
	assert.False(t, v.strict)

	vStrict := NewValidator(true)
	assert.NotNil(t, vStrict)
	assert.True(t, vStrict.strict)
}

func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		issues   []ValidationIssue
		expected bool
	}{
		{
			name:     "no issues",
			issues:   []ValidationIssue{},
			expected: false,
		},
		{
			name: "only warnings",
			issues: []ValidationIssue{
				{Level: "warning", Key: "/test", Message: "test warning"},
			},
			expected: false,
		},
		{
			name: "has errors",
			issues: []ValidationIssue{
				{Level: "error", Key: "/test", Message: "test error"},
			},
			expected: true,
		},
		{
			name: "mixed errors and warnings",
			issues: []ValidationIssue{
				{Level: "warning", Key: "/test1", Message: "test warning"},
				{Level: "error", Key: "/test2", Message: "test error"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{Issues: tt.issues}
			assert.Equal(t, tt.expected, result.HasErrors())
		})
	}
}

func TestValidationResult_HasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		issues   []ValidationIssue
		expected bool
	}{
		{
			name:     "no issues",
			issues:   []ValidationIssue{},
			expected: false,
		},
		{
			name: "only errors",
			issues: []ValidationIssue{
				{Level: "error", Key: "/test", Message: "test error"},
			},
			expected: false,
		},
		{
			name: "has warnings",
			issues: []ValidationIssue{
				{Level: "warning", Key: "/test", Message: "test warning"},
			},
			expected: true,
		},
		{
			name: "mixed errors and warnings",
			issues: []ValidationIssue{
				{Level: "warning", Key: "/test1", Message: "test warning"},
				{Level: "error", Key: "/test2", Message: "test error"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{Issues: tt.issues}
			assert.Equal(t, tt.expected, result.HasWarnings())
		})
	}
}

func TestValidator_ValidateKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		errorMsg    string
		expectError bool
	}{
		{
			name:        "valid simple key",
			key:         "/app/name",
			expectError: false,
		},
		{
			name:        "valid nested key",
			key:         "/app/database/host",
			expectError: false,
		},
		{
			name:        "valid key with numbers",
			key:         "/app/server1/port",
			expectError: false,
		},
		{
			name:        "valid key with dash",
			key:         "/app/my-service",
			expectError: false,
		},
		{
			name:        "valid key with underscore",
			key:         "/app/my_service",
			expectError: false,
		},
		{
			name:        "valid key with dot",
			key:         "/app/service.name",
			expectError: false,
		},
		{
			name:        "missing leading slash",
			key:         "app/name",
			expectError: true,
			errorMsg:    "must start with '/'",
		},
		{
			name:        "invalid characters - space",
			key:         "/app/my service",
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "invalid characters - special char",
			key:         "/app/name@test",
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "key too long",
			key:         "/" + strings.Repeat("a", 1001),
			expectError: true,
			errorMsg:    "exceeds maximum",
		},
		{
			name:        "key too deep",
			key:         "/" + strings.Repeat("a/", 25),
			expectError: true,
			errorMsg:    "depth exceeds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			result := &ValidationResult{Issues: []ValidationIssue{}}

			v.validateKey(tt.key, result)

			if tt.expectError {
				assert.True(t, len(result.Issues) > 0, "Expected validation error")
				if tt.errorMsg != "" {
					found := false
					for _, issue := range result.Issues {
						if strings.Contains(issue.Message, tt.errorMsg) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message containing: %s", tt.errorMsg)
				}
			} else {
				assert.Empty(t, result.Issues, "Expected no validation errors")
			}
		})
	}
}

func TestValidator_ValidateValue(t *testing.T) {
	tests := []struct {
		pair         *models.ConfigPair
		name         string
		messageMatch string
		expectError  bool
		expectWarn   bool
	}{
		{
			name:        "valid string value",
			pair:        &models.ConfigPair{Key: "/app/name", Value: "myapp"},
			expectError: false,
			expectWarn:  false,
		},
		{
			name:        "valid integer value",
			pair:        &models.ConfigPair{Key: "/app/port", Value: 8080},
			expectError: false,
			expectWarn:  false,
		},
		{
			name:         "nil value",
			pair:         &models.ConfigPair{Key: "/app/nil", Value: nil},
			expectError:  true,
			messageMatch: "cannot be nil",
		},
		{
			name:         "empty string warning",
			pair:         &models.ConfigPair{Key: "/app/empty", Value: ""},
			expectError:  false,
			expectWarn:   true,
			messageMatch: "empty string",
		},
		{
			name:         "value too large",
			pair:         &models.ConfigPair{Key: "/app/large", Value: strings.Repeat("a", 101*1024)},
			expectError:  true,
			messageMatch: "exceeds maximum",
		},
		{
			name:         "value size warning",
			pair:         &models.ConfigPair{Key: "/app/medium", Value: strings.Repeat("a", 11*1024)},
			expectError:  false,
			expectWarn:   true,
			messageMatch: "exceeds recommended size",
		},
		{
			name:        "valid JSON value",
			pair:        &models.ConfigPair{Key: "/app/config", Value: `{"key": "value"}`},
			expectError: false,
			expectWarn:  false,
		},
		{
			name:         "invalid JSON-like value",
			pair:         &models.ConfigPair{Key: "/app/config", Value: `{invalid json`},
			expectError:  false,
			expectWarn:   true,
			messageMatch: "not valid JSON or YAML",
		},
		{
			name:        "valid URL",
			pair:        &models.ConfigPair{Key: "/app/api_url", Value: "https://example.com"},
			expectError: false,
			expectWarn:  false,
		},
		{
			name:         "URL without scheme",
			pair:         &models.ConfigPair{Key: "/app/url", Value: "example.com"},
			expectError:  false,
			expectWarn:   true,
			messageMatch: "missing scheme",
		},
		{
			name:         "URL with unusual scheme",
			pair:         &models.ConfigPair{Key: "/app/url", Value: "ftp://example.com"},
			expectError:  false,
			expectWarn:   true,
			messageMatch: "unusual scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			result := &ValidationResult{Issues: []ValidationIssue{}}

			v.validateValue(tt.pair, result)

			if tt.expectError {
				assert.True(t, result.HasErrors(), "Expected validation error")
			} else {
				assert.False(t, result.HasErrors(), "Expected no validation errors")
			}

			if tt.expectWarn {
				assert.True(t, result.HasWarnings(), "Expected validation warning")
			}

			if tt.messageMatch != "" {
				found := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Message, tt.messageMatch) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected message containing: %s", tt.messageMatch)
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		pairs       []*models.ConfigPair
		strict      bool
		expectValid bool
		expectError bool
		expectWarn  bool
	}{
		{
			name:   "valid pairs",
			strict: false,
			pairs: []*models.ConfigPair{
				{Key: "/app/name", Value: "myapp"},
				{Key: "/app/version", Value: "1.0.0"},
			},
			expectValid: true,
			expectError: false,
			expectWarn:  false,
		},
		{
			name:   "duplicate keys",
			strict: false,
			pairs: []*models.ConfigPair{
				{Key: "/app/name", Value: "myapp"},
				{Key: "/app/name", Value: "otherapp"},
			},
			expectValid: false,
			expectError: true,
			expectWarn:  false,
		},
		{
			name:   "invalid key",
			strict: false,
			pairs: []*models.ConfigPair{
				{Key: "app/name", Value: "myapp"},
			},
			expectValid: false,
			expectError: true,
			expectWarn:  false,
		},
		{
			name:   "warnings in non-strict mode",
			strict: false,
			pairs: []*models.ConfigPair{
				{Key: "/app/empty", Value: ""},
			},
			expectValid: true,
			expectError: false,
			expectWarn:  true,
		},
		{
			name:   "warnings in strict mode",
			strict: true,
			pairs: []*models.ConfigPair{
				{Key: "/app/empty", Value: ""},
			},
			expectValid: false,
			expectError: false,
			expectWarn:  true,
		},
		{
			name:        "empty pairs",
			strict:      false,
			pairs:       []*models.ConfigPair{},
			expectValid: true,
			expectError: false,
			expectWarn:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(tt.strict)
			result := v.Validate(tt.pairs)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Equal(t, tt.expectError, result.HasErrors())
			assert.Equal(t, tt.expectWarn, result.HasWarnings())
		})
	}
}

func TestValidator_LooksLikeStructuredData(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"JSON object", `{"key": "value"}`, true},
		{"JSON array", `["item1", "item2"]`, true},
		{"YAML multiline", "key1: value1\nkey2: value2", true},
		{"Go map literal", "map[string]string{...}", true},
		{"Go slice literal", "[]string{...}", true},
		{"simple string", "hello world", false},
		{"number", "123", false},
		{"URL", "https://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.looksLikeStructuredData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidator_IsValidStructuredData(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid JSON", `{"key": "value"}`, true},
		{"valid JSON array", `["item1", "item2"]`, true},
		{"valid YAML", "key: value", true},
		{"simple string", "hello", true}, // YAML can parse simple strings
		{"empty", "", true},              // Empty string is valid YAML (null)
		// Note: YAML is very permissive, so most strings are valid YAML
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.isValidStructuredData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidator_ValidateURL(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		url          string
		messageMatch string
		expectWarn   bool
	}{
		{
			name:       "valid https URL",
			key:        "/app/url",
			url:        "https://example.com",
			expectWarn: false,
		},
		{
			name:       "valid http URL",
			key:        "/app/url",
			url:        "http://example.com",
			expectWarn: false,
		},
		{
			name:         "URL without scheme",
			key:          "/app/url",
			url:          "example.com",
			expectWarn:   true,
			messageMatch: "missing scheme",
		},
		{
			name:         "URL with ftp scheme",
			key:          "/app/url",
			url:          "ftp://example.com",
			expectWarn:   true,
			messageMatch: "unusual scheme",
		},
		{
			name:       "empty URL",
			key:        "/app/url",
			url:        "",
			expectWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			result := &ValidationResult{Issues: []ValidationIssue{}}

			v.validateURL(tt.key, tt.url, result)

			if tt.expectWarn {
				assert.True(t, result.HasWarnings(), "Expected validation warning")
				if tt.messageMatch != "" {
					found := false
					for _, issue := range result.Issues {
						if strings.Contains(issue.Message, tt.messageMatch) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected message containing: %s", tt.messageMatch)
				}
			} else {
				assert.False(t, result.HasWarnings(), "Expected no warnings")
			}
		})
	}
}
