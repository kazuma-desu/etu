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

// Func defines a validation function that can be plugged into the validator.
type Func func(pair *models.ConfigPair, result *ValidationResult)

// KeyFormatValidator validates the format of an etcd key
func KeyFormatValidator(pair *models.ConfigPair, result *ValidationResult) {
	key := pair.Key

	if !strings.HasPrefix(key, "/") {
		result.addError(key, "key must start with '/'")
		return
	}

	if len(key) > maxKeyLength {
		result.addError(key, fmt.Sprintf("key length exceeds maximum of %d characters", maxKeyLength))
	}

	depth := strings.Count(key, "/") - 1
	if depth > maxKeyDepth {
		result.addError(key, fmt.Sprintf("key depth exceeds maximum of %d levels", maxKeyDepth))
	}

	if !validKeyRE.MatchString(key) {
		result.addError(key, "key contains invalid characters (allowed: a-z, A-Z, 0-9, /, _, -, .)")
	}
}

// ValueValidator validates the value of a configuration pair
func ValueValidator(pair *models.ConfigPair, result *ValidationResult) {
	if pair.Value == nil {
		result.addError(pair.Key, "value cannot be nil")
		return
	}

	valueStr := fmt.Sprintf("%v", pair.Value)

	if valueStr == "" {
		result.addWarning(pair.Key, "value is empty string")
	}

	size := len(valueStr)
	if size > maxValueSize {
		result.addError(pair.Key, fmt.Sprintf("value size (%d bytes) exceeds maximum of %d bytes", size, maxValueSize))
	} else if size > warnValueSize {
		result.addWarning(pair.Key, fmt.Sprintf("value size (%d bytes) exceeds recommended size of %d bytes", size, warnValueSize))
	}
}

// StructuredDataValidator validates structured data (JSON/YAML)
func StructuredDataValidator(pair *models.ConfigPair, result *ValidationResult) {
	if pair.Value == nil {
		return
	}

	valueStr := fmt.Sprintf("%v", pair.Value)

	if !looksLikeStructuredData(valueStr) {
		return
	}

	if !isValidStructuredData(valueStr) {
		result.addWarning(pair.Key, "value looks like structured data but is not valid JSON or YAML")
	}
}

// URLValidator validates URL values in keys containing "url"
func URLValidator(pair *models.ConfigPair, result *ValidationResult) {
	if pair.Value == nil {
		return
	}

	if !strings.Contains(strings.ToLower(pair.Key), "url") {
		return
	}

	str, ok := pair.Value.(string)
	if !ok {
		return
	}

	if str == "" {
		return
	}

	parsed, err := url.Parse(str)
	if err != nil {
		result.addWarning(pair.Key, fmt.Sprintf("value looks like URL but failed to parse: %v", err))
		return
	}

	if parsed.Scheme == "" {
		result.addWarning(pair.Key, "URL is missing scheme (http:// or https:// recommended)")
	} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
		result.addWarning(pair.Key, fmt.Sprintf("URL has unusual scheme: %s", parsed.Scheme))
	}
}

// looksLikeStructuredData checks if a string looks like JSON or YAML
func looksLikeStructuredData(s string) bool {
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
func isValidStructuredData(s string) bool {
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
	strict     bool // If true, treat warnings as errors
	validators []Func
}

// NewValidator creates a new validator with optional custom validators
func NewValidator(strict bool, custom ...Func) *Validator {
	validators := []Func{
		KeyFormatValidator,
		ValueValidator,
		StructuredDataValidator,
		URLValidator,
	}

	// Append custom validators
	validators = append(validators, custom...)

	return &Validator{
		strict:     strict,
		validators: validators,
	}
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

		// Run all validators
		for _, validator := range v.validators {
			validator(pair, result)
		}
	}

	// Determine if validation passed
	result.Valid = !result.HasErrors()
	if v.strict && result.HasWarnings() {
		result.Valid = false
	}

	return result
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
