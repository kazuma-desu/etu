package cmd

import (
	"fmt"
	"os"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"
	"github.com/kazuma-desu/etu/pkg/validator"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	validateOpts models.ValidateOptions

	validateCmd = &cobra.Command{
		Use:   "validate -f FILE",
		Short: "Validate configuration without applying",
		Long: `Parse and validate configuration from a file without applying it to etcd.

The validate command performs comprehensive checks on your configuration including:
  - Key format validation (must start with /, valid characters, length limits)
  - Value validation (non-null, size limits, structured data validity)
  - URL validation for keys containing "url"
  - Duplicate key detection

This is useful for CI/CD pipelines or pre-deployment checks.`,
		Example: `  # Validate a configuration file
  etu validate -f config.txt

  # Strict mode (treat warnings as errors)
  etu validate -f config.txt --strict

  # Specify format explicitly
  etu validate -f config.txt --format etcdctl`,
		RunE: runValidate,
	}
)

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&validateOpts.FilePath, "file", "f", "",
		"path to configuration file (required)")
	validateCmd.Flags().StringVar((*string)(&validateOpts.Format), "format", "",
		"file format: auto, etcdctl (overrides config)")
	validateCmd.Flags().BoolVar(&validateOpts.Strict, "strict", false,
		"treat validation warnings as errors (overrides config)")

	if err := validateCmd.MarkFlagRequired("file"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}

func runValidate(cmd *cobra.Command, _ []string) error {
	// Load config for defaults
	appCfg, err := config.LoadConfig()
	if err != nil {
		log.Warn("Failed to load config, using defaults", "error", err)
	}

	// Apply config defaults if flags not set
	// Priority: flag > config > default
	format := validateOpts.Format
	if format == "" && appCfg != nil && appCfg.DefaultFormat != "" {
		format = models.FormatType(appCfg.DefaultFormat)
	}
	if format == "" {
		format = models.FormatAuto
	}

	strict := validateOpts.Strict
	if !cmd.Flags().Changed("strict") && appCfg != nil {
		strict = appCfg.Strict
	}

	// Parse the file
	registry := parsers.NewRegistry()
	if format == models.FormatAuto {
		var detectErr error
		format, detectErr = registry.DetectFormat(validateOpts.FilePath)
		if detectErr != nil {
			return fmt.Errorf("failed to detect format: %w", detectErr)
		}
		log.Debug("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return err
	}

	log.Info("Parsing configuration", "file", validateOpts.FilePath, "format", format)
	pairs, err := parser.Parse(validateOpts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	log.Info(fmt.Sprintf("Parsed %d configuration items", len(pairs)))
	fmt.Println()

	// Validate
	log.Info("Validating configuration")
	v := validator.NewValidator(strict)
	result := v.Validate(pairs)

	output.PrintValidationResult(result, strict)

	if !result.Valid {
		os.Exit(1)
	}

	return nil
}
