package cmd

import (
	"fmt"
	"os"

	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/charmbracelet/log"
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

// configureLogging sets up the logger with the specified level
// Priority: flag > config file > default (warn)
func configureLogging() {
	effectiveLogLevel := "warn" // Default

	// Try to load from config file
	cfg, err := config.LoadConfig()
	if err == nil && cfg.LogLevel != "" {
		effectiveLogLevel = cfg.LogLevel
	}

	// Flag overrides config file (if flag was explicitly set)
	if logLevel != "" {
		effectiveLogLevel = logLevel
	}

	var level log.Level
	switch effectiveLogLevel {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
		level = log.WarnLevel
	}

	log.SetLevel(level)
	log.SetReportTimestamp(false)
}
