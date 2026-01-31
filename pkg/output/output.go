package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/validator"
)

// All styles are now defined in styles.go

// PrintConfigPairs prints configuration pairs to stdout in a styled, human-readable layout.
// If jsonOutput is true, it writes the pairs as indented JSON; otherwise it prints each key
// followed by its formatted value on separate lines. It returns any error encountered while
// producing JSON output.
func PrintConfigPairs(pairs []*models.ConfigPair, jsonOutput bool) error {
	if jsonOutput {
		return printJSON(pairs)
	}

	logger.Log.Info(fmt.Sprintf("Found %d configuration items", len(pairs)))
	fmt.Println()

	for _, pair := range pairs {
		key := StyleIfTerminal(keyStyle, pair.Key)
		value := formatValue(pair.Value)
		fmt.Printf("%s\n%s\n\n", key, StyleIfTerminal(valueStyle, value))
	}

	return nil
}

// PrintConfigPairsWithFormat prints configuration pairs using the specified format.
// Supported formats are "simple", "json", "table", and "tree".
// Returns an error if the provided format is not supported.
func PrintConfigPairsWithFormat(pairs []*models.ConfigPair, format string) error {
	switch format {
	case FormatSimple.String():
		return PrintConfigPairs(pairs, false)
	case FormatJSON.String():
		return printJSON(pairs)
	case FormatTable.String():
		return printConfigPairsTable(pairs)
	case FormatTree.String():
		return PrintTree(pairs)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// printConfigPairsTable prints config pairs as a table
func printConfigPairsTable(pairs []*models.ConfigPair) error {
	if len(pairs) == 0 {
		Info("No configuration items found")
		return nil
	}

	headers := []string{"KEY", "VALUE"}
	rows := make([][]string, len(pairs))

	for i, pair := range pairs {
		value := formatValue(pair.Value)
		// Truncate long values for table display
		if len(value) > 60 {
			value = value[:57] + "..."
		}
		rows[i] = []string{pair.Key, value}
	}

	table := RenderTable(TableConfig{
		Headers: headers,
		Rows:    rows,
	})

	fmt.Println(table)
	return nil
}

// PrintValidationResult prints validation results to stdout using styled sections.
// It displays a success panel when there are no issues; otherwise it prints a summary of error and warning counts, each issue on its own line prefixed by its severity and key, and a final verdict message. When `strict` is true, warnings are treated as failures if there are no errors.
func PrintValidationResult(result *validator.ValidationResult, strict bool) {
	if len(result.Issues) == 0 {
		msg := StyleIfTerminal(successStyle, "✓ Validation passed - no issues found")
		fmt.Println(StyleIfTerminal(successPanelStyle, msg))
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
		summary.WriteString(StyleIfTerminal(errorStyle, fmt.Sprintf("✗ %d error(s)", errorCount)))
	}
	if warningCount > 0 {
		if summary.Len() > 0 {
			summary.WriteString(", ")
		}
		summary.WriteString(StyleIfTerminal(warningStyle, fmt.Sprintf("⚠ %d warning(s)", warningCount)))
	}

	fmt.Println(StyleIfTerminal(infoPanelStyle, summary.String()))
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

		key := StyleIfTerminal(keyStyle, issue.Key)
		msg := fmt.Sprintf("%s %s: %s", prefix, key, issue.Message)
		fmt.Println(StyleIfTerminal(style, msg))
	}
	fmt.Println()

	// Print final verdict
	if result.Valid {
		fmt.Println(StyleIfTerminal(successStyle, "✓ Validation passed"))
	} else {
		if strict && warningCount > 0 && errorCount == 0 {
			fmt.Println(StyleIfTerminal(errorStyle, "✗ Validation failed (strict mode: warnings treated as errors)"))
		} else {
			fmt.Println(StyleIfTerminal(errorStyle, "✗ Validation failed"))
		}
	}
}

// PrintValidationWithFormat prints validation results using the specified output format.
// Supported formats are "simple", "json", and "table". The `strict` flag causes warnings
// to be treated as failures when true. Returns an error if the format is unsupported or
// if the chosen formatter fails.
func PrintValidationWithFormat(result *validator.ValidationResult, strict bool, format string) error {
	switch format {
	case FormatSimple.String():
		PrintValidationResult(result, strict)
		return nil
	case FormatJSON.String():
		return printValidationJSON(result, strict)
	case FormatTable.String():
		return printValidationTable(result, strict)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// printValidationJSON prints validation results as JSON
func printValidationJSON(result *validator.ValidationResult, strict bool) error {
	output := map[string]interface{}{
		"valid":  result.Valid,
		"strict": strict,
		"issues": result.Issues,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// printValidationTable prints validation results as a table
func printValidationTable(result *validator.ValidationResult, strict bool) error {
	if len(result.Issues) == 0 {
		Info("✓ Validation passed - no issues found")
		return nil
	}

	headers := []string{"LEVEL", "KEY", "MESSAGE"}
	rows := make([][]string, len(result.Issues))

	for i, issue := range result.Issues {
		level := issue.Level
		if level == "error" {
			level = "✗ ERROR"
		} else {
			level = "⚠ WARNING"
		}
		rows[i] = []string{level, issue.Key, issue.Message}
	}

	table := RenderTable(TableConfig{
		Headers: headers,
		Rows:    rows,
	})

	fmt.Println(table)
	fmt.Println()

	// Print verdict
	if result.Valid {
		Success("✓ Validation passed")
	} else {
		if strict && hasOnlyWarnings(result) {
			Error("✗ Validation failed (strict mode: warnings treated as errors)")
		} else {
			Error("✗ Validation failed")
		}
	}

	return nil
}

// hasOnlyWarnings checks if all issues are warnings
func hasOnlyWarnings(result *validator.ValidationResult) bool {
	for _, issue := range result.Issues {
		if issue.Level == "error" {
			return false
		}
	}
	return true
}

// PrintDryRun writes a summary of the configuration items that would be applied without making any changes.
// 
// Each item is displayed with a progress indicator, an action label ("PUT"), the item's key, and its formatted value.
// After listing all items a final message indicates that no changes were made to etcd.
func PrintDryRun(pairs []*models.ConfigPair) {
	title := StyleIfTerminal(warningStyle, fmt.Sprintf("DRY RUN - Would apply %d configuration items", len(pairs)))
	fmt.Println(StyleIfTerminal(warningPanelStyle, title))
	fmt.Println()

	for i, pair := range pairs {
		value := formatValue(pair.Value)
		key := StyleIfTerminal(keyStyle, pair.Key)

		progress := fmt.Sprintf("[%d/%d]", i+1, len(pairs))
		fmt.Printf("%s %s → %s\n", StyleIfTerminal(valueStyle, progress), StyleIfTerminal(successStyle, "PUT"), key)

		fmt.Printf("%s\n\n", StyleIfTerminal(valueStyle, value))
	}

	fmt.Println(StyleIfTerminal(warningStyle, "DRY RUN complete - no changes made to etcd"))
}

// PrintApplyProgress logs the apply operation progress as a "[current/total]" prefix followed by the key being applied.
// The key is styled when running in a terminal.
func PrintApplyProgress(current, total int, key string) {
	progress := fmt.Sprintf("[%d/%d]", current, total)
	k := StyleIfTerminal(keyStyle, key)
	logger.Log.Info(fmt.Sprintf("%s Applying %s", progress, k))
}

// PrintApplySuccess prints a styled success message indicating how many items were applied to etcd.
func PrintApplySuccess(count int) {
	msg := StyleIfTerminal(successStyle, fmt.Sprintf("✓ Successfully applied %d items to etcd", count))
	fmt.Println()
	fmt.Println(StyleIfTerminal(successPanelStyle, msg))
}

// PrintApplyResultsWithFormat prints the results of an apply operation using the given output format.
// 
// For the "simple" format it prints a dry-run listing when dryRun is true or a success summary otherwise.
// For "json" and "table" formats it delegates to the respective JSON/table renderers.
// Returns an error when the chosen format is unsupported or when a delegated renderer reports an error.
func PrintApplyResultsWithFormat(pairs []*models.ConfigPair, format string, dryRun bool) error {
	switch format {
	case FormatSimple.String():
		if dryRun {
			PrintDryRun(pairs)
			return nil
		}
		PrintApplySuccess(len(pairs))
		return nil
	case FormatJSON.String():
		return printApplyJSON(pairs, dryRun)
	case FormatTable.String():
		return printApplyTable(pairs, dryRun)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func printApplyJSON(pairs []*models.ConfigPair, dryRun bool) error {
	result := map[string]any{
		"applied": len(pairs),
		"dry_run": dryRun,
		"items":   make([]map[string]string, len(pairs)),
	}

	for i, pair := range pairs {
		result["items"].([]map[string]string)[i] = map[string]string{
			"key":   pair.Key,
			"value": formatValue(pair.Value),
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func printApplyTable(pairs []*models.ConfigPair, dryRun bool) error {
	if len(pairs) == 0 {
		Info("No items to apply")
		return nil
	}

	// Show dry-run or applied header
	action := "APPLIED"
	if dryRun {
		action = "DRY-RUN"
	}

	headers := []string{"#", "KEY", "VALUE"}
	rows := make([][]string, len(pairs))

	for i, pair := range pairs {
		value := formatValue(pair.Value)
		if len(value) > 50 {
			value = value[:47] + "..."
		}
		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			pair.Key,
			value,
		}
	}

	table := RenderTable(TableConfig{
		Headers: headers,
		Rows:    rows,
	})

	fmt.Printf("\n%s - %d items:\n", action, len(pairs))
	fmt.Println(table)

	return nil
}

// PrintContextsWithFormat prints contexts using the given format ("simple", "json", "table").
// It returns an error if the provided format is unsupported.
func PrintContextsWithFormat(contexts map[string]*config.ContextConfig, currentContext string, format string) error {
	switch format {
	case FormatSimple.String():
		return printContextsSimple(contexts, currentContext)
	case FormatJSON.String():
		return printContextsJSON(contexts, currentContext)
	case FormatTable.String():
		return printContextsTable(contexts, currentContext)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func printContextsSimple(contexts map[string]*config.ContextConfig, currentContext string) error {
	if len(contexts) == 0 {
		Info("No contexts found")
		return nil
	}

	var contextNames []string
	for name := range contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)

	for _, name := range contextNames {
		if name == currentContext {
			fmt.Printf("* %s\n", name)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}

func printContextsJSON(contexts map[string]*config.ContextConfig, currentContext string) error {
	type contextOutput struct {
		Name      string   `json:"name"`
		Username  string   `json:"username,omitempty"`
		Endpoints []string `json:"endpoints"`
		Current   bool     `json:"current"`
	}

	output := make([]contextOutput, 0, len(contexts))
	for name, ctx := range contexts {
		output = append(output, contextOutput{
			Name:      name,
			Current:   name == currentContext,
			Endpoints: ctx.Endpoints,
			Username:  ctx.Username,
		})
	}

	// Sort by name for consistent output
	sort.Slice(output, func(i, j int) bool {
		return output[i].Name < output[j].Name
	})

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func printContextsTable(contexts map[string]*config.ContextConfig, currentContext string) error {
	if len(contexts) == 0 {
		Info("No contexts found")
		return nil
	}

	var contextNames []string
	for name := range contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)

	headers := []string{"CURRENT", "NAME", "ENDPOINTS", "USER"}
	rows := make([][]string, len(contextNames))

	for i, name := range contextNames {
		ctx := contexts[name]
		current := ""
		if name == currentContext {
			current = "*"
		}

		endpoints := ""
		if len(ctx.Endpoints) > 0 {
			endpoints = ctx.Endpoints[0]
			if len(ctx.Endpoints) > 1 {
				endpoints += fmt.Sprintf(" (+%d)", len(ctx.Endpoints)-1)
			}
		}

		rows[i] = []string{current, name, endpoints, ctx.Username}
	}

	table := RenderTable(TableConfig{
		Headers: headers,
		Rows:    rows,
	})

	fmt.Println(table)
	return nil
}

// Supported formats are "simple", "json", and "table". It returns an error if the format is unsupported.
func PrintConfigViewWithFormat(cfg *config.Config, format string) error {
	switch format {
	case FormatSimple.String():
		return printConfigViewSimple(cfg)
	case FormatJSON.String():
		return printConfigViewJSON(cfg)
	case FormatTable.String():
		return printConfigViewTable(cfg)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func printConfigViewSimple(cfg *config.Config) error {
	fmt.Printf("current-context: %s\n", cfg.CurrentContext)
	fmt.Printf("log-level: %s\n", cfg.LogLevel)
	fmt.Printf("default-format: %s\n", cfg.DefaultFormat)
	fmt.Printf("strict: %v\n", cfg.Strict)
	fmt.Printf("no-validate: %v\n", cfg.NoValidate)
	fmt.Printf("contexts: %d\n", len(cfg.Contexts))
	return nil
}

func printConfigViewJSON(cfg *config.Config) error {
	// Create a sanitized version without passwords
	type sanitizedContext struct {
		Username  string   `json:"username,omitempty"`
		Endpoints []string `json:"endpoints"`
	}

	output := map[string]any{
		"current_context": cfg.CurrentContext,
		"log_level":       cfg.LogLevel,
		"default_format":  cfg.DefaultFormat,
		"strict":          cfg.Strict,
		"no_validate":     cfg.NoValidate,
		"contexts":        make(map[string]sanitizedContext),
	}

	for name, ctx := range cfg.Contexts {
		output["contexts"].(map[string]sanitizedContext)[name] = sanitizedContext{
			Endpoints: ctx.Endpoints,
			Username:  ctx.Username,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func printConfigViewTable(cfg *config.Config) error {
	// Settings table
	fmt.Println("Settings:")
	settingsHeaders := []string{"SETTING", "VALUE"}
	settingsRows := [][]string{
		{"current-context", cfg.CurrentContext},
		{"log-level", cfg.LogLevel},
		{"default-format", cfg.DefaultFormat},
		{"strict", fmt.Sprintf("%v", cfg.Strict)},
		{"no-validate", fmt.Sprintf("%v", cfg.NoValidate)},
	}

	settingsTable := RenderTable(TableConfig{
		Headers: settingsHeaders,
		Rows:    settingsRows,
	})
	fmt.Println(settingsTable)
	fmt.Println()

	// Contexts table
	if len(cfg.Contexts) > 0 {
		fmt.Printf("Contexts (%d):\n", len(cfg.Contexts))
		return printContextsTable(cfg.Contexts, cfg.CurrentContext)
	}

	Info("No contexts configured")
	return nil
}

// PrintError prints the provided error inside a styled error panel prefixed with a cross.
// Applies terminal styling when supported.
func PrintError(err error) {
	msg := StyleIfTerminal(errorStyle, fmt.Sprintf("✗ Error: %v", err))
	fmt.Println(StyleIfTerminal(errorPanelStyle, msg))
}

// formatValue formats val for display using models.FormatValue.
// It exists as a package-local alias retained for backward compatibility.
func formatValue(val any) string {
	return models.FormatValue(val)
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

// Info prints an informational message prefixed with an ellipsis and styled for terminal output when supported.
func Info(msg string) {
	fmt.Println(StyleIfTerminal(valueStyle, "⋯ "+msg))
}

// Success prints a success message prefixed with a checkmark, applying terminal styling when appropriate.
func Success(msg string) {
	fmt.Println(StyleIfTerminal(successStyle, "✓ "+msg))
}

// Error prints an error message
func Error(msg string) {
	fmt.Println(StyleIfTerminal(errorStyle, "✗ "+msg))
}

// Warning prints a warning message prefixed with a warning symbol.
// When output is a terminal, the message is rendered with warning styling; otherwise it is printed as plain text.
func Warning(msg string) {
	fmt.Println(StyleIfTerminal(warningStyle, "⚠ "+msg))
}

// Prompt writes a prompt message to standard output prefixed by a styled "? " indicator.
// The message is written without a trailing newline; styling is applied only when output is a terminal.
func Prompt(msg string) {
	fmt.Print(StyleIfTerminal(keyStyle, "? ") + msg)
}

// - restrict config file permissions (e.g., 0600).
func PrintSecurityWarning() {
	fmt.Println()
	Warning("Security Warning:")
	fmt.Println(StyleIfTerminal(valueStyle, "  Your password is stored in plain text in the config file."))
	fmt.Println(StyleIfTerminal(valueStyle, "  For better security:"))
	fmt.Println(StyleIfTerminal(valueStyle, "    - Use --password flag at runtime instead of storing it"))
	fmt.Println(StyleIfTerminal(valueStyle, "    - Use --password-stdin for CI/CD pipelines"))
	fmt.Println(StyleIfTerminal(valueStyle, "    - Ensure config file permissions are restrictive (0600)"))
}

// PrintTree renders etcd configuration as a tree structure
func PrintTree(pairs []*models.ConfigPair) error {
	logger.Log.Info(fmt.Sprintf("Found %d configuration items", len(pairs)))
	fmt.Println()

	t := buildEtcdTree(pairs)
	fmt.Println(t)
	return nil
}

// buildEtcdTree builds a hierarchical tree representing the provided etcd-style key/value pairs for display.
// It arranges pairs into folder nodes for intermediate path segments and leaf nodes that show keys with formatted values.
// Styling and enumerator appearance are applied when output is a terminal.
// It returns the root tree ready for rendering.
func buildEtcdTree(pairs []*models.ConfigPair) *tree.Tree {
	root := tree.Root("/").
		Enumerator(tree.RoundedEnumerator)

	if IsTerminal() {
		root = root.RootStyle(treeRootStyle).EnumeratorStyle(treeEnumeratorStyle)
	}

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

			if _, exists := pathMap[currentPath]; !exists {
				parent := pathMap[parentPath]

				if i == len(parts)-1 {
					valueStr := formatTreeValue(pair.Value)
					display := StyleIfTerminal(treeKeyStyle, part) + " " + StyleIfTerminal(treeValueStyle, valueStr)
					leaf := tree.New().Root(display)
					parent.Child(leaf)
				} else {
					folder := tree.New().
						Root(StyleIfTerminal(treeFolderStyle, part+"/"))
					if IsTerminal() {
						folder = folder.EnumeratorStyle(treeEnumeratorStyle)
					}
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