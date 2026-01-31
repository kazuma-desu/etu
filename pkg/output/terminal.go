package output

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// IsTerminal returns true if stdout is a terminal (TTY).
// It returns true for POSIX terminals and for Windows ConPTY/Cygwin terminals.
func IsTerminal() bool {
	fd := os.Stdout.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// StyleIfTerminal applies the provided lipgloss.Style to the content when stdout is a terminal.
// If stdout is not a terminal, it returns the content unchanged.
func StyleIfTerminal(style lipgloss.Style, content string) string {
	if IsTerminal() {
		return style.Render(content)
	}
	return content
}