package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"
)

var (
	convertOpts struct {
		FilePath string
		Format   string
	}

	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Convert configuration files to YAML",
		Long:  `Convert configuration files from various formats (etcdctl, JSON, YAML) to hierarchical YAML.`,
		Example: `  # Convert etcdctl dump to YAML
  etu convert -f dump.txt > config.yaml

  # Convert from stdin
  cat dump.txt | etu convert -f -
  
  # Convert with explicit format
  etu convert -f data.json --format json`,
		RunE: runConvert,
	}
)

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&convertOpts.FilePath, "file", "f", "",
		"path to configuration file (supports stdin via '-')")
	convertCmd.Flags().StringVar(&convertOpts.Format, "format", "",
		"input format: auto, etcdctl, json, yaml")
}

func runConvert(_ *cobra.Command, _ []string) error {
	ctx, cancel := getOperationContext()
	defer cancel()

	appCfg := loadAppConfig()

	filePath := convertOpts.FilePath
	if filePath == "-" || (filePath == "" && hasStdinData()) {
		tmpFile, err := os.CreateTemp("", "etu-convert-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
			tmpFile.Close()
			return fmt.Errorf("failed to write stdin to temp file: %w", err)
		}
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close temp file: %w", err)
		}
		filePath = tmpFile.Name()
	}

	if filePath == "" {
		return fmt.Errorf("input file required: use -f <file> or pipe data to stdin")
	}

	pairs, err := parseConfigFile(ctx, filePath, models.FormatType(convertOpts.Format), appCfg)
	if err != nil {
		return err
	}

	data, err := parsers.UnflattenMap(pairs)
	if err != nil {
		return err
	}

	bytes, err := output.SerializeYAML(data)
	if err != nil {
		return fmt.Errorf("failed to serialize to YAML: %w", err)
	}

	fmt.Print(string(bytes))
	return nil
}

func hasStdinData() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
