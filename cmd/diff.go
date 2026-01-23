package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
)

var (
	diffOpts struct {
		Format        string
		Prefix        string
		FilePath      string
		ShowUnchanged bool
		Full          bool
	}

	diffCmd = &cobra.Command{
		Use:   "diff -f <file>",
		Short: "Compare local file with etcd state",
		Long: `Compare a local configuration file with the current state in etcd.

By default, only compares keys that exist in the input file (file-scoped diff).
Use --full with --prefix to compare all keys under a prefix (server-scoped diff).`,
		Example: `  # Compare only keys in file against etcd (default)
  etu diff -f config.txt

  # Filter to keys with prefix, still file-scoped
  etu diff -f config.txt --prefix /app/config

  # Full comparison: all keys under prefix (shows keys in etcd but not in file)
  etu diff -f config.txt --prefix /app/config --full

  # Include unchanged keys in output
  etu diff -f config.txt --show-unchanged

  # JSON output for scripting
  etu diff -f config.txt --format json`,
		RunE: runDiff,
	}
)

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&diffOpts.FilePath, "file", "f", "",
		"path to configuration file (required)")
	diffCmd.Flags().StringVar(&diffOpts.Format, "format", "simple",
		"output format: simple, json, table")
	diffCmd.Flags().BoolVar(&diffOpts.ShowUnchanged, "show-unchanged", false,
		"show keys that are unchanged")
	diffCmd.Flags().StringVar(&diffOpts.Prefix, "prefix", "",
		"only compare keys with this prefix")
	diffCmd.Flags().BoolVar(&diffOpts.Full, "full", false,
		"compare all keys under prefix (requires --prefix); shows keys in etcd but not in file as deleted")

	if err := diffCmd.MarkFlagRequired("file"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}

	registerFileCompletion(diffCmd, "file")
}

func runDiff(_ *cobra.Command, _ []string) error {
	// Validate flags: --full requires --prefix
	if diffOpts.Full && diffOpts.Prefix == "" {
		return fmt.Errorf("--full requires --prefix to scope the comparison\nHint: etu diff -f %s --full --prefix /your/prefix", diffOpts.FilePath)
	}

	ctx, cancel := getOperationContext()
	defer cancel()

	appCfg := loadAppConfig()

	// Parse the configuration file
	pairs, err := parseConfigFile(diffOpts.FilePath, models.FormatAuto, appCfg)
	if err != nil {
		return err
	}
	logVerboseInfo(fmt.Sprintf("Parsed %d configuration items from file", len(pairs)))

	// Filter by prefix if specified
	if diffOpts.Prefix != "" {
		filtered := make([]*models.ConfigPair, 0)
		for _, p := range pairs {
			if strings.HasPrefix(p.Key, diffOpts.Prefix) {
				filtered = append(filtered, p)
			}
		}
		pairs = filtered
		logVerboseInfo(fmt.Sprintf("Filtered to %d items with prefix %s", len(pairs), diffOpts.Prefix))
	}

	// Fetch current etcd state
	etcdClient, cleanup, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer cleanup()

	// Get etcd state based on mode:
	// - Default: fetch only exact keys from file
	// - Full mode: fetch all keys under prefix
	var etcdPairs []*models.ConfigPair
	if diffOpts.Full {
		etcdPairs, err = fetchEtcdStateByPrefix(ctx, etcdClient, diffOpts.Prefix)
	} else {
		etcdPairs, err = fetchEtcdStateForExactKeys(ctx, etcdClient, pairs)
	}
	if err != nil {
		return err
	}

	logVerboseInfo(fmt.Sprintf("Fetched %d items from etcd", len(etcdPairs)))

	// Compute diff
	fileMap := make(map[string]string)
	for _, p := range pairs {
		fileMap[p.Key] = formatValue(p.Value)
	}

	etcdMap := make(map[string]string)
	for _, p := range etcdPairs {
		etcdMap[p.Key] = formatValue(p.Value)
	}

	result := output.DiffKeyValues(fileMap, etcdMap)

	return output.PrintDiffResult(result, diffOpts.Format, diffOpts.ShowUnchanged)
}

func fetchEtcdStateForExactKeys(ctx context.Context, etcdClient client.EtcdClient, filePairs []*models.ConfigPair) ([]*models.ConfigPair, error) {
	if len(filePairs) == 0 {
		return nil, nil
	}

	result := make([]*models.ConfigPair, 0, len(filePairs))
	for _, p := range filePairs {
		resp, err := etcdClient.GetWithOptions(ctx, p.Key, &client.GetOptions{Prefix: false})
		if err != nil {
			return nil, fmt.Errorf("failed to get key %s: %w", p.Key, err)
		}
		if len(resp.Kvs) > 0 {
			kv := resp.Kvs[0]
			result = append(result, &models.ConfigPair{Key: kv.Key, Value: kv.Value})
		}
	}
	return result, nil
}

func fetchEtcdStateByPrefix(ctx context.Context, etcdClient client.EtcdClient, prefix string) ([]*models.ConfigPair, error) {
	resp, err := etcdClient.GetWithOptions(ctx, prefix, &client.GetOptions{Prefix: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get keys with prefix %s: %w", prefix, err)
	}

	result := make([]*models.ConfigPair, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		result = append(result, &models.ConfigPair{Key: kv.Key, Value: kv.Value})
	}
	return result, nil
}

func formatValue(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
