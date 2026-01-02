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
	case "json":
		return printDryRunJSON(ops)
	case "simple", "table":
		return printDryRunSimple(ops)
	default:
		return fmt.Errorf("unsupported format for dry-run: %s", format)
	}
}

func printDryRunSimple(ops []client.Operation) error {
	title := warningStyle.Render(fmt.Sprintf("DRY RUN - Would perform %d operations", len(ops)))
	fmt.Println(warningPanelStyle.Render(title))
	fmt.Println()

	for i, op := range ops {
		progress := fmt.Sprintf("[%d/%d]", i+1, len(ops))

		var actionStyle string
		if op.Type == "PUT" {
			actionStyle = successStyle.Render("PUT")
		} else {
			actionStyle = keyStyle.Render(op.Type)
		}

		key := keyStyle.Render(op.Key)
		fmt.Printf("%s %s → %s\n", valueStyle.Render(progress), actionStyle, key)

		if op.Value != "" {
			fmt.Printf("%s\n\n", valueStyle.Render(op.Value))
		}
	}

	fmt.Println(warningStyle.Render("DRY RUN complete - no changes made to etcd"))
	return nil
}

func printDryRunJSON(ops []client.Operation) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ops)
}
