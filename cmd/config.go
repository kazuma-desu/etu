package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage etu configuration",
	Long: `Manage etu configuration including contexts and settings.

Available subcommands:
  get-contexts     - List all available contexts
  current-context  - Display the current active context
  use-context      - Switch to a different context
  delete-context   - Delete a context
  set              - Set a configuration value
  view             - View the current configuration`,
}

var getContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "List all available contexts",
	Long:  `List all saved contexts with their connection details.`,
	Example: `  # List all contexts
  etu config get-contexts

  # List contexts in JSON format
  etu config get-contexts -o json

  # List contexts in table format
  etu config get-contexts -o table`,
	RunE: runGetContexts,
}

var currentContextCmd = &cobra.Command{
	Use:   "current-context",
	Short: "Display current active context",
	Long:  `Display name of currently active context.`,
	RunE:  runCurrentContext,
}

var useContextCmd = &cobra.Command{
	Use:   "use-context [context-name]",
	Short: "Switch to a different context",
	Long: `Switch active context to a different saved configuration.

Examples:
  # Switch to production context
  etu config use-context prod

  # Switch to development context
  etu config use-context dev`,
	Args: cobra.ExactArgs(1),
	RunE: runUseContext,
}

var deleteContextCmd = &cobra.Command{
	Use:   "delete-context [context-name]",
	Short: "Delete a context",
	Long: `Delete a saved context from configuration.

Examples:
  # Delete a context
  etu config delete-context old-dev`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteContext,
}

var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long:  `Display current configuration file contents with sensitive information redacted.`,
	Example: `  # View current configuration
  etu config view

  # View configuration in JSON format
  etu config view -o json

  # View configuration in table format
  etu config view -o table`,
	RunE: runViewConfig,
}

var setConfigCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value in config file.

Available settings:
  log-level       - Default log level (debug, info, warn, error)
  default-format  - Default file format (auto, etcdctl)
  strict          - Enable strict validation by default (true, false)
  no-validate     - Skip validation by default (true, false)

Examples:
  # Set default log level to debug
  etu config set log-level debug

  # Set default format to etcdctl
  etu config set default-format etcdctl

  # Enable strict validation by default
  etu config set strict true

  # Disable validation by default (not recommended)
  etu config set no-validate true`,
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
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.Contexts) == 0 {
		output.Info("No contexts found. Use 'etu login <context-name>' to create one.")
		return nil
	}

	// Normalize format (tree not supported for config get-contexts)
	supportedFormats := []string{"simple", "json", "table"}
	normalizedFormat, err := output.NormalizeFormat(outputFormat, supportedFormats)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Print contexts in requested format
	if err := output.PrintContextsWithFormat(cfg.Contexts, cfg.CurrentContext, normalizedFormat); err != nil {
		return fmt.Errorf("failed to print contexts: %w", err)
	}

	return nil
}

func runCurrentContext(_ *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
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
		return fmt.Errorf("failed to load configuration: %w", err)
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
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Normalize format (tree not supported for config view)
	supportedFormats := []string{"simple", "json", "table"}
	normalizedFormat, err := output.NormalizeFormat(outputFormat, supportedFormats)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Print config in requested format
	if err := output.PrintConfigViewWithFormat(cfg, normalizedFormat); err != nil {
		return fmt.Errorf("failed to print configuration: %w", err)
	}

	return nil
}
