package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/output"
)

const defaultOperationTimeout = 30 * time.Second

var (
	logLevel                    string
	contextName                 string
	outputFormat                string
	operationTimeout            time.Duration
	globalCACert                string
	globalCert                  string
	globalKey                   string
	globalInsecureSkipTLSVerify bool
	globalUsername              string
	globalPassword              string
	globalPasswordStdin         bool

	rootCmd = &cobra.Command{
		Use:   "etu",
		Short: "Etcd Terminal Utility - kubectl-like CLI for etcd",
		Long: `A CLI tool for managing etcd configurations with kubectl-like UX.

Exit Codes:
  0  Success
  1  General error
  2  Validation error (invalid input, missing arguments)
  3  Connection error (failed to connect to etcd)
  4  Key not found

Use 'etu options' to see all available global flags.`,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			configureLogging()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "",
		"log level (debug, info, warn, error) - overrides config file")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "",
		"context to use for etcd connection (overrides current context)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", output.FormatSimple.String(),
		fmt.Sprintf("output format (%s)", strings.Join(formatNames(), ", ")))
	rootCmd.PersistentFlags().DurationVar(&operationTimeout, "timeout", defaultOperationTimeout,
		"timeout for etcd operations (e.g., 30s, 1m, 2m30s)")

	// Auth/TLS flags - hidden from main help, visible via 'etu options'
	rootCmd.PersistentFlags().StringVar(&globalCACert, "cacert", "",
		"path to CA certificate (overrides context)")
	rootCmd.PersistentFlags().StringVar(&globalCert, "cert", "",
		"path to client certificate (overrides context)")
	rootCmd.PersistentFlags().StringVar(&globalKey, "key", "",
		"path to client key (overrides context)")
	rootCmd.PersistentFlags().BoolVar(&globalInsecureSkipTLSVerify, "insecure-skip-tls-verify", false,
		"skip TLS verification (overrides context)")
	rootCmd.PersistentFlags().StringVar(&globalUsername, "username", "",
		"username for etcd authentication (overrides context)")
	rootCmd.PersistentFlags().StringVar(&globalPassword, "password", "",
		"password for etcd authentication (overrides context)")
	rootCmd.PersistentFlags().BoolVar(&globalPasswordStdin, "password-stdin", false,
		"read password from stdin (mutually exclusive with --password)")

	// Hide all global flags from main help - use 'etu options' to see them
	hideAllGlobalFlags()
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func configureLogging() {
	effectiveLogLevel := "warn"

	cfg, err := config.LoadConfig()
	if err == nil && cfg.LogLevel != "" {
		effectiveLogLevel = cfg.LogLevel
	}

	if logLevel != "" {
		effectiveLogLevel = logLevel
	}

	logger.SetLevel(effectiveLogLevel)
}

// formatNames returns all available output format names for flag help.
func formatNames() []string {
	formats := output.AllFormats()
	names := make([]string, len(formats))
	for i, f := range formats {
		names[i] = f.String()
	}
	return names
}

// hideAllGlobalFlags hides most persistent flags from the main help output.
// Commonly-used flags like --output and --context are kept visible for discoverability.
// Use 'etu options' to see all available global flags.
func hideAllGlobalFlags() {
	visibleFlags := map[string]bool{
		"output":  true,
		"context": true,
	}
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if !visibleFlags[f.Name] {
			_ = rootCmd.PersistentFlags().MarkHidden(f.Name)
		}
	})
}
