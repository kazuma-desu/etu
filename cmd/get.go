package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/logger"
)

var (
	getOpts struct {
		sortOrder    string
		sortTarget   string
		outputFormat string
		consistency  string
		rangeEnd     string
		limit        int64
		revision     int64
		minModRev    int64
		maxModRev    int64
		minCreateRev int64
		maxCreateRev int64
		prefix       bool
		fromKey      bool
		keysOnly     bool
		countOnly    bool
		printValue   bool
		showMetadata bool
	}

	getCmd = &cobra.Command{
		Use:   "get [options] <key> [range_end]",
		Short: "Get the key or a range of keys",
		Long: `Get the key or a range of keys from etcd.

The get command retrieves keys and their values from etcd. It supports various
options for filtering, sorting, and formatting the output. Compatible with
etcdctl get command and provides additional features.`,
		Example: `  # Get a single key
  etu get /config/app/host

  # Get all keys with a prefix
  etu get /config/app/ --prefix

  # Get keys in a range
  etu get /config/app/a /config/app/z

  # Get only keys (no values)
  etu get /config/app/ --prefix --keys-only

  # Get count of keys with prefix
  etu get /config/app/ --prefix --count-only

  # Get with limit and sorted
  etu get /config/ --prefix --limit 10 --sort-by CREATE --order DESCEND

  # Get at specific revision
  etu get /config/app/host --rev 100

  # Get with metadata in table format
  etu get /config/app/ --prefix --show-metadata -w table

  # Get only values (for scripting)
  etu get /config/app/host --print-value-only

  # Get from a key onwards
  etu get /config/app/m --from-key

  # Get with revision filters
  etu get /config/ --prefix --min-mod-revision 50 --max-mod-revision 100

  # JSON output
  etu get /config/app/ --prefix -w json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runGet,
	}
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolVar(&getOpts.prefix, "prefix", false,
		"get keys with matching prefix")
	getCmd.Flags().BoolVar(&getOpts.fromKey, "from-key", false,
		"get keys that are greater than or equal to the given key using byte compare")
	getCmd.Flags().Int64Var(&getOpts.limit, "limit", 0,
		"maximum number of results")
	getCmd.Flags().Int64Var(&getOpts.revision, "rev", 0,
		"specify the kv revision")
	getCmd.Flags().StringVar(&getOpts.sortOrder, "order", "",
		"order of results; ASCEND or DESCEND (ASCEND by default)")
	getCmd.Flags().StringVar(&getOpts.sortTarget, "sort-by", "",
		"sort target; CREATE, KEY, MODIFY, VALUE, or VERSION")
	getCmd.Flags().BoolVar(&getOpts.keysOnly, "keys-only", false,
		"get only the keys")
	getCmd.Flags().BoolVar(&getOpts.countOnly, "count-only", false,
		"get only the count")
	getCmd.Flags().BoolVar(&getOpts.printValue, "print-value-only", false,
		"only write values when using the simple output format")
	getCmd.Flags().StringVarP(&getOpts.outputFormat, "write-out", "w", "simple",
		"set the output format (simple, json, table, fields)")
	getCmd.Flags().StringVar(&getOpts.consistency, "consistency", "l",
		"linearizable(l) or serializable(s)")
	getCmd.Flags().Int64Var(&getOpts.minModRev, "min-mod-revision", 0,
		"minimum modify revision")
	getCmd.Flags().Int64Var(&getOpts.maxModRev, "max-mod-revision", 0,
		"maximum modify revision")
	getCmd.Flags().Int64Var(&getOpts.minCreateRev, "min-create-revision", 0,
		"minimum create revision")
	getCmd.Flags().Int64Var(&getOpts.maxCreateRev, "max-create-revision", 0,
		"maximum create revision")
	getCmd.Flags().BoolVar(&getOpts.showMetadata, "show-metadata", false,
		"show metadata (revisions, version, lease) in output")
}

func runGet(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	key := args[0]

	// Handle range_end if provided
	if len(args) > 1 {
		getOpts.rangeEnd = args[1]
	}

	// Connect to etcd
	logger.Log.Debug("Connecting to etcd")
	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return fmt.Errorf("failed to get etcd config: %w", err)
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer etcdClient.Close()

	// Build get options
	opts := &client.GetOptions{
		Prefix:       getOpts.prefix,
		FromKey:      getOpts.fromKey,
		Limit:        getOpts.limit,
		Revision:     getOpts.revision,
		SortOrder:    getOpts.sortOrder,
		SortTarget:   getOpts.sortTarget,
		KeysOnly:     getOpts.keysOnly,
		CountOnly:    getOpts.countOnly,
		RangeEnd:     getOpts.rangeEnd,
		MinModRev:    getOpts.minModRev,
		MaxModRev:    getOpts.maxModRev,
		MinCreateRev: getOpts.minCreateRev,
		MaxCreateRev: getOpts.maxCreateRev,
	}

	// Execute get
	logger.Log.Debugw("Fetching keys", "key", key, "options", opts)
	resp, err := etcdClient.GetWithOptions(ctx, key, opts)
	if err != nil {
		return err
	}

	// Handle count-only
	if getOpts.countOnly {
		fmt.Println(resp.Count)
		return nil
	}

	// Check if no keys found
	if len(resp.Kvs) == 0 {
		if getOpts.prefix || getOpts.fromKey || getOpts.rangeEnd != "" {
			// For range queries, empty result is not an error
			logger.Log.Debug("No keys found")
			return nil
		}
		return fmt.Errorf("key not found: %s", key)
	}

	// Output results based on format
	switch getOpts.outputFormat {
	case "simple":
		printSimple(resp)
		return nil
	case "json":
		return printJSON(resp)
	case "table":
		return printTable(resp)
	case "fields":
		printFields(resp)
		return nil
	default:
		return fmt.Errorf("invalid output format: %s (use simple, json, table, or fields)", getOpts.outputFormat)
	}
}

func printSimple(resp *client.GetResponse) {
	for _, kv := range resp.Kvs {
		switch {
		case getOpts.printValue:
			// Only print value (useful for scripting)
			fmt.Println(kv.Value)
		case getOpts.keysOnly:
			// Only print key
			fmt.Println(kv.Key)
		case getOpts.showMetadata:
			// Print with metadata
			fmt.Printf("%s\n", kv.Key)
			fmt.Printf("%s\n", kv.Value)
			fmt.Printf("CreateRevision: %d, ModRevision: %d, Version: %d\n",
				kv.CreateRevision, kv.ModRevision, kv.Version)
			if kv.Lease > 0 {
				fmt.Printf("Lease: %d\n", kv.Lease)
			}
			fmt.Println()
		default:
			// Standard key-value output (etcdctl compatible)
			fmt.Println(kv.Key)
			fmt.Println(kv.Value)
		}
	}
}

func printJSON(resp *client.GetResponse) error {
	output := map[string]any{
		"header": map[string]any{
			"cluster_id": 0, // We don't have cluster ID in our response
			"member_id":  0,
			"revision":   0,
			"raft_term":  0,
		},
		"kvs":   resp.Kvs,
		"count": resp.Count,
		"more":  resp.More,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func printTable(resp *client.GetResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Create styled header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	switch {
	case getOpts.keysOnly:
		fmt.Fprintln(w, headerStyle.Render("KEY"))
		fmt.Fprintln(w, strings.Repeat("-", 40))
		for _, kv := range resp.Kvs {
			fmt.Fprintf(w, "%s\n", kv.Key)
		}
	case getOpts.showMetadata:
		fmt.Fprintln(w, headerStyle.Render("KEY\tVALUE\tCREATE_REV\tMOD_REV\tVERSION\tLEASE"))
		fmt.Fprintln(w, strings.Repeat("-", 100))
		for _, kv := range resp.Kvs {
			value := truncateValue(kv.Value, 30)
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\n",
				kv.Key, value, kv.CreateRevision, kv.ModRevision, kv.Version, kv.Lease)
		}
	default:
		fmt.Fprintln(w, headerStyle.Render("KEY\tVALUE"))
		fmt.Fprintln(w, strings.Repeat("-", 60))
		for _, kv := range resp.Kvs {
			value := truncateValue(kv.Value, 50)
			fmt.Fprintf(w, "%s\t%s\n", kv.Key, value)
		}
	}

	return nil
}

func printFields(resp *client.GetResponse) {
	for i, kv := range resp.Kvs {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("Key:            %s\n", kv.Key)
		if !getOpts.keysOnly {
			fmt.Printf("Value:          %s\n", kv.Value)
		}
		fmt.Printf("CreateRevision: %d\n", kv.CreateRevision)
		fmt.Printf("ModRevision:    %d\n", kv.ModRevision)
		fmt.Printf("Version:        %d\n", kv.Version)
		if kv.Lease > 0 {
			fmt.Printf("Lease:          %d\n", kv.Lease)
		}
	}
}

func truncateValue(value string, maxLen int) string {
	// Replace newlines with space for table display
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")

	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen-3] + "..."
}
