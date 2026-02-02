package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/parsers"
)

func loadAppConfig() *config.Config {
	appCfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Debug("Failed to load config, using defaults", "error", err)
		return nil
	}
	return appCfg
}

func resolveFormat(flagFormat models.FormatType, appCfg *config.Config) models.FormatType {
	if flagFormat != "" {
		return flagFormat
	}
	if appCfg != nil && appCfg.DefaultFormat != "" {
		return models.FormatType(appCfg.DefaultFormat)
	}
	return models.FormatAuto
}

func getParserForFile(filePath string, format models.FormatType) (parsers.Parser, models.FormatType, error) {
	registry := parsers.NewRegistry()

	if format == models.FormatAuto {
		var err error
		format, err = registry.DetectFormat(filePath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to detect format: %w", err)
		}
		logger.Log.Debug("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return nil, "", err
	}

	return parser, format, nil
}

func parseConfigFile(ctx context.Context, filePath string, flagFormat models.FormatType, appCfg *config.Config) ([]*models.ConfigPair, error) {
	format := resolveFormat(flagFormat, appCfg)
	parser, format, err := getParserForFile(filePath, format)
	if err != nil {
		return nil, err
	}

	logVerbose("Parsing configuration", "file", filePath, "format", format)
	pairs, err := parser.Parse(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return pairs, nil
}

func getOperationContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigChan:
			logger.Log.Debug("Received interrupt signal, canceling operation")
			signal.Stop(sigChan)
			cancel()
		case <-ctx.Done():
			// Context canceled by timeout, clean up signal handler
			signal.Stop(sigChan)
		}
	}()

	return ctx, func() {
		signal.Stop(sigChan)
		cancel()
	}
}

func wrapContextError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("operation timed out after %v: consider increasing --timeout", operationTimeout)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("operation canceled by user")
	}
	return err
}

func newEtcdClient() (client.EtcdClient, func(), error) {
	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get etcd config: %w", err)
	}

	if overrideErr := applyGlobalOverrides(cfg); overrideErr != nil {
		return nil, nil, overrideErr
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	cleanup := func() {
		etcdClient.Close()
	}

	return etcdClient, cleanup, nil
}

func applyGlobalOverrides(cfg *client.Config) error {
	if globalCACert != "" {
		cfg.CACert = globalCACert
	}
	if globalCert != "" {
		cfg.Cert = globalCert
	}
	if globalKey != "" {
		cfg.Key = globalKey
	}
	if globalInsecureSkipTLSVerify {
		cfg.InsecureSkipTLSVerify = true
	}

	if globalPassword != "" && globalPasswordStdin {
		return fmt.Errorf("--password and --password-stdin are mutually exclusive")
	}

	if globalPasswordStdin {
		password, err := readPasswordFromStdin()
		if err != nil {
			return fmt.Errorf("failed to read password from stdin: %w", err)
		}
		cfg.Password = password
	} else if globalPassword != "" {
		cfg.Password = globalPassword
	}

	if globalUsername != "" {
		cfg.Username = globalUsername
	}

	return nil
}

func readPasswordFromStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("stdin is a terminal; use a pipe or redirect")
	}

	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.TrimSpace(password), nil
}

func newEtcdClientOrDryRun(dryRun bool) (client.EtcdClient, func(), error) {
	if dryRun {
		// DryRunClient has no resources to release, so cleanup is a no-op
		return client.NewDryRunClient(), func() { /* no-op: DryRunClient has no resources to release */ }, nil
	}
	return newEtcdClient()
}

func normalizeOutputFormat(supportedFormats []string) (string, error) {
	return output.NormalizeFormat(outputFormat, supportedFormats)
}

var (
	formatsWithoutTree = []string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
		output.FormatTable.String(),
	}
	formatsWithTree = []string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
		output.FormatTable.String(),
		output.FormatTree.String(),
	}
)

func isQuietOutput() bool {
	// Check global output format (used by most commands)
	if outputFormat == output.FormatJSON.String() {
		return true
	}
	// Check diff command's format (diff uses its own format option)
	if diffOpts.Format == output.FormatJSON.String() {
		return true
	}
	return false
}

func logVerbose(msg string, keyvals ...any) {
	if !isQuietOutput() {
		// Format keyvals into message if provided
		if len(keyvals) > 0 {
			msg = fmt.Sprintf("%s %v", msg, keyvals)
		}
		output.Info(msg)
	}
}

func logVerboseInfo(msg string) {
	if !isQuietOutput() {
		output.Info(msg)
	}
}

func resolveStrictOption(flagValue, flagChanged bool, appCfg *config.Config) bool {
	if flagChanged {
		return flagValue
	}
	if appCfg != nil {
		return appCfg.Strict
	}
	return false
}

func resolveNoValidateOption(flagValue, flagChanged bool, appCfg *config.Config) bool {
	if flagChanged {
		return flagValue
	}
	if appCfg != nil {
		return appCfg.NoValidate
	}
	return false
}
