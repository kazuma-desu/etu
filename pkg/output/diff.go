package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// DiffStatus represents the status of a key in the diff
type DiffStatus string

const (
	DiffStatusAdded     DiffStatus = "added"
	DiffStatusModified  DiffStatus = "modified"
	DiffStatusDeleted   DiffStatus = "deleted"
	DiffStatusUnchanged DiffStatus = "unchanged"
)

// DiffEntry represents a single key difference
type DiffEntry struct {
	Key      string     `json:"key"`
	Status   DiffStatus `json:"status"`
	OldValue string     `json:"old_value,omitempty"`
	NewValue string     `json:"new_value,omitempty"`
}

// DiffResult contains the diff result with summary counts
type DiffResult struct {
	Entries   []*DiffEntry
	Added     int
	Modified  int
	Deleted   int
	Unchanged int
}

// PrintDiffResult prints the diff result in the specified format
func PrintDiffResult(result *DiffResult, format string, showUnchanged bool) error {
	switch format {
	case FormatSimple.String():
		return printDiffSimple(result, showUnchanged)
	case FormatJSON.String():
		return printDiffJSON(result, showUnchanged)
	case FormatYAML.String():
		return printDiffYAML(result, showUnchanged)
	case FormatTable.String():
		return printDiffTable(result, showUnchanged)
	default:
		return fmt.Errorf("unsupported format: %s (use simple, json, yaml, or table)", format)
	}
}

// printDiffSimple prints diff in simple human-readable format
func printDiffSimple(result *DiffResult, showUnchanged bool) error {
	Info(fmt.Sprintf("Diff result: +%d ~%d -%d", result.Added, result.Modified, result.Deleted))
	fmt.Println()

	added, modified, deleted, unchanged := groupEntriesByStatus(result.Entries, showUnchanged)

	if len(added) == 0 && len(modified) == 0 && len(deleted) == 0 {
		Success("No changes detected")
		if !showUnchanged {
			return nil
		}
	}

	printDiffEntries(added, DiffStatusAdded)
	printDiffEntries(modified, DiffStatusModified)
	printDiffEntries(deleted, DiffStatusDeleted)
	printDiffEntries(unchanged, DiffStatusUnchanged)

	printDiffSummary(result, showUnchanged)
	return nil
}

// groupEntriesByStatus groups diff entries by their status.
func groupEntriesByStatus(entries []*DiffEntry, showUnchanged bool) (added, modified, deleted, unchanged []*DiffEntry) {
	for _, e := range entries {
		switch e.Status {
		case DiffStatusAdded:
			added = append(added, e)
		case DiffStatusModified:
			modified = append(modified, e)
		case DiffStatusDeleted:
			deleted = append(deleted, e)
		case DiffStatusUnchanged:
			if showUnchanged {
				unchanged = append(unchanged, e)
			}
		}
	}
	return
}

// printDiffEntries prints a group of diff entries with appropriate styling.
func printDiffEntries(entries []*DiffEntry, status DiffStatus) {
	if len(entries) == 0 {
		return
	}

	titleStyle, prefix, printFunc, title := getDiffPrintConfig(status)
	fmt.Println(StyleIfTerminal(titleStyle, fmt.Sprintf("%s (%d):", title, len(entries))))
	for _, e := range entries {
		printFunc(prefix, e)
	}
	fmt.Println()
}

// getDiffPrintConfig returns the style, prefix, print function, and title for a given status.
func getDiffPrintConfig(status DiffStatus) (lipgloss.Style, string, func(string, *DiffEntry), string) {
	switch status {
	case DiffStatusAdded:
		return addedStyle, "+", printAddedEntry, "Added"
	case DiffStatusModified:
		return modifiedStyle, "~", printModifiedEntry, "Modified"
	case DiffStatusDeleted:
		return deletedStyle, "-", printDeletedEntry, "Deleted"
	default:
		return unchangedStyle, "=", printUnchangedEntry, "Unchanged"
	}
}

func printAddedEntry(prefix string, e *DiffEntry) {
	fmt.Printf("  %s %s\n", StyleIfTerminal(addedStyle, prefix), StyleIfTerminal(keyStyle, e.Key))
	fmt.Printf("    %s\n", StyleIfTerminal(valueStyle, e.NewValue))
}

