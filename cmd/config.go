package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
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
	Run: runGetContexts,
}

var currentContextCmd = &cobra.Command{
	Use:   "current-context",
	Short: "Display the current active context",
	Long:  `Display the name of the currently active context.`,
	Run:   runCurrentContext,
}

var useContextCmd = &cobra.Command{
	Use:   "use-context [context-name]",
	Short: "Switch to a different context",
	Long: `Switch the active context to a different saved configuration.

Examples:
  # Switch to production context
  etu config use-context prod

  # Switch to development context
  etu config use-context dev`,
	Args: cobra.ExactArgs(1),
	Run:  runUseContext,
}

var deleteContextCmd = &cobra.Command{
	Use:   "delete-context [context-name]",
	Short: "Delete a context",
	Long: `Delete a saved context from the configuration.

Examples:
  # Delete a context
  etu config delete-context old-dev`,
	Args: cobra.ExactArgs(1),
	Run:  runDeleteContext,
}

var viewConfigCmd = &cobra.Command{
	Use:   "view",
	Short: "View the current configuration",
	Long:  `Display the current configuration file contents with sensitive information redacted.`,
	Example: `  # View current configuration
  etu config view

  # View configuration in JSON format
  etu config view -o json

  # View configuration in table format
  etu config view -o table`,
	Run: runViewConfig,
}

var setConfigCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value in the config file.

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
	Run:  runSetConfig,
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

func runGetContexts(_ *cobra.Command, _ []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalw("Failed to load configuration", "error", err)
	}

	if len(cfg.Contexts) == 0 {
		output.Info("No contexts found. Use 'etu login <context-name>' to create one.")
		return
	}

	// Normalize format (tree not supported for config get-contexts)
	supportedFormats := []string{"simple", "json", "table"}
	normalizedFormat, err := output.NormalizeFormat(outputFormat, supportedFormats)
	if err != nil {
		logger.Log.Fatalw("Invalid output format", "error", err)
	}

	// Print contexts in requested format
	if err := output.PrintContextsWithFormat(cfg.Contexts, cfg.CurrentContext, normalizedFormat); err != nil {
		logger.Log.Fatalw("Failed to print contexts", "error", err)
	}
}

func runCurrentContext(_ *cobra.Command, _ []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalw("Failed to load configuration", "error", err)
	}

	if cfg.CurrentContext == "" {
		output.Info("No current context set. Use 'etu login <context-name>' or 'etu config use-context <context-name>'.")
		return
	}

	fmt.Print(cfg.CurrentContext)
}

func runUseContext(_ *cobra.Command, args []string) {
	ctxName := args[0]

	if err := config.UseContext(ctxName); err != nil {
		logger.Log.Fatalw("Failed to switch context", "error", err)
	}

	output.Success(fmt.Sprintf("Switched to context '%s'", ctxName))
}

func runDeleteContext(_ *cobra.Command, args []string) {
	ctxName := args[0]

	if err := config.DeleteContext(ctxName); err != nil {
		logger.Log.Fatalw("Failed to delete context", "error", err)
	}

	output.Success(fmt.Sprintf("Context '%s' deleted", ctxName))
}

func runSetConfig(_ *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalw("Failed to load configuration", "error", err)
	}

	switch key {
	case "log-level":
		// Validate log level
		validLevels := []string{"debug", "info", "warn", "error"}
		if !slices.Contains(validLevels, value) {
			logger.Log.Fatalw("Invalid log level", "level", value, "valid", "debug, info, warn, error")
		}
		cfg.LogLevel = value
	case "default-format":
		// Validate format
		validFormats := []string{"auto", "etcdctl"}
		if !slices.Contains(validFormats, value) {
			logger.Log.Fatalw("Invalid format", "format", value, "valid", "auto, etcdctl")
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
			logger.Log.Fatalw("Invalid boolean value", "value", value, "valid", "true, false")
		}
	case "no-validate":
		// Parse boolean
		switch value {
		case "true":
			cfg.NoValidate = true
		case "false":
			cfg.NoValidate = false
		default:
			logger.Log.Fatalw("Invalid boolean value", "value", value, "valid", "true, false")
		}
	default:
		logger.Log.Fatalw("Unknown configuration key", "key", key)
	}

	if err := config.SaveConfig(cfg); err != nil {
		logger.Log.Fatalw("Failed to save configuration", "error", err)
	}

	output.Success(fmt.Sprintf("Configuration updated: %s = %s", key, value))
}

func runViewConfig(_ *cobra.Command, _ []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalw("Failed to load configuration", "error", err)
	}

	// Normalize format (tree not supported for config view)
	supportedFormats := []string{"simple", "json", "table"}
	normalizedFormat, err := output.NormalizeFormat(outputFormat, supportedFormats)
	if err != nil {
		logger.Log.Fatalw("Invalid output format", "error", err)
	}

	// Print config in requested format
	if err := output.PrintConfigViewWithFormat(cfg, normalizedFormat); err != nil {
		logger.Log.Fatalw("Failed to print configuration", "error", err)
	}
}
