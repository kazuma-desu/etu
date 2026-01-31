package output

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// Format represents a supported output format.
type Format string

const (
	FormatSimple Format = "simple"
	FormatJSON   Format = "json"
	FormatTable  Format = "table"
	FormatTree   Format = "tree"
	FormatFields Format = "fields"
)

// formatSet for O(1) validation lookup.
var formatSet = map[Format]struct{}{
	FormatSimple: {},
	FormatJSON:   {},
	FormatTable:  {},
	FormatTree:   {},
	FormatFields: {},
}

var allFormats = []Format{
	FormatSimple,
	FormatJSON,
	FormatTable,
	FormatTree,
	FormatFields,
}

// AllFormats returns a copy of all supported formats.
func AllFormats() []Format {
	out := make([]Format, len(allFormats))
	copy(out, allFormats)
	return out
}

// String returns the string representation.
func (f Format) String() string {
	return string(f)
}

// IsValid checks if the format is supported (O(1) lookup).
func (f Format) IsValid() bool {
	_, ok := formatSet[f]
	return ok
}

// ParseFormat parses a string into Format, validating it.
// Maintains backward-compatible error message format.
func ParseFormat(s string) (Format, error) {
	f := Format(s)
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
