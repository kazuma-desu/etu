package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/validator"
)

var (
	applyOpts models.ApplyOptions

	applyCmd = &cobra.Command{
		Use:   "apply -f <file>",
		Short: "Apply configuration to etcd",
		Long:  `Parse, validate, and apply configuration from a file to etcd.`,
		Example: `  # Apply configuration
  etu apply -f config.txt

  # Preview changes without applying
  etu apply -f config.txt --dry-run

  # JSON output for CI/CD
  etu apply -f config.txt -o json

  # Strict validation (warnings as errors)
  etu apply -f config.txt --strict`,
		RunE: runApply,
	}
)

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().StringVarP(&applyOpts.FilePath, "file", "f", "",
		"path to configuration file (required)")
	applyCmd.Flags().StringVar((*string)(&applyOpts.Format), "format", "",
		"file format: auto, etcdctl (overrides config)")
	applyCmd.Flags().BoolVar(&applyOpts.DryRun, "dry-run", false,
		"preview changes without applying to etcd")
	applyCmd.Flags().BoolVar(&applyOpts.NoValidate, "no-validate", false,
		"skip validation (overrides config, not recommended)")
	applyCmd.Flags().BoolVar(&applyOpts.Strict, "strict", false,
		"treat validation warnings as errors (overrides config)")

	if err := applyCmd.MarkFlagRequired("file"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}

func runApply(cmd *cobra.Command, _ []string) error {
	ctx, cancel := getOperationContext()
	defer cancel()

	appCfg := loadAppConfig()
	noValidate := resolveNoValidateOption(applyOpts.NoValidate, cmd.Flags().Changed("no-validate"), appCfg)
	strict := resolveStrictOption(applyOpts.Strict, cmd.Flags().Changed("strict"), appCfg)

	pairs, err := parseConfigFile(applyOpts.FilePath, applyOpts.Format, appCfg)
	if err != nil {
		return err
	}
	logVerboseInfo(fmt.Sprintf("Parsed %d configuration items", len(pairs)))

	if !noValidate {
		logVerboseInfo("Validating configuration")

		v := validator.NewValidator(strict)
		result := v.Validate(pairs)

		if !isQuietOutput() {
			output.PrintValidationResult(result, strict)
		}

		if !result.Valid {
			if !isQuietOutput() {
				logger.Log.Error("Validation failed - not applying to etcd")
			}
			return fmt.Errorf("validation failed")
		}

		if !isQuietOutput() {
			logger.Log.Info("Validation passed")
			fmt.Println()
		}
	}

	normalizedFormat, err := normalizeOutputFormat(formatsWithoutTree)
	if err != nil {
		return err
	}

	etcdClient, cleanup, err := newEtcdClientOrDryRun(applyOpts.DryRun)
	if err != nil {
		return err
	}
	defer cleanup()

	logVerboseInfo(fmt.Sprintf("Applying %d items to etcd", len(pairs)))

	var onProgress client.ProgressFunc
	if normalizedFormat == "simple" && !applyOpts.DryRun {
		onProgress = func(current, total int, key string) {
			output.PrintApplyProgress(current, total, key)
		}
	}

	result, err := etcdClient.PutAllWithProgress(ctx, pairs, onProgress)
	if err != nil {
		if result != nil && result.Succeeded > 0 {
			output.Warning(fmt.Sprintf("Partial failure: %d/%d items applied before error",
				result.Succeeded, result.Total))
		}
		return wrapTimeoutError(fmt.Errorf("failed to apply configuration: %w", err))
	}

	if recorder, ok := etcdClient.(client.OperationRecorder); ok {
		return output.PrintDryRunOperations(recorder.Operations(), normalizedFormat)
	}

	return output.PrintApplyResultsWithFormat(pairs, normalizedFormat, applyOpts.DryRun)
}
