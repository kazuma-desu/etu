package cmd

import (
	"fmt"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	parseOpts models.ParseOptions
	treeView  bool

	parseCmd = &cobra.Command{
		Use:   "parse -f FILE",
		Short: "Parse and display configuration",
		Long: `Parse configuration from a file and display it without validation or applying to etcd.

The parse command is useful for:
  - Inspecting configuration file contents
  - Converting between formats (with --json flag)
  - Visualizing hierarchical structure (with --tree flag)
  - Debugging parser behavior
  - Integrating with other tools via JSON output`,
		Example: `  # Parse and display configuration
  etu parse -f config.txt

  # Display as a tree view
  etu parse -f config.txt --tree

  # Output as JSON for scripting
  etu parse -f config.txt --json

  # Pipe to jq for filtering
  etu parse -f config.txt --json | jq '.[] | select(.key | contains("database"))'`,
		RunE: runParse,
	}
)

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVarP(&parseOpts.FilePath, "file", "f", "",
		"path to configuration file (required)")
	parseCmd.Flags().StringVar((*string)(&parseOpts.Format), "format", "",
		"file format: auto, etcdctl (overrides config)")
	parseCmd.Flags().BoolVar(&parseOpts.JSONOutput, "json", false,
		"output as JSON")
	parseCmd.Flags().BoolVar(&treeView, "tree", false,
		"display as tree view")

	parseCmd.MarkFlagRequired("file")
	parseCmd.MarkFlagsMutuallyExclusive("json", "tree")
}

func runParse(cmd *cobra.Command, args []string) error {
	// Load config for defaults
	appCfg, _ := config.LoadConfig()

	// Apply config defaults if flags not set
	// Priority: flag > config > default
	format := parseOpts.Format
	if format == "" && appCfg != nil && appCfg.DefaultFormat != "" {
		format = models.FormatType(appCfg.DefaultFormat)
	}
	if format == "" {
		format = models.FormatAuto
	}

	// Parse the file
	registry := parsers.NewRegistry()
	if format == models.FormatAuto {
		var err error
		format, err = registry.DetectFormat(parseOpts.FilePath)
		if err != nil {
			return fmt.Errorf("failed to detect format: %w", err)
		}
		log.Debug("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return err
	}

	if !parseOpts.JSONOutput {
		log.Info("Parsing configuration", "file", parseOpts.FilePath, "format", format)
	}

	pairs, err := parser.Parse(parseOpts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Display based on output format
	if treeView {
		return output.PrintTree(pairs)
	}

	return output.PrintConfigPairs(pairs, parseOpts.JSONOutput)
}
