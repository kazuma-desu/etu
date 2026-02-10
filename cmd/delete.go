package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/output"
)

var (
	deleteOpts struct {
		prefix bool
		force  bool
		dryRun bool
	}

	deleteCmd = &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete keys from etcd",
		Long:  `Delete a single key or all keys with a prefix from etcd.`,
		Example: `  # Delete single key
  etu delete /app/config/host

  # Delete all keys with prefix (requires confirmation)
  etu delete /app/config/ --prefix

  # Skip confirmation
  etu delete /app/config/ --prefix --force

  # Preview what would be deleted
  etu delete /app/config/ --prefix --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}
)

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVar(&deleteOpts.prefix, "prefix", false,
		"delete all keys with the given prefix")
	deleteCmd.Flags().BoolVar(&deleteOpts.force, "force", false,
		"skip confirmation prompt for prefix deletion")
	deleteCmd.Flags().BoolVar(&deleteOpts.dryRun, "dry-run", false,
		"preview what would be deleted without actually deleting")
}

func runDelete(_ *cobra.Command, args []string) error {
	ctx, cancel := getOperationContext()
	defer cancel()

	key := args[0]

	if err := validateKeyPrefix(key); err != nil {
		return err
	}

	if deleteOpts.prefix {
		return runDeletePrefix(ctx, key)
	}

	return runDeleteSingle(ctx, key)
}

func runDeleteSingle(ctx context.Context, key string) error {
	if deleteOpts.dryRun {
		output.Info(fmt.Sprintf("Would delete: %s", key))
		return nil
	}

	etcdClient, cleanup, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer cleanup()

	deleted, err := etcdClient.Delete(ctx, key)
	if err != nil {
		return wrapContextError(fmt.Errorf("failed to delete key: %w", err))
	}

	if deleted == 0 {
		output.Warning(fmt.Sprintf("Key not found: %s", key))
	} else {
		output.Success(fmt.Sprintf("Deleted: %s", key))
	}

	return nil
}

func runDeletePrefix(ctx context.Context, prefix string) error {
	etcdClient, cleanup, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer cleanup()

	keys, err := fetchKeysWithPrefix(ctx, etcdClient, prefix)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		output.Warning(fmt.Sprintf("No keys found with prefix: %s", prefix))
		return nil
	}

	if deleteOpts.dryRun {
		printKeysToDelete(keys, prefix)
		return nil
	}

	if !deleteOpts.force {
		if !confirmDeletion(keys, prefix, os.Stdin, os.Stdout) {
			output.Info("Deletion canceled")
			return nil
		}
	}

	deleted, err := etcdClient.DeletePrefix(ctx, prefix)
	if err != nil {
		return wrapContextError(fmt.Errorf("failed to delete prefix: %w", err))
	}

	output.Success(fmt.Sprintf("Deleted %d keys with prefix: %s", deleted, prefix))
	return nil
}

func fetchKeysWithPrefix(ctx context.Context, etcdClient client.EtcdClient, prefix string) ([]string, error) {
	resp, err := etcdClient.GetWithOptions(ctx, prefix, &client.GetOptions{
		Prefix:   true,
		KeysOnly: true,
	})
	if err != nil {
		return nil, wrapContextError(fmt.Errorf("failed to fetch keys: %w", err))
	}

	keys := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		keys[i] = kv.Key
	}
	return keys, nil
}

func printKeysToDelete(keys []string, prefix string) {
	output.Info(fmt.Sprintf("Would delete %d keys with prefix %q:", len(keys), prefix))
	for _, k := range keys {
		fmt.Printf("  %s\n", k)
	}
}

func confirmDeletion(keys []string, prefix string, in io.Reader, out io.Writer) bool {
	fmt.Fprintf(out, "The following %d keys will be deleted:\n", len(keys))
	for _, k := range keys {
		fmt.Fprintf(out, "  %s\n", k)
	}
	fmt.Fprintf(out, "\nDelete all keys with prefix %q? [y/N]: ", prefix)

	scanner := bufio.NewScanner(in)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}
	return false
}
