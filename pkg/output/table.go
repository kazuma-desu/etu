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
		BorderStyle(tableBorderStyle).
		Headers(config.Headers...).
		Rows(config.Rows...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			// Header row (row 0 in lipgloss table is the header)
			if row == 0 {
				return tableHeaderStyle
			}

			// Alternate row colors for data rows
			if row%2 == 0 {
				return tableEvenRowStyle
			}
			return tableOddRowStyle
		})

	return t.Render()
}
