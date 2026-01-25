package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"

	"github.com/spf13/cobra"
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

	rootCmd = &cobra.Command{
		Use:   "etu",
		Short: "Etcd Terminal Utility - kubectl-like CLI for etcd",
		Long:  `A CLI tool for managing etcd configurations with kubectl-like UX.`,
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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "simple",
		"output format (simple, json, table, tree)")
	rootCmd.PersistentFlags().DurationVar(&operationTimeout, "timeout", defaultOperationTimeout,
		"timeout for etcd operations (e.g., 30s, 1m, 2m30s)")
	rootCmd.PersistentFlags().StringVar(&globalCACert, "cacert", "",
		"path to CA certificate (overrides context)")
	rootCmd.PersistentFlags().StringVar(&globalCert, "cert", "",
		"path to client certificate (overrides context)")
	rootCmd.PersistentFlags().StringVar(&globalKey, "key", "",
		"path to client key (overrides context)")
	rootCmd.PersistentFlags().BoolVar(&globalInsecureSkipTLSVerify, "insecure-skip-tls-verify", false,
		"skip TLS verification (overrides context)")
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
