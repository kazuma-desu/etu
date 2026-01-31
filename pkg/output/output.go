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

// PrintConfigPairsWithFormat prints configuration pairs in the specified format
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

// PrintValidationWithFormat prints validation results in the specified format
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

// PrintApplyResultsWithFormat prints apply results in the specified format
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

// PrintContextsWithFormat prints contexts in the specified format
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

// PrintConfigViewWithFormat prints config view in the specified format
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

// PrintError prints an error message
func PrintError(err error) {
	msg := errorStyle.Render(fmt.Sprintf("✗ Error: %v", err))
	fmt.Println(errorPanelStyle.Render(msg))
}

// formatValue is a package-local alias to models.FormatValue for backward compatibility.
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
	fmt.Println("    - Use --password-stdin for CI/CD pipelines")
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
