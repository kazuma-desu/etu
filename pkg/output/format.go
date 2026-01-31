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

var allFormats = []Format{
	FormatSimple,
	FormatJSON,
	FormatTable,
	FormatTree,
	FormatFields,
}

// formatSet is derived from allFormats for O(1) validation lookup.
// Initialized in init() to ensure it stays in sync with allFormats.
var formatSet map[Format]struct{}

// init populates the internal formatSet map with all supported Format values for O(1) validation.
func init() {
	formatSet = make(map[Format]struct{}, len(allFormats))
	for _, f := range allFormats {
		formatSet[f] = struct{}{}
	}
}

// AllFormats returns a copy of the supported output formats.
// Modifying the returned slice does not affect the package's internal list.
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
// ParseFormat parses s into a Format and validates that it is one of the supported formats.
// If s is not supported, it returns an error formatted as "invalid output format: <value> (use <list>)"
// where <list> is a comma-separated list of all valid formats.
func ParseFormat(s string) (Format, error) {
	f := Format(s)
	if !f.IsValid() {
		validFormats := make([]string, len(allFormats))
		for i, format := range allFormats {
			validFormats[i] = string(format)
		}
		return "", fmt.Errorf("invalid output format: %s (use %s)", s, strings.Join(validFormats, ", "))
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