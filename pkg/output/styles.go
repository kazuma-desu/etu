package output

import "github.com/charmbracelet/lipgloss"

// Color palette - Modern, balanced colors
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // Purple
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorError     = lipgloss.Color("#EF4444") // Red
	colorInfo      = lipgloss.Color("#3B82F6") // Blue
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorHighlight = lipgloss.Color("#06B6D4") // Cyan
)

// Text styles
var (
	keyStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)
)

// Panel styles
var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	errorPanelStyle = panelStyle.BorderForeground(colorError)

	warningPanelStyle = panelStyle.BorderForeground(colorWarning)

	successPanelStyle = panelStyle.BorderForeground(colorSuccess)

	infoPanelStyle = panelStyle.BorderForeground(colorInfo)
)

// Tree styles
var (
	treeRootStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	treeFolderStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	treeKeyStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	treeValueStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	treeEnumeratorStyle = lipgloss.NewStyle().
				Foreground(colorMuted)
)

// Table styles
var (
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				Padding(0, 1)

	tableOddRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	tableEvenRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("246")).
				Padding(0, 1)

	tableBorderStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)
)

// Diff styles
var (
	addedStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	modifiedStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	deletedStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	unchangedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	oldValueStyle = lipgloss.NewStyle().
			Foreground(colorError)

	newValueStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)
)
