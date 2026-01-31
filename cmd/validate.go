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
		Use:   "validate -f <file>",
		Short: "Validate configuration without applying",
		Long:  `Validate a configuration file without applying to etcd. Useful for CI/CD pipelines.`,
		Example: `  # Validate a configuration file
  etu validate -f config.txt

  # Strict mode (warnings as errors)
  etu validate -f config.txt --strict

  # JSON output for CI/CD
  etu validate -f config.txt -o json`,
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

	registerFileCompletion(validateCmd, "file")
}

func runValidate(cmd *cobra.Command, _ []string) error {
	ctx, cancel := getOperationContext()
	defer cancel()

	appCfg := loadAppConfig()

	pairs, err := parseConfigFile(ctx, validateOpts.FilePath, validateOpts.Format, appCfg)
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
