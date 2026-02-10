package output

import "strings"

// Truncate truncates a string to maxLen characters, appending "..." if truncated.
// It properly handles Unicode characters by operating on runes.
// Newline characters (\n) are escaped to \n for display safety.
// If maxLen <= 0, returns empty string.
// If maxLen <= 3, returns the first maxLen characters without "...".
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	s = strings.ReplaceAll(s, "\n", "\\n")
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
