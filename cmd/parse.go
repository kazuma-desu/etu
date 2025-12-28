package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
)

var (
	parseOpts models.ParseOptions

	parseCmd = &cobra.Command{
		Use:   "parse -f FILE",
		Short: "Parse and display configuration",
		Long: `Parse configuration from a file and display it without validation or applying to etcd.

The parse command is useful for:
  - Inspecting configuration file contents
  - Converting between formats (with -o json)
  - Visualizing hierarchical structure (with -o tree)
  - Debugging parser behavior
  - Integrating with other tools via JSON output`,
		Example: `  # Parse and display configuration
  etu parse -f config.txt

  # Display as a tree view
  etu parse -f config.txt -o tree

  # Display as a table
  etu parse -f config.txt -o table

  # Output as JSON for scripting
  etu parse -f config.txt -o json

  # Pipe to jq for filtering
  etu parse -f config.txt -o json | jq '.[] | select(.key | contains("database"))'`,
		RunE: runParse,
	}
)

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVarP(&parseOpts.FilePath, "file", "f", "",
		"path to configuration file (required)")
	parseCmd.Flags().StringVar((*string)(&parseOpts.Format), "format", "",
		"file format: auto, etcdctl (overrides config)")

	if err := parseCmd.MarkFlagRequired("file"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}

func runParse(_ *cobra.Command, _ []string) error {
	appCfg := loadAppConfig()

	pairs, err := parseConfigFile(parseOpts.FilePath, parseOpts.Format, appCfg)
	if err != nil {
		return err
	}

	normalizedFormat, err := normalizeOutputFormat(formatsWithTree)
	if err != nil {
		return err
	}

	return output.PrintConfigPairsWithFormat(pairs, normalizedFormat)
}
