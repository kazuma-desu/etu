package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/validator"
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

  # JSON output for CI/CD pipelines
  etu validate -f config.txt -o json

  # Table format for summary view
  etu validate -f config.txt -o table`,
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
	appCfg := loadAppConfig()

	pairs, err := parseConfigFile(validateOpts.FilePath, validateOpts.Format, appCfg)
	if err != nil {
		return err
	}

	strict := resolveStrictOption(validateOpts.Strict, cmd.Flags().Changed("strict"), appCfg)

	if !isQuietOutput() {
		logVerboseInfo(fmt.Sprintf("Parsed %d configuration items", len(pairs)))
		fmt.Println()
		logVerboseInfo("Validating configuration")
	}

	v := validator.NewValidator(strict)
	result := v.Validate(pairs)

	normalizedFormat, err := normalizeOutputFormat(formatsWithoutTree)
	if err != nil {
		return err
	}

	if err := output.PrintValidationWithFormat(result, strict, normalizedFormat); err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}
