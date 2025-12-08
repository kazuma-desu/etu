package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"
	"github.com/kazuma-desu/etu/pkg/validator"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
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

  # Apply with strict validation (warnings treated as errors)
  etu apply -f config.txt --strict

  # Skip validation (not recommended)
  etu apply -f config.txt --no-validate`,
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

func resolveApplyOptions(cmd *cobra.Command) (models.ApplyOptions, error) {
	appCfg, err := config.LoadConfig()
	if err != nil {
		log.Warn("Failed to load config, using defaults", "error", err)
	}

	opts := applyOpts

	// Resolve format: flag > config > default
	if opts.Format == "" && appCfg != nil && appCfg.DefaultFormat != "" {
		opts.Format = models.FormatType(appCfg.DefaultFormat)
	}
	if opts.Format == "" {
		opts.Format = models.FormatAuto
	}

	// Resolve no-validate flag
	if !cmd.Flags().Changed("no-validate") && appCfg != nil {
		opts.NoValidate = appCfg.NoValidate
	}

	// Resolve strict flag
	if !cmd.Flags().Changed("strict") && appCfg != nil {
		opts.Strict = appCfg.Strict
	}

	return opts, nil
}

func parseConfigurationFile(format models.FormatType) ([]*models.ConfigPair, error) {
	registry := parsers.NewRegistry()

	if format == models.FormatAuto {
		detectedFormat, err := registry.DetectFormat(applyOpts.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to detect format: %w", err)
		}
		format = detectedFormat
		log.Debug("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return nil, err
	}

	log.Info("Parsing configuration", "file", applyOpts.FilePath, "format", format)
	pairs, err := parser.Parse(applyOpts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	log.Info(fmt.Sprintf("Parsed %d configuration items", len(pairs)))
	return pairs, nil
}

func validateConfiguration(pairs []*models.ConfigPair, strict bool) error {
	log.Info("Validating configuration")
	v := validator.NewValidator(strict)
	result := v.Validate(pairs)

	output.PrintValidationResult(result, strict)

	if !result.Valid {
		log.Error("Validation failed - not applying to etcd")
		os.Exit(1)
	}

	log.Info("Validation passed")
	fmt.Println()
	return nil
}

func applyToEtcd(ctx context.Context, pairs []*models.ConfigPair) error {
	log.Info("Connecting to etcd")
	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return fmt.Errorf("failed to get etcd config: %w", err)
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer etcdClient.Close()

	log.Info(fmt.Sprintf("Applying %d items to etcd", len(pairs)))
	for i, pair := range pairs {
		output.PrintApplyProgress(i+1, len(pairs), pair.Key)
		if err := etcdClient.PutAll(ctx, []*models.ConfigPair{pair}); err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
	}

	output.PrintApplySuccess(len(pairs))
	return nil
}

func runApply(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	opts, err := resolveApplyOptions(cmd)
	if err != nil {
		return err
	}

	pairs, err := parseConfigurationFile(opts.Format)
	if err != nil {
		return err
	}

	if !opts.NoValidate {
		if err := validateConfiguration(pairs, opts.Strict); err != nil {
			return err
		}
	}

	if opts.DryRun {
		output.PrintDryRun(pairs)
		return nil
	}

	return applyToEtcd(ctx, pairs)
}
