package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"
)

const errFailedToLoadConfiguration = "failed to load configuration: %w"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage etu configuration",
	Long:  `Manage contexts and settings.`,
}

var getContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "List all available contexts",
	RunE:  runGetContexts,
}

var currentContextCmd = &cobra.Command{
	Use:   "current-context",
	Short: "Display current active context",
	RunE:  runCurrentContext,
}

var useContextCmd = &cobra.Command{
	Use:               "use-context <context-name>",
	Short:             "Switch to a different context",
	Args:              cobra.ExactArgs(1),
	RunE:              runUseContext,
	ValidArgsFunction: CompleteContextNamesForArg,
}

var deleteContextCmd = &cobra.Command{
	Use:               "delete-context <context-name>",
	Short:             "Delete a context",
	Args:              cobra.ExactArgs(1),
	RunE:              runDeleteContext,
	ValidArgsFunction: CompleteContextNamesForArg,
}

var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	RunE:  runViewConfig,
}

var setConfigCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Keys: log-level, default-format, strict, no-validate`,
	Example: `  etu config set log-level debug
  etu config set strict true`,
	Args: cobra.MinimumNArgs(2),
	RunE: runSetConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(getContextsCmd)
	configCmd.AddCommand(currentContextCmd)
	configCmd.AddCommand(useContextCmd)
	configCmd.AddCommand(deleteContextCmd)
	configCmd.AddCommand(setConfigCmd)
	configCmd.AddCommand(viewConfigCmd)
}

func runGetContexts(_ *cobra.Command, _ []string) error {
	allowedFormats := []string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
		output.FormatTable.String(),
	}
	if err := validateOutputFormat(allowedFormats); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf(errFailedToLoadConfiguration, err)
	}

	if len(cfg.Contexts) == 0 {
		output.Info("No contexts found. Use 'etu login <context-name>' to create one.")
		return nil
	}

	contextViews := make(map[string]*output.ContextView, len(cfg.Contexts))
	for name, ctx := range cfg.Contexts {
		contextViews[name] = &output.ContextView{
			Username:  ctx.Username,
			Endpoints: ctx.Endpoints,
		}
	}

	if err := output.PrintContextsWithFormat(contextViews, cfg.CurrentContext, outputFormat); err != nil {
		return fmt.Errorf("failed to print contexts: %w", err)
	}

	return nil
}

func runCurrentContext(_ *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf(errFailedToLoadConfiguration, err)
	}

	if cfg.CurrentContext == "" {
		output.Info("No current context set. Use 'etu login <context-name>' or 'etu config use-context <context-name>'.")
		return nil
	}

	fmt.Print(cfg.CurrentContext)
	return nil
}

func runUseContext(_ *cobra.Command, args []string) error {
	ctxName := args[0]

	if err := config.UseContext(ctxName); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	output.Success(fmt.Sprintf("Switched to context '%s'", ctxName))
	return nil
}

func runDeleteContext(_ *cobra.Command, args []string) error {
	ctxName := args[0]

	if err := config.DeleteContext(ctxName); err != nil {
		return fmt.Errorf("failed to delete context: %w", err)
	}

	output.Success(fmt.Sprintf("Context '%s' deleted", ctxName))
	return nil
}

func runSetConfig(_ *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf(errFailedToLoadConfiguration, err)
	}

	switch key {
	case "log-level":
		// Validate log level
		validLevels := []string{"debug", "info", "warn", "error"}
		if !slices.Contains(validLevels, value) {
			return fmt.Errorf("invalid log level %s, valid: debug, info, warn, error", value)
		}
		cfg.LogLevel = value
	case "default-format":
		// Validate format
		validFormats := []string{"auto", "etcdctl"}
		if !slices.Contains(validFormats, value) {
			return fmt.Errorf("invalid format %s, valid: auto, etcdctl", value)
		}
		cfg.DefaultFormat = value
	case "strict":
		// Parse boolean
		switch value {
		case "true":
			cfg.Strict = true
		case "false":
			cfg.Strict = false
		default:
			return fmt.Errorf("invalid boolean value %s, valid: true, false", value)
		}
	case "no-validate":
		// Parse boolean
		switch value {
		case "true":
			cfg.NoValidate = true
		case "false":
			cfg.NoValidate = false
		default:
			return fmt.Errorf("invalid boolean value %s, valid: true, false", value)
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	output.Success(fmt.Sprintf("Configuration updated: %s = %s", key, value))
	return nil
}

func runViewConfig(_ *cobra.Command, _ []string) error {
	allowedFormats := []string{
		output.FormatJSON.String(),
		output.FormatYAML.String(),
		output.FormatTable.String(),
	}
	if err := validateOutputFormat(allowedFormats); err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf(errFailedToLoadConfiguration, err)
	}

	contextViews := make(map[string]*output.ContextView, len(cfg.Contexts))
	for name, ctx := range cfg.Contexts {
		contextViews[name] = &output.ContextView{
			Username:  ctx.Username,
			Endpoints: ctx.Endpoints,
		}
	}

	configView := &output.ConfigView{
		CurrentContext: cfg.CurrentContext,
		LogLevel:       cfg.LogLevel,
		DefaultFormat:  cfg.DefaultFormat,
		Strict:         cfg.Strict,
		NoValidate:     cfg.NoValidate,
		Contexts:       contextViews,
	}

	if err := output.PrintConfigViewWithFormat(configView, outputFormat); err != nil {
		return fmt.Errorf("failed to print configuration: %w", err)
	}

	return nil
}
