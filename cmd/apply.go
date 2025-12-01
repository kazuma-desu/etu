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

	applyCmd.MarkFlagRequired("file")
}

func runApply(cmd *cobra.Command, args []string) error {
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
		log.Debug("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return err
	}

	log.Info("Parsing configuration", "file", applyOpts.FilePath, "format", format)
	pairs, err := parser.Parse(applyOpts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	log.Info(fmt.Sprintf("Parsed %d configuration items", len(pairs)))

	// Validate unless --no-validate is set
	if !noValidate {
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
	}

	// Dry run or apply
	if applyOpts.DryRun {
		output.PrintDryRun(pairs)
		return nil
	}

	// Apply to etcd
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