func printModifiedEntry(prefix string, e *DiffEntry) {
	fmt.Printf("  %s %s\n", StyleIfTerminal(modifiedStyle, prefix), StyleIfTerminal(keyStyle, e.Key))
	fmt.Printf("    %sold: %s\n", StyleIfTerminal(oldValueStyle, "  "), StyleIfTerminal(oldValueStyle, e.OldValue))
	fmt.Printf("    %snew: %s\n", StyleIfTerminal(newValueStyle, "  "), StyleIfTerminal(newValueStyle, e.NewValue))
}

func printDeletedEntry(prefix string, e *DiffEntry) {
	fmt.Printf("  %s %s\n", StyleIfTerminal(deletedStyle, prefix), StyleIfTerminal(keyStyle, e.Key))
	fmt.Printf("    %s\n", StyleIfTerminal(oldValueStyle, e.OldValue))
}

func printUnchangedEntry(prefix string, e *DiffEntry) {
	fmt.Printf("  %s %s\n", StyleIfTerminal(unchangedStyle, prefix), StyleIfTerminal(keyStyle, e.Key))
}

// printDiffSummary prints the diff summary line.
func printDiffSummary(result *DiffResult, showUnchanged bool) {
	fmt.Printf("Summary: +%d ~%d -%d", result.Added, result.Modified, result.Deleted)
	if showUnchanged {
		fmt.Printf(" =%d", result.Unchanged)
	}
	fmt.Printf(" = %d total\n", len(result.Entries))
}

