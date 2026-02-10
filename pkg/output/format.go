package output

import (
	"fmt"
	"slices"
	"strings"
)

// Format represents a supported output format.
type Format string

const (
	FormatSimple Format = "simple"
	FormatJSON   Format = "json"
	FormatYAML   Format = "yaml"
	FormatTable  Format = "table"
	FormatTree   Format = "tree"
)

var allFormats = []Format{
	FormatSimple,
	FormatJSON,
	FormatYAML,
	FormatTable,
	FormatTree,
}

// formatSet is derived from allFormats for O(1) validation lookup.
// Initialized in init() to ensure it stays in sync with allFormats.
var formatSet map[Format]struct{}

func init() {
	formatSet = make(map[Format]struct{}, len(allFormats))
	for _, f := range allFormats {
		formatSet[f] = struct{}{}
	}
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
		validFormats := make([]string, len(allFormats))
		for i, format := range allFormats {
			validFormats[i] = string(format)
		}
		return "", fmt.Errorf("invalid output format: %s (use %s)", s, strings.Join(validFormats, ", "))
	}
	return f, nil
}

// ValidateFormat validates the requested output format against allowed formats.
// Returns an error if the format is not in the allowed list.
func ValidateFormat(requested string, allowed []string) error {
	if slices.Contains(allowed, requested) {
		return nil
	}
	return fmt.Errorf("invalid format: %s (valid: %s)",
		requested, strings.Join(allowed, ", "))
}
