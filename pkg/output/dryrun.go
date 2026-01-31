package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kazuma-desu/etu/pkg/client"
)

// PrintDryRunOperations displays recorded operations from dry-run mode.
func PrintDryRunOperations(ops []client.Operation, format string) error {
	switch format {
	case FormatJSON.String():
		return printDryRunJSON(ops)
	case FormatSimple.String(), FormatTable.String():
		return printDryRunSimple(ops)
	default:
		return fmt.Errorf("unsupported format for dry-run: %s", format)
	}
}

func printDryRunSimple(ops []client.Operation) error {
	title := StyleIfTerminal(warningStyle, fmt.Sprintf("DRY RUN - Would perform %d operations", len(ops)))
	fmt.Println(StyleIfTerminal(warningPanelStyle, title))
	fmt.Println()

	for i, op := range ops {
		progress := fmt.Sprintf("[%d/%d]", i+1, len(ops))

		var actionStr string
		if op.Type == "PUT" {
			actionStr = StyleIfTerminal(successStyle, "PUT")
		} else {
			actionStr = StyleIfTerminal(keyStyle, op.Type)
		}

		key := StyleIfTerminal(keyStyle, op.Key)
		fmt.Printf("%s %s → %s\n", StyleIfTerminal(valueStyle, progress), actionStr, key)

		if op.Value != "" {
			fmt.Printf("%s\n\n", StyleIfTerminal(valueStyle, op.Value))
		}
	}

	fmt.Println(StyleIfTerminal(warningStyle, "DRY RUN complete - no changes made to etcd"))
	return nil
}

func printDryRunJSON(ops []client.Operation) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ops)
}
