package output

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// IsTerminal returns true if stdout is a terminal (TTY).
// Uses go-isatty for cross-platform detection including Windows ConPTY.
func IsTerminal() bool {
	fd := os.Stdout.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// StyleIfTerminal applies the style only if output is to a terminal.
func StyleIfTerminal(style lipgloss.Style, content string) string {
	if IsTerminal() {
		return style.Render(content)
	}
	return content
}