// printDiffJSON prints diff as JSON
func printDiffJSON(result *DiffResult, showUnchanged bool) error {
	type jsonEntry struct {
		Key      string `json:"key"`
		Status   string `json:"status"`
		OldValue string `json:"old_value,omitempty"`
		NewValue string `json:"new_value,omitempty"`
	}

	type jsonOutput struct {
		Entries   []jsonEntry `json:"entries"`
		Added     int         `json:"added"`
		Modified  int         `json:"modified"`
		Deleted   int         `json:"deleted"`
		Unchanged int         `json:"unchanged,omitempty"`
	}

	entries := make([]jsonEntry, 0)
	if showUnchanged {
		for _, e := range result.Entries {
			entries = append(entries, jsonEntry{
				Key:      e.Key,
				Status:   string(e.Status),
				OldValue: e.OldValue,
				NewValue: e.NewValue,
			})
		}
	} else {
		for _, e := range result.Entries {
			if e.Status != DiffStatusUnchanged {
				entries = append(entries, jsonEntry{
					Key:      e.Key,
					Status:   string(e.Status),
					OldValue: e.OldValue,
					NewValue: e.NewValue,
				})
			}
		}
	}

	output := jsonOutput{
		Added:     result.Added,
		Modified:  result.Modified,
		Deleted:   result.Deleted,
		Unchanged: result.Unchanged,
		Entries:   entries,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func printDiffYAML(result *DiffResult, showUnchanged bool) error {
	type yamlEntry struct {
		Key      string `yaml:"key"`
		Status   string `yaml:"status"`
		OldValue string `yaml:"old_value,omitempty"`
		NewValue string `yaml:"new_value,omitempty"`
	}

	type yamlOutput struct {
		Entries   []yamlEntry `yaml:"entries"`
		Added     int         `yaml:"added"`
		Modified  int         `yaml:"modified"`
		Deleted   int         `yaml:"deleted"`
		Unchanged int         `yaml:"unchanged,omitempty"`
	}

	entries := make([]yamlEntry, 0)
	if showUnchanged {
		for _, e := range result.Entries {
			entries = append(entries, yamlEntry{
				Key:      e.Key,
				Status:   string(e.Status),
				OldValue: e.OldValue,
				NewValue: e.NewValue,
			})
		}
	} else {
		for _, e := range result.Entries {
			if e.Status != DiffStatusUnchanged {
				entries = append(entries, yamlEntry{
					Key:      e.Key,
					Status:   string(e.Status),
					OldValue: e.OldValue,
					NewValue: e.NewValue,
				})
			}
		}
	}

	output := yamlOutput{
		Added:     result.Added,
		Modified:  result.Modified,
		Deleted:   result.Deleted,
		Unchanged: result.Unchanged,
		Entries:   entries,
	}

	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal diff to YAML: %w", err)
	}

	_, err = os.Stdout.Write(data)
	return err
}

// printDiffTable prints diff as a table
func printDiffTable(result *DiffResult, showUnchanged bool) error {
	var entries []*DiffEntry
	if showUnchanged {
		entries = result.Entries
	} else {
		for _, e := range result.Entries {
			if e.Status != DiffStatusUnchanged {
				entries = append(entries, e)
			}
		}
	}

	if len(entries) == 0 {
		Success("No changes detected")
		return nil
	}

	headers := []string{"STATUS", "KEY", "OLD VALUE", "NEW VALUE"}
	rows := make([][]string, len(entries))

	for i, e := range entries {
		var status string
		var sStyle lipgloss.Style
		switch e.Status {
		case DiffStatusAdded:
			status = "+"
			sStyle = addedStyle
		case DiffStatusModified:
			status = "~"
			sStyle = modifiedStyle
		case DiffStatusDeleted:
			status = "-"
			sStyle = deletedStyle
		case DiffStatusUnchanged:
			status = "="
			sStyle = unchangedStyle
		}

		oldVal := e.OldValue
		if len(oldVal) > 40 {
			oldVal = oldVal[:37] + "..."
		}
		newVal := e.NewValue
		if len(newVal) > 40 {
			newVal = newVal[:37] + "..."
		}

		rows[i] = []string{
			StyleIfTerminal(sStyle, status),
			e.Key,
			StyleIfTerminal(oldValueStyle, oldVal),
			StyleIfTerminal(newValueStyle, newVal),
		}
	}

	table := RenderTable(TableConfig{
		Headers: headers,
		Rows:    rows,
	})

	fmt.Println(table)
	fmt.Println()
	fmt.Printf("Summary: +%d ~%d -%d", result.Added, result.Modified, result.Deleted)
	if showUnchanged {
		fmt.Printf(" =%d", result.Unchanged)
	}
	fmt.Printf(" = %d total\n", len(result.Entries))

	return nil
}

// DiffKeyValues computes the diff between two key-value maps
func DiffKeyValues(fileMap, etcdMap map[string]string) *DiffResult {
	result := &DiffResult{
		Entries: make([]*DiffEntry, 0),
	}

	// Collect all keys
	allKeys := make([]string, 0, len(fileMap)+len(etcdMap))
	for k := range fileMap {
		allKeys = append(allKeys, k)
	}
	for k := range etcdMap {
		if _, exists := fileMap[k]; !exists {
			allKeys = append(allKeys, k)
		}
	}
	sort.Strings(allKeys)

	// Compute diff for each key
	for _, key := range allKeys {
		fileVal, fileExists := fileMap[key]
		etcdVal, etcdExists := etcdMap[key]

		entry := &DiffEntry{Key: key}

		// Determine status based on presence in each map
		switch {
		case !fileExists && etcdExists:
			// Key only in etcd -> deleted
			entry.Status = DiffStatusDeleted
			entry.OldValue = etcdVal
		case fileExists && !etcdExists:
			// Key only in file -> added
			entry.Status = DiffStatusAdded
			entry.NewValue = fileVal
		case fileVal != etcdVal:
			// Key in both, values differ -> modified
			entry.Status = DiffStatusModified
			entry.OldValue = etcdVal
			entry.NewValue = fileVal
		default:
			// Key in both, values same -> unchanged
			entry.Status = DiffStatusUnchanged
			entry.OldValue = fileVal
			entry.NewValue = fileVal
		}

		result.Entries = append(result.Entries, entry)

		switch entry.Status {
		case DiffStatusAdded:
			result.Added++
		case DiffStatusModified:
			result.Modified++
		case DiffStatusDeleted:
			result.Deleted++
		case DiffStatusUnchanged:
			result.Unchanged++
		}
	}

	return result
}
