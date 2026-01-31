package models

import (
	"fmt"
	"sort"
	"strings"
)

// FormatValue converts a value to a display string.
// Supports: string, int, int64, float64, map[string]any, fmt.Stringer, and nil.
// FormatValue converts a value to a human-readable string.
// For a nil input it returns an empty string. Strings are returned unchanged.
// Integer and unsigned integer types are formatted as decimal; floats use default float formatting; booleans are formatted as "true" or "false".
// For map[string]any an empty map yields an empty string; otherwise entries are rendered as "key: value" lines with keys sorted lexicographically.
// If the value implements fmt.Stringer its String method is used; all other types are formatted with the default %#v representation.
func FormatValue(val any) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint8:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%f", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case map[string]any:
		if len(v) == 0 {
			return ""
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var lines []string
		for _, k := range keys {
			lines = append(lines, fmt.Sprintf("%s: %v", k, v[k]))
		}
		return strings.Join(lines, "\n")
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}