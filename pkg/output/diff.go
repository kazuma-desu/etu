package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
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
	case FormatTable.String():
		return printDiffTable(result, showUnchanged)
	default:
		return fmt.Errorf("unsupported format: %s (use simple, json, or table)", format)
	}
}

// printDiffSimple prints diff in simple human-readable format
func printDiffSimple(result *DiffResult, showUnchanged bool) error {
	Info(fmt.Sprintf("Diff result: +%d ~%d -%d", result.Added, result.Modified, result.Deleted))
	fmt.Println()

	// Group by status
	var added, modified, deleted, unchanged []*DiffEntry
	for _, e := range result.Entries {
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

	if len(added) == 0 && len(modified) == 0 && len(deleted) == 0 {
		Success("No changes detected")
		// If we only have unchanged items and showUnchanged is false, we might want to return here.
		// But if showUnchanged is true, we proceed to print them.
		if !showUnchanged {
			return nil
		}
	}

	// Print added
	if len(added) > 0 {
		fmt.Println(addedStyle.Render(fmt.Sprintf("Added (%d):", len(added))))
		for _, e := range added {
			fmt.Printf("  %s %s\n", addedStyle.Render("+"), keyStyle.Render(e.Key))
			fmt.Printf("    %s\n", valueStyle.Render(e.NewValue))
		}
		fmt.Println()
	}

	// Print modified
	if len(modified) > 0 {
		fmt.Println(modifiedStyle.Render(fmt.Sprintf("Modified (%d):", len(modified))))
		for _, e := range modified {
			fmt.Printf("  %s %s\n", modifiedStyle.Render("~"), keyStyle.Render(e.Key))
			fmt.Printf("    %sold: %s\n", oldValueStyle.Render("  "), oldValueStyle.Render(e.OldValue))
			fmt.Printf("    %snew: %s\n", newValueStyle.Render("  "), newValueStyle.Render(e.NewValue))
		}
		fmt.Println()
	}

	// Print deleted
	if len(deleted) > 0 {
		fmt.Println(deletedStyle.Render(fmt.Sprintf("Deleted (%d):", len(deleted))))
		for _, e := range deleted {
			fmt.Printf("  %s %s\n", deletedStyle.Render("-"), keyStyle.Render(e.Key))
			fmt.Printf("    %s\n", oldValueStyle.Render(e.OldValue))
		}
		fmt.Println()
	}

	// Print unchanged if requested
	if len(unchanged) > 0 {
		fmt.Println(unchangedStyle.Render(fmt.Sprintf("Unchanged (%d):", len(unchanged))))
		for _, e := range unchanged {
			fmt.Printf("  %s %s\n", unchangedStyle.Render("="), keyStyle.Render(e.Key))
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("Summary: +%d ~%d -%d", result.Added, result.Modified, result.Deleted)
	if showUnchanged {
		fmt.Printf(" =%d", result.Unchanged)
	}
	fmt.Printf(" = %d total\n", len(result.Entries))

	return nil
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
		var statusStyle lipgloss.Style
		switch e.Status {
		case DiffStatusAdded:
			status = "+"
			statusStyle = addedStyle
		case DiffStatusModified:
			status = "~"
			statusStyle = modifiedStyle
		case DiffStatusDeleted:
			status = "-"
			statusStyle = deletedStyle
		case DiffStatusUnchanged:
			status = "="
			statusStyle = unchangedStyle
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
			statusStyle.Render(status),
			e.Key,
			oldValueStyle.Render(oldVal),
			newValueStyle.Render(newVal),
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
