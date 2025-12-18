package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"
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
		logger.Log.Debugw("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return err
	}

	// Only show info message for human-readable formats
	if outputFormat != "json" {
		logger.Log.Infow("Parsing configuration", "file", parseOpts.FilePath, "format", format)
	}

	pairs, err := parser.Parse(parseOpts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Normalize format (tree is supported for parse)
	supportedFormats := []string{"simple", "json", "table", "tree"}
	normalizedFormat, err := output.NormalizeFormat(outputFormat, supportedFormats)
	if err != nil {
		return err
	}

	// Display using the new format function
	return output.PrintConfigPairsWithFormat(pairs, normalizedFormat)
}
