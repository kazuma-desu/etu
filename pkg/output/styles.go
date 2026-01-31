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

	// Table row colors - hex format for consistency
	colorTableOdd  = lipgloss.Color("#FCFCFA") // Light gray
	colorTableEven = lipgloss.Color("#A0A0A0") // Medium gray
)

var (
	// keyStyle renders configuration keys with high contrast for visibility.
	// Used for: config keys, tree nodes, primary identifiers.
	keyStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	// valueStyle renders configuration values in muted tone for secondary information.
	// Used for: config values, progress indicators, supporting text.
	valueStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// errorStyle renders error messages with high visibility.
	// Used for: error messages, failure indicators, critical alerts.
	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	// warningStyle renders warning messages with attention-grabbing color.
	// Used for: warning messages, caution indicators, dry-run notices.
	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	// successStyle renders success messages with positive emphasis.
	// Used for: success messages, completion indicators, confirmation.
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
				Foreground(colorTableOdd).
				Padding(0, 1)

	tableEvenRowStyle = lipgloss.NewStyle().
				Foreground(colorTableEven).
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
