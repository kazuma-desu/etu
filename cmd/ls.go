package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/output"
)

var (
	lsOpts struct {
		prefix bool
	}

	lsCmd = &cobra.Command{
		Use:   "ls <prefix>",
		Short: "List keys from etcd",
		Long: `List keys from etcd with prefix matching.

This is a simplified command for listing keys, equivalent to 'get --prefix --keys-only'.`,
		Example: `  # List all keys
  etu ls /

  # List keys under /app
  etu ls /app

  # List keys under /app/config (same as above, prefix is default)
  etu ls /app/config/

  # JSON output for scripting
  etu ls /app -o json

  # YAML output
  etu ls /app -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: runLs,
	}
)

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolVar(&lsOpts.prefix, "prefix", true,
		"list keys with matching prefix (default true)")
}

func runLs(_ *cobra.Command, args []string) error {
	ctx, cancel := getOperationContext()
	defer cancel()

	prefix := args[0]

	if err := validateKeyPrefix(prefix); err != nil {
		return err
	}

	if err := validateOutputFormat([]string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
		output.FormatYAML.String(),
		output.FormatTable.String(),
	}); err != nil {
		return err
	}

	etcdClient, cleanup, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer cleanup()

	opts := &client.GetOptions{
		Prefix:   lsOpts.prefix,
		KeysOnly: true,
	}

	logger.Log.Debug("Listing keys", "prefix", prefix, "options", opts)
	resp, err := etcdClient.GetWithOptions(ctx, prefix, opts)
	if err != nil {
		return err
	}

	switch outputFormat {
	case output.FormatSimple.String():
		printLsSimple(resp)
		return nil
	case output.FormatJSON.String():
		return printLsJSON(resp)
	case output.FormatYAML.String():
		return printLsYAML(resp)
	case output.FormatTable.String():
		return printLsTable(resp)
	}
	return nil
}

func printLsSimple(resp *client.GetResponse) {
	for _, kv := range resp.Kvs {
		fmt.Println(kv.Key)
	}
}

func extractKeys(resp *client.GetResponse) []string {
	keys := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		keys[i] = kv.Key
	}
	return keys
}

func printLsJSON(resp *client.GetResponse) error {
	keys := extractKeys(resp)
	data := map[string]any{
		"keys":  keys,
		"count": resp.Count,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func printLsYAML(resp *client.GetResponse) error {
	keys := extractKeys(resp)
	data := map[string]any{
		"keys":  keys,
		"count": resp.Count,
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Print(string(yamlBytes))
	return nil
}

func printLsTable(resp *client.GetResponse) error {
	headers := []string{"KEY"}
	rows := make([][]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		rows[i] = []string{kv.Key}
	}

	table := output.RenderTable(output.TableConfig{
		Headers: headers,
		Rows:    rows,
	})

	fmt.Println(table)
	return nil
}
