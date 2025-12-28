package cmd

import (
	"context"
	"errors"
	"fmt"

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
		logger.Log.Debugw("Failed to load config, using defaults", "error", err)
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
		logger.Log.Debugw("Auto-detected format", "format", format)
	}

	parser, err := registry.GetParser(format)
	if err != nil {
		return nil, "", err
	}

	return parser, format, nil
}

func parseConfigFile(filePath string, flagFormat models.FormatType, appCfg *config.Config) ([]*models.ConfigPair, error) {
	format := resolveFormat(flagFormat, appCfg)
	parser, format, err := getParserForFile(filePath, format)
	if err != nil {
		return nil, err
	}

	logVerbose("Parsing configuration", "file", filePath, "format", format)
	pairs, err := parser.Parse(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return pairs, nil
}

func getOperationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), operationTimeout)
}

func wrapTimeoutError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("operation timed out after %v: consider increasing --timeout", operationTimeout)
	}
	return err
}

func newEtcdClient() (*client.Client, func(), error) {
	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get etcd config: %w", err)
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

func normalizeOutputFormat(supportedFormats []string) (string, error) {
	return output.NormalizeFormat(outputFormat, supportedFormats)
}

var (
	formatsWithoutTree = []string{"simple", "json", "table"}
	formatsWithTree    = []string{"simple", "json", "table", "tree"}
)

func isQuietOutput() bool {
	return outputFormat == "json"
}

func logVerbose(msg string, keyvals ...any) {
	if !isQuietOutput() {
		logger.Log.Infow(msg, keyvals...)
	}
}

func logVerboseInfo(msg string) {
	if !isQuietOutput() {
		logger.Log.Info(msg)
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
