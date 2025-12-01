package cmd

import (
	"fmt"
	"sort"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
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
	Run:   runGetContexts,
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
	Run:   runViewConfig,
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

func runGetContexts(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	if len(cfg.Contexts) == 0 {
		output.Info("No contexts found. Use 'etu login <context-name>' to create one.")
		return
	}

	// Sort context names for consistent output
	var contextNames []string
	for name := range cfg.Contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)

	// Print header
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("CONTEXTS"))
	fmt.Println()

	// Print each context
	for _, name := range contextNames {
		ctx := cfg.Contexts[name]

		// Mark current context
		marker := "  "
		if name == cfg.CurrentContext {
			marker = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("* ")
		}

		contextName := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true).Render(name)
		fmt.Printf("%s%s\n", marker, contextName)

		// Print endpoints
		for _, endpoint := range ctx.Endpoints {
			fmt.Printf("    %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(endpoint))
		}

		// Print username if set
		if ctx.Username != "" {
			username := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("User: %s", ctx.Username))
			fmt.Printf("    %s\n", username)
		}

		fmt.Println()
	}
}

func runCurrentContext(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	if cfg.CurrentContext == "" {
		output.Info("No current context set. Use 'etu login <context-name>' or 'etu config use-context <context-name>'.")
		return
	}

	fmt.Print(cfg.CurrentContext)
}

func runUseContext(cmd *cobra.Command, args []string) {
	contextName := args[0]

	if err := config.UseContext(contextName); err != nil {
		log.Fatal("Failed to switch context", "error", err)
	}

	output.Success(fmt.Sprintf("Switched to context '%s'", contextName))
}

func runDeleteContext(cmd *cobra.Command, args []string) {
	contextName := args[0]

	if err := config.DeleteContext(contextName); err != nil {
		log.Fatal("Failed to delete context", "error", err)
	}

	output.Success(fmt.Sprintf("Context '%s' deleted", contextName))
}

func runSetConfig(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	switch key {
	case "log-level":
		// Validate log level
		validLevels := []string{"debug", "info", "warn", "error"}
		valid := false
		for _, level := range validLevels {
			if value == level {
				valid = true
				break
			}
		}
		if !valid {
			log.Fatal("Invalid log level", "level", value, "valid", "debug, info, warn, error")
		}
		cfg.LogLevel = value
	case "default-format":
		// Validate format
		validFormats := []string{"auto", "etcdctl"}
		valid := false
		for _, format := range validFormats {
			if value == format {
				valid = true
				break
			}
		}
		if !valid {
			log.Fatal("Invalid format", "format", value, "valid", "auto, etcdctl")
		}
		cfg.DefaultFormat = value
	case "strict":
		// Parse boolean
		if value == "true" {
			cfg.Strict = true
		} else if value == "false" {
			cfg.Strict = false
		} else {
			log.Fatal("Invalid boolean value", "value", value, "valid", "true, false")
		}
	case "no-validate":
		// Parse boolean
		if value == "true" {
			cfg.NoValidate = true
		} else if value == "false" {
			cfg.NoValidate = false
		} else {
			log.Fatal("Invalid boolean value", "value", value, "valid", "true, false")
		}
	default:
		log.Fatal("Unknown configuration key", "key", key)
	}

	if err := config.SaveConfig(cfg); err != nil {
		log.Fatal("Failed to save configuration", "error", err)
	}

	output.Success(fmt.Sprintf("Configuration updated: %s = %s", key, value))
}

func runViewConfig(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	configPath, _ := config.GetConfigPath()

	// Print header
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(fmt.Sprintf("Configuration: %s", configPath)))
	fmt.Println()

	// Print current context
	if cfg.CurrentContext != "" {
		fmt.Printf("%s %s\n\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Current Context:"),
			lipgloss.NewStyle().Bold(true).Render(cfg.CurrentContext))
	}

	// Print settings
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Settings:"))

	logLevelDisplay := cfg.LogLevel
	if logLevelDisplay == "" {
		logLevelDisplay = "warn (default)"
	}
	fmt.Printf("  log-level: %s\n", logLevelDisplay)

	formatDisplay := cfg.DefaultFormat
	if formatDisplay == "" {
		formatDisplay = "auto (default)"
	}
	fmt.Printf("  default-format: %s\n", formatDisplay)

	fmt.Printf("  strict: %v\n", cfg.Strict)
	fmt.Printf("  no-validate: %v\n", cfg.NoValidate)
	fmt.Println()

	// Print contexts
	if len(cfg.Contexts) == 0 {
		output.Info("No contexts configured")
		return
	}

	// Sort context names
	var contextNames []string
	for name := range cfg.Contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)

	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Contexts:"))
	for _, name := range contextNames {
		ctx := cfg.Contexts[name]

		fmt.Printf("\n  %s\n", lipgloss.NewStyle().Bold(true).Render(name))
		fmt.Printf("    endpoints: %v\n", ctx.Endpoints)

		if ctx.Username != "" {
			fmt.Printf("    username: %s\n", ctx.Username)
		}

		if ctx.Password != "" {
			fmt.Printf("    password: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("[REDACTED]"))
		}
	}
	fmt.Println()
}
