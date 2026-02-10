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
		Use:   "parse -f <file>",
		Short: "Parse and display configuration",
		Long:  `Parse a configuration file and display its contents without applying to etcd.`,
		Example: `  # Parse and display
  etu parse -f config.txt

  # Tree view
  etu parse -f config.txt -o tree

  # JSON output for scripting
  etu parse -f config.txt -o json`,
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

	registerFileCompletion(parseCmd, "file")
}

func runParse(_ *cobra.Command, _ []string) error {
	allowedFormats := []string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
		output.FormatYAML.String(),
		output.FormatTable.String(),
		output.FormatTree.String(),
	}
	if err := validateOutputFormat(allowedFormats); err != nil {
		return err
	}

	ctx, cancel := getOperationContext()
	defer cancel()

	appCfg := loadAppConfig()

	pairs, err := parseConfigFile(ctx, parseOpts.FilePath, parseOpts.Format, appCfg)
	if err != nil {
		return err
	}

	return output.PrintConfigPairsWithFormat(pairs, outputFormat)
}
