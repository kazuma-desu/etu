package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/validator"
)

var (
	applyOpts models.ApplyOptions

	applyCmd = &cobra.Command{
		Use:   "apply -f FILE",
		Short: "Apply configuration to etcd",
		Long: `Parse configuration from a file, validate it, and apply it to etcd.

The apply command reads configuration from a file, performs validation checks,
and writes the configuration to etcd. Similar to 'kubectl apply', this command
ensures your configuration is validated before being applied.`,
		Example: `  # Apply configuration from a file
  etu apply -f config.txt

  # Preview changes without applying (dry run)
  etu apply -f config.txt --dry-run

  # JSON output for CI/CD pipelines
  etu apply -f config.txt -o json

  # Table format showing applied keys
  etu apply -f config.txt -o table

  # Apply with strict validation (warnings treated as errors)
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

	if applyOpts.DryRun {
		return output.PrintApplyResultsWithFormat(pairs, normalizedFormat, true)
	}

	logVerboseInfo("Connecting to etcd")

	etcdClient, cleanup, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer cleanup()

	logVerboseInfo(fmt.Sprintf("Applying %d items to etcd", len(pairs)))

	for i, pair := range pairs {
		if normalizedFormat == "simple" {
			output.PrintApplyProgress(i+1, len(pairs), pair.Key)
		}
		if err := etcdClient.PutAll(ctx, []*models.ConfigPair{pair}); err != nil {
			return wrapTimeoutError(fmt.Errorf("failed to apply configuration: %w", err))
		}
	}

	return output.PrintApplyResultsWithFormat(pairs, normalizedFormat, false)
}
