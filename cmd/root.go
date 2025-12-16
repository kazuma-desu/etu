package cmd

import (
	"fmt"
	"os"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	logLevel    string
	contextName string

	rootCmd = &cobra.Command{
		Use:   "etu",
		Short: "Etcd Terminal Utility",
		Long: `etu (Etcd Terminal Utility) is a CLI tool for managing etcd configurations from multiple file formats.

It provides a familiar interface similar to kubectl for parsing, validating, and
applying etcd configuration from various sources.`,
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
