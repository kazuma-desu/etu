package output

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// TableConfig holds table configuration
type TableConfig struct {
	Headers []string
	Rows    [][]string
}

// RenderTable creates a styled lipgloss table with rounded borders and alternating row colors
// All table styles are defined in styles.go
func RenderTable(config TableConfig) string {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		Headers(config.Headers...).
		Rows(config.Rows...)

	if IsTerminal() {
		t = t.BorderStyle(tableBorderStyle).
			StyleFunc(func(row, _ int) lipgloss.Style {
				if row == 0 {
					return tableHeaderStyle
				}
				if row%2 == 0 {
					return tableEvenRowStyle
				}
				return tableOddRowStyle
			})
	}

	return t.Render()
}
