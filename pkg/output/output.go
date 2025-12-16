package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"

	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/validator"
)

var (
	// Color palette - Modern, balanced colors
	colorPrimary   = lipgloss.Color("#7C3AED") // Purple
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorError     = lipgloss.Color("#EF4444") // Red
	colorInfo      = lipgloss.Color("#3B82F6") // Blue
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorHighlight = lipgloss.Color("#06B6D4") // Cyan

	// Styles
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

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	errorPanelStyle = panelStyle.BorderForeground(colorError)

	warningPanelStyle = panelStyle.BorderForeground(colorWarning)

	successPanelStyle = panelStyle.BorderForeground(colorSuccess)

	infoPanelStyle = panelStyle.BorderForeground(colorInfo)

	// Tree styles
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

// PrintConfigPairs prints configuration pairs in a human-readable format
func PrintConfigPairs(pairs []*models.ConfigPair, jsonOutput bool) error {
	if jsonOutput {
		return printJSON(pairs)
	}

	logger.Log.Info(fmt.Sprintf("Found %d configuration items", len(pairs)))
	fmt.Println()

	for _, pair := range pairs {
		key := keyStyle.Render(pair.Key)
		value := formatValue(pair.Value)
		fmt.Printf("%s\n%s\n\n", key, valueStyle.Render(value))
	}

	return nil
}

// PrintValidationResult prints validation results with styling
func PrintValidationResult(result *validator.ValidationResult, strict bool) {
	if len(result.Issues) == 0 {
		msg := successStyle.Render("✓ Validation passed - no issues found")
		fmt.Println(successPanelStyle.Render(msg))
		return
	}

	// Count errors and warnings
	errorCount := 0
	warningCount := 0
	for _, issue := range result.Issues {
		if issue.Level == "error" {
			errorCount++
		} else {
			warningCount++
		}
	}

	// Print summary
	var summary strings.Builder
	if errorCount > 0 {
		summary.WriteString(errorStyle.Render(fmt.Sprintf("✗ %d error(s)", errorCount)))
	}
	if warningCount > 0 {
		if summary.Len() > 0 {
			summary.WriteString(", ")
		}
		summary.WriteString(warningStyle.Render(fmt.Sprintf("⚠ %d warning(s)", warningCount)))
	}

	fmt.Println(infoPanelStyle.Render(summary.String()))
	fmt.Println()

	// Print each issue
	for _, issue := range result.Issues {
		var prefix string
		var style lipgloss.Style
		if issue.Level == "error" {
			prefix = "✗"
			style = errorStyle
		} else {
			prefix = "⚠"
			style = warningStyle
		}

		key := keyStyle.Render(issue.Key)
		msg := fmt.Sprintf("%s %s: %s", prefix, key, issue.Message)
		fmt.Println(style.Render(msg))
	}
	fmt.Println()

	// Print final verdict
	if result.Valid {
		fmt.Println(successStyle.Render("✓ Validation passed"))
	} else {
		if strict && warningCount > 0 && errorCount == 0 {
			fmt.Println(errorStyle.Render("✗ Validation failed (strict mode: warnings treated as errors)"))
		} else {
			fmt.Println(errorStyle.Render("✗ Validation failed"))
		}
	}
}

// PrintDryRun prints what would be applied in a dry run
func PrintDryRun(pairs []*models.ConfigPair) {
	title := warningStyle.Render(fmt.Sprintf("DRY RUN - Would apply %d configuration items", len(pairs)))
	fmt.Println(warningPanelStyle.Render(title))
	fmt.Println()

	for i, pair := range pairs {
		value := formatValue(pair.Value)
		key := keyStyle.Render(pair.Key)

		// Show progress indicator
		progress := fmt.Sprintf("[%d/%d]", i+1, len(pairs))
		fmt.Printf("%s %s → %s\n", valueStyle.Render(progress), successStyle.Render("PUT"), key)

		// Show the value being written
		fmt.Printf("%s\n\n", valueStyle.Render(value))
	}

	fmt.Println(warningStyle.Render("DRY RUN complete - no changes made to etcd"))
}

// PrintApplyProgress prints progress during apply operation
func PrintApplyProgress(current, total int, key string) {
	progress := fmt.Sprintf("[%d/%d]", current, total)
	k := keyStyle.Render(key)
	logger.Log.Info(fmt.Sprintf("%s Applying %s", progress, k))
}

// PrintApplySuccess prints success message after apply
func PrintApplySuccess(count int) {
	msg := successStyle.Render(fmt.Sprintf("✓ Successfully applied %d items to etcd", count))
	fmt.Println()
	fmt.Println(successPanelStyle.Render(msg))
}

// PrintError prints an error message
func PrintError(err error) {
	msg := errorStyle.Render(fmt.Sprintf("✗ Error: %v", err))
	fmt.Println(errorPanelStyle.Render(msg))
}

// formatValue converts a value to a display string
func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case map[string]any:
		var lines []string
		for k, val := range v {
			lines = append(lines, fmt.Sprintf("%s: %v", k, val))
		}
		return strings.Join(lines, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// printJSON outputs configuration as JSON
func printJSON(pairs []*models.ConfigPair) error {
	var output []map[string]any
	for _, pair := range pairs {
		output = append(output, map[string]any{
			"key":   pair.Key,
			"value": pair.Value,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// Info prints an info message
func Info(msg string) {
	fmt.Println(valueStyle.Render("⋯ " + msg))
}

// Success prints a success message
func Success(msg string) {
	fmt.Println(successStyle.Render("✓ " + msg))
}

// Error prints an error message
func Error(msg string) {
	fmt.Println(errorStyle.Render("✗ " + msg))
}

// Warning prints a warning message
func Warning(msg string) {
	fmt.Println(warningStyle.Render("⚠ " + msg))
}

// Prompt prints a styled prompt
func Prompt(msg string) {
	fmt.Print(keyStyle.Render("? ") + msg)
}

// PrintSecurityWarning prints the password storage security warning
func PrintSecurityWarning() {
	fmt.Println()
	Warning("Security Warning:")
	fmt.Println("  Your password is stored in plain text in the config file.")
	fmt.Println("  For better security:")
	fmt.Println("    - Use --password flag at runtime instead of storing it")
	fmt.Println("    - Use ETCD_PASSWORD environment variable")
	fmt.Println("    - Ensure config file permissions are restrictive (0600)")
}

// PrintTree renders etcd configuration as a tree structure
func PrintTree(pairs []*models.ConfigPair) error {
	logger.Log.Info(fmt.Sprintf("Found %d configuration items", len(pairs)))
	fmt.Println()

	t := buildEtcdTree(pairs)
	fmt.Println(t)
	return nil
}

// buildEtcdTree builds a lipgloss tree from config pairs
func buildEtcdTree(pairs []*models.ConfigPair) *tree.Tree {
	// Create root
	root := tree.Root("/").
		RootStyle(treeRootStyle).
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(treeEnumeratorStyle)

	// Build hierarchical structure
	pathMap := make(map[string]*tree.Tree)
	pathMap["/"] = root

	// Sort pairs by path for consistent output
	sorted := make([]*models.ConfigPair, len(pairs))
	copy(sorted, pairs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	for _, pair := range sorted {
		parts := strings.Split(strings.Trim(pair.Key, "/"), "/")
		currentPath := ""

		for i, part := range parts {
			if part == "" {
				continue
			}

			parentPath := currentPath
			if parentPath == "" {
				parentPath = "/"
			}
			currentPath = currentPath + "/" + part

			// Check if this path already exists
			if _, exists := pathMap[currentPath]; !exists {
				parent := pathMap[parentPath]

				// Last part = leaf node with value
				if i == len(parts)-1 {
					valueStr := formatTreeValue(pair.Value)
					display := treeKeyStyle.Render(part) + " " + treeValueStyle.Render(valueStr)
					leaf := tree.New().Root(display)
					parent.Child(leaf)
				} else {
					// Intermediate folder
					folder := tree.New().
						Root(treeFolderStyle.Render(part + "/")).
						EnumeratorStyle(treeEnumeratorStyle)
					parent.Child(folder)
					pathMap[currentPath] = folder
				}
			}
		}
	}

	return root
}

// formatTreeValue formats a value for tree display
func formatTreeValue(val any) string {
	switch v := val.(type) {
	case string:
		if strings.Contains(v, "\n") {
			lines := strings.Split(strings.TrimSpace(v), "\n")
			return "[" + fmt.Sprintf("%d lines", len(lines)) + "]"
		}
		if len(v) > 50 {
			return v[:47] + "..."
		}
		return v
	case map[string]any:
		return "[dict: " + fmt.Sprintf("%d keys", len(v)) + "]"
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 50 {
			return str[:47] + "..."
		}
		return str
	}
}
