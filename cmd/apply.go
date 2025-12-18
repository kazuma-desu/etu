package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"
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
	ctx := context.Background()

	// Load config for defaults
	appCfg, _ := config.LoadConfig()

	// Apply config defaults if flags not set
	// Priority: flag > config > default
	format := applyOpts.Format
	if format == "" && appCfg != nil && appCfg.DefaultFormat != "" {
		format = models.FormatType(appCfg.DefaultFormat)
	}
	if format == "" {
		format = models.FormatAuto
	}

	noValidate := applyOpts.NoValidate
	if !cmd.Flags().Changed("no-validate") && appCfg != nil {
		noValidate = appCfg.NoValidate
	}

	strict := applyOpts.Strict
	if !cmd.Flags().Changed("strict") && appCfg != nil {
		strict = appCfg.Strict
	}

	// Parse the file
	registry := parsers.NewRegistry()
	if format == models.FormatAuto {
		var err error
		format, err = registry.DetectFormat(applyOpts.FilePath)
		if err != nil {
			return fmt.Errorf("failed to detect format: %w", err)
		}
		logger.Log.Debugw("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return err
	}

	// Only show info messages for human-readable formats
	if outputFormat != "json" {
		logger.Log.Infow("Parsing configuration", "file", applyOpts.FilePath, "format", format)
	}

	pairs, err := parser.Parse(applyOpts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if outputFormat != "json" {
		logger.Log.Info(fmt.Sprintf("Parsed %d configuration items", len(pairs)))
	}

	// Validate unless --no-validate is set
	if !noValidate {
		if outputFormat != "json" {
			logger.Log.Info("Validating configuration")
		}

		v := validator.NewValidator(strict)
		result := v.Validate(pairs)

		// Only print validation results for non-JSON formats
		// (JSON format will include validation status in final output)
		if outputFormat != "json" {
			output.PrintValidationResult(result, strict)
		}

		if !result.Valid {
			if outputFormat != "json" {
				logger.Log.Error("Validation failed - not applying to etcd")
			}
			os.Exit(1)
		}

		if outputFormat != "json" {
			logger.Log.Info("Validation passed")
			fmt.Println()
		}
	}

	// Normalize format (tree not supported for apply)
	supportedFormats := []string{"simple", "json", "table"}
	normalizedFormat, err := output.NormalizeFormat(outputFormat, supportedFormats)
	if err != nil {
		return err
	}

	// Dry run - just show what would be applied
	if applyOpts.DryRun {
		return output.PrintApplyResultsWithFormat(pairs, normalizedFormat, true)
	}

	// Apply to etcd
	if outputFormat != "json" {
		logger.Log.Info("Connecting to etcd")
	}

	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return fmt.Errorf("failed to get etcd config: %w", err)
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer etcdClient.Close()

	if outputFormat != "json" {
		logger.Log.Info(fmt.Sprintf("Applying %d items to etcd", len(pairs)))
	}

	// Apply each item
	for i, pair := range pairs {
		// Only show progress for simple format
		if normalizedFormat == "simple" {
			output.PrintApplyProgress(i+1, len(pairs), pair.Key)
		}
		if err := etcdClient.PutAll(ctx, []*models.ConfigPair{pair}); err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
	}

	// Print results in requested format
	return output.PrintApplyResultsWithFormat(pairs, normalizedFormat, false)
}
