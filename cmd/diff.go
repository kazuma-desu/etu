package cmd

import (
	"context"
	"fmt"
	"sort"
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
	}

	diffCmd = &cobra.Command{
		Use:   "diff -f <file>",
		Short: "Compare local file with etcd state",
		Long: `Compare a local configuration file with the current state in etcd.
Shows which keys would be added, modified, or deleted.`,
		Example: `  # Compare config file with etcd
  etu diff -f config.txt

  # Include unchanged keys
  etu diff -f config.txt --show-unchanged

  # JSON output for scripting
  etu diff -f config.txt -o json

  # Compare with prefix filter
  etu diff -f config.txt --prefix /app/config`,
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

	if err := diffCmd.MarkFlagRequired("file"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}

	registerFileCompletion(diffCmd, "file")
}

func runDiff(_ *cobra.Command, _ []string) error {
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

	// Get all keys from etcd that match our file keys
	etcdPairs, err := fetchEtcdStateForKeys(ctx, etcdClient, pairs, diffOpts.Prefix)
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

// fetchEtcdStateForKeys fetches current values from etcd for the given keys
func fetchEtcdStateForKeys(ctx context.Context, etcdClient client.EtcdClient, filePairs []*models.ConfigPair, prefix string) ([]*models.ConfigPair, error) {
	if len(filePairs) == 0 {
		return nil, nil
	}

	// Build a map for quick lookup
	etcdMap := make(map[string]*models.ConfigPair)

	// If we have a prefix, use GetWithOptions with prefix
	if prefix != "" {
		resp, err := etcdClient.GetWithOptions(ctx, prefix, &client.GetOptions{Prefix: true})
		if err != nil {
			return nil, err
		}
		for _, kv := range resp.Kvs {
			etcdMap[kv.Key] = &models.ConfigPair{Key: kv.Key, Value: kv.Value}
		}
	} else {
		// Otherwise, fetch by extracting common prefixes from file keys
		prefixes := extractPrefixes(filePairs)
		for _, p := range prefixes {
			resp, err := etcdClient.GetWithOptions(ctx, p, &client.GetOptions{Prefix: true})
			if err != nil {
				return nil, err
			}
			for _, kv := range resp.Kvs {
				etcdMap[kv.Key] = &models.ConfigPair{Key: kv.Key, Value: kv.Value}
			}
		}
	}

	// Convert map to slice
	result := make([]*models.ConfigPair, 0, len(etcdMap))
	for _, v := range etcdMap {
		result = append(result, v)
	}
	return result, nil
}

// extractPrefixes extracts unique parent prefixes from keys
func extractPrefixes(pairs []*models.ConfigPair) []string {
	prefixSet := make(map[string]bool)
	for _, p := range pairs {
		parts := strings.Split(strings.Trim(p.Key, "/"), "/")
		if len(parts) > 1 {
			prefix := "/" + parts[0]
			prefixSet[prefix] = true
		}
	}

	prefixes := make([]string, 0, len(prefixSet))
	for p := range prefixSet {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)
	return prefixes
}

// formatValue converts a value to a display string
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
