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
// RenderTable builds a lipgloss table from the provided TableConfig and renders it as a string.
// When running in a terminal, a rounded border and per-row/header styles are applied; otherwise the table is rendered without border or coloring.
// The table uses config.Headers for header titles and config.Rows for data rows.
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