package output

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// OutputFormat represents a supported output format.
type OutputFormat string

const (
	FormatSimple OutputFormat = "simple"
	FormatJSON   OutputFormat = "json"
	FormatTable  OutputFormat = "table"
	FormatTree   OutputFormat = "tree"
	FormatFields OutputFormat = "fields"
)

// formatSet for O(1) validation lookup.
var formatSet = map[OutputFormat]struct{}{
	FormatSimple: {},
	FormatJSON:   {},
	FormatTable:  {},
	FormatTree:   {},
	FormatFields: {},
}

// AllFormats contains all supported formats.
var AllFormats = []OutputFormat{
	FormatSimple,
	FormatJSON,
	FormatTable,
	FormatTree,
	FormatFields,
}

// String returns the string representation.
func (f OutputFormat) String() string {
	return string(f)
}

// IsValid checks if the format is supported (O(1) lookup).
func (f OutputFormat) IsValid() bool {
	_, ok := formatSet[f]
	return ok
}

// ParseFormat parses a string into OutputFormat, validating it.
// Maintains backward-compatible error message format.
func ParseFormat(s string) (OutputFormat, error) {
	f := OutputFormat(s)
	if !f.IsValid() {
		return "", fmt.Errorf("invalid output format: %s (use simple, json, table, tree, or fields)", s)
	}
	return f, nil
}

// NormalizeFormat validates and normalizes the requested output format.
// If the format is not supported, it attempts to fall back to a compatible format
// and prints a warning to stderr.
func NormalizeFormat(requestedFormat string, supportedFormats []string) (string, error) {
	// Check if requested format is supported
	if slices.Contains(supportedFormats, requestedFormat) {
		return requestedFormat, nil
	}

	// Fallback logic for unsupported formats
	fallbackMap := map[string]string{
		"tree":   "table", // Tree → Table (closest visual match)
		"fields": "table", // Fields (deprecated) → Table
	}

	fallback, hasFallback := fallbackMap[requestedFormat]
	if !hasFallback {
		return "", fmt.Errorf("invalid format: %s (valid formats: %s)",
			requestedFormat, strings.Join(supportedFormats, ", "))
	}

	// Warn user about fallback (to stderr, so it doesn't break pipes)
	fmt.Fprintf(os.Stderr, "Warning: '%s' format not supported, using '%s' instead\n",
		requestedFormat, fallback)

	return fallback, nil
}
