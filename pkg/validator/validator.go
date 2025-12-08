package validator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kazuma-desu/etu/pkg/models"

	"gopkg.in/yaml.v3"
)

const (
	maxKeyLength  = 1000
	maxKeyDepth   = 20
	maxValueSize  = 100 * 1024 // 100KB
	warnValueSize = 10 * 1024  // 10KB
)

var (
	validKeyRE = regexp.MustCompile(`^/[a-zA-Z0-9/_\-\.]+$`)
)

// ValidationIssue represents a single validation issue
type ValidationIssue struct {
	Key     string
	Message string
	Level   string // "error" or "warning"
}

// ValidationResult contains the results of validation
type ValidationResult struct {
	Issues []ValidationIssue
	Valid  bool
}

// HasErrors returns true if there are any error-level issues
func (v *ValidationResult) HasErrors() bool {
	for _, issue := range v.Issues {
		if issue.Level == "error" {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warning-level issues
func (v *ValidationResult) HasWarnings() bool {
	for _, issue := range v.Issues {
		if issue.Level == "warning" {
			return true
		}
	}
	return false
}

// Validator validates etcd configuration pairs
type Validator struct {
	strict bool // If true, treat warnings as errors
}

// NewValidator creates a new validator
func NewValidator(strict bool) *Validator {
	return &Validator{strict: strict}
}

// Validate validates a slice of configuration pairs
func (v *Validator) Validate(pairs []*models.ConfigPair) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Issues: []ValidationIssue{},
	}

	seenKeys := make(map[string]bool)

	for _, pair := range pairs {
		// Check for duplicates
		if seenKeys[pair.Key] {
			result.addError(pair.Key, "duplicate key found")
			continue
		}
		seenKeys[pair.Key] = true

		// Validate key
		v.validateKey(pair.Key, result)

		// Validate value
		v.validateValue(pair, result)
	}

	// Determine if validation passed
	result.Valid = !result.HasErrors()
	if v.strict && result.HasWarnings() {
		result.Valid = false
	}

	return result
}

// validateKey validates an etcd key
func (v *Validator) validateKey(key string, result *ValidationResult) {
	// Must start with /
	if !strings.HasPrefix(key, "/") {
		result.addError(key, "key must start with '/'")
		return
	}

	// Check key length
	if len(key) > maxKeyLength {
		result.addError(key, fmt.Sprintf("key length exceeds maximum of %d characters", maxKeyLength))
	}

	// Check key depth
	depth := strings.Count(key, "/") - 1
	if depth > maxKeyDepth {
		result.addError(key, fmt.Sprintf("key depth exceeds maximum of %d levels", maxKeyDepth))
	}

	// Check valid characters
	if !validKeyRE.MatchString(key) {
		result.addError(key, "key contains invalid characters (allowed: a-z, A-Z, 0-9, /, _, -, .)")
	}
}

// validateValue validates a configuration value
func (v *Validator) validateValue(pair *models.ConfigPair, result *ValidationResult) {
	// Check for nil value
	if pair.Value == nil {
		result.addError(pair.Key, "value cannot be nil")
		return
	}

	// Convert value to string for size check
	valueStr := fmt.Sprintf("%v", pair.Value)

	// Check if value is empty string
	if valueStr == "" {
		result.addWarning(pair.Key, "value is empty string")
	}

	// Check value size
	size := len(valueStr)
	if size > maxValueSize {
		result.addError(pair.Key, fmt.Sprintf("value size (%d bytes) exceeds maximum of %d bytes", size, maxValueSize))
	} else if size > warnValueSize {
		result.addWarning(pair.Key, fmt.Sprintf("value size (%d bytes) exceeds recommended size of %d bytes", size, warnValueSize))
	}

	// Validate structured data (JSON/YAML)
	if v.looksLikeStructuredData(valueStr) {
		if !v.isValidStructuredData(valueStr) {
			result.addWarning(pair.Key, "value looks like structured data but is not valid JSON or YAML")
		}
	}

	// Validate URLs
	if strings.Contains(strings.ToLower(pair.Key), "url") {
		if str, ok := pair.Value.(string); ok {
			v.validateURL(pair.Key, str, result)
		}
	}
}

// looksLikeStructuredData checks if a string looks like JSON or YAML
func (v *Validator) looksLikeStructuredData(s string) bool {
	s = strings.TrimSpace(s)
	// Check for JSON
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return true
	}
	// Check for YAML multi-line with key: value pattern
	if strings.Contains(s, "\n") && strings.Contains(s, ":") {
		return true
	}
	// Check for Go map/slice literals
	if strings.HasPrefix(s, "map[") || strings.HasPrefix(s, "[]") {
		return true
	}
	return false
}

// isValidStructuredData checks if data is valid JSON or YAML
func (v *Validator) isValidStructuredData(s string) bool {
	// Try JSON
	var jsonData any
	if err := json.Unmarshal([]byte(s), &jsonData); err == nil {
		return true
	}

	// Try YAML
	var yamlData any
	if err := yaml.Unmarshal([]byte(s), &yamlData); err == nil {
		return true
	}

	return false
}

// validateURL validates a URL value
func (v *Validator) validateURL(key, urlStr string, result *ValidationResult) {
	if urlStr == "" {
		return
	}

	// Try parsing as-is
	parsed, err := url.Parse(urlStr)
	if err != nil {
		result.addWarning(key, fmt.Sprintf("value looks like URL but failed to parse: %v", err))
		return
	}

	// Warn if scheme is missing
	if parsed.Scheme == "" {
		result.addWarning(key, "URL is missing scheme (http:// or https:// recommended)")
	} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
		result.addWarning(key, fmt.Sprintf("URL has unusual scheme: %s", parsed.Scheme))
	}
}

// addError adds an error-level issue
func (v *ValidationResult) addError(key, message string) {
	v.Issues = append(v.Issues, ValidationIssue{
		Level:   "error",
		Key:     key,
		Message: message,
	})
}

// addWarning adds a warning-level issue
func (v *ValidationResult) addWarning(key, message string) {
	v.Issues = append(v.Issues, ValidationIssue{
		Level:   "warning",
		Key:     key,
		Message: message,
	})
}
