package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// FormatValue converts a value to a display string.
// Supports: string, all integer types (int, int8-64, uint, uint8-64),
// float32/64, bool, map[string]any, fmt.Stringer, and nil.
// Returns empty string for nil values.
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
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
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
