package output

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

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
