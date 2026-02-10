package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"
)

var (
	watchOpts struct {
		prefix bool
		rev    int64
		prevKV bool
	}

	watchCmd = &cobra.Command{
		Use:   "watch <key>",
		Short: "Watch for changes on a key or prefix",
		Long: `Watch for changes on a key or prefix in real-time.

Monitors etcd for PUT and DELETE events on the specified key.
Use --prefix to watch all keys with a given prefix.
Use -o flag to control output format (simple=raw value, json=full event).
Press Ctrl+C to stop watching.`,
		Example: `  # Watch a single key
  etu watch /config/app/host

  # Watch all keys with a prefix
  etu watch /config/app/ --prefix

  # Watch from a specific revision
  etu watch /config/app/ --prefix --rev 100

  # JSON output for scripting
  etu watch /config/app/ --prefix -o json`,
		Args: cobra.ExactArgs(1),
		RunE: runWatch,
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolVar(&watchOpts.prefix, "prefix", false,
		"watch all keys with the given prefix")
	watchCmd.Flags().Int64Var(&watchOpts.rev, "rev", 0,
		"revision to start watching from (0 = current)")
	watchCmd.Flags().BoolVar(&watchOpts.prevKV, "prev-kv", false,
		"include previous key-value pair in events")
}

func runWatch(_ *cobra.Command, args []string) error {
	allowedFormats := []string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
	}
	if err := validateOutputFormat(allowedFormats); err != nil {
		return err
	}

	key := args[0]

	if watchOpts.rev < 0 {
		return fmt.Errorf("✗ invalid --rev: must be non-negative")
	}

	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return wrapNotConnectedError(err)
	}

	etcdClient, cleanup, err := newEtcdClient(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	// Setup context with cancellation on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer signal.Stop(sigChan)
		select {
		case <-sigChan:
			output.Info("Stopping watch...")
			cancel()
		case <-ctx.Done():
			// Context canceled elsewhere, exit cleanly
		}
	}()

	opts := &client.WatchOptions{
		Prefix:   watchOpts.prefix,
		Revision: watchOpts.rev,
		PrevKV:   watchOpts.prevKV,
	}

	if outputFormat != output.FormatJSON.String() {
		if watchOpts.prefix {
			output.Info(fmt.Sprintf("Watching keys with prefix: %s", key))
		} else {
			output.Info(fmt.Sprintf("Watching key: %s", key))
		}
		fmt.Println("Press Ctrl+C to stop")
		fmt.Println()
	}

	watchChan := etcdClient.Watch(ctx, key, opts)

	for resp := range watchChan {
		if resp.Err != nil {
			return fmt.Errorf("watch error: %w", resp.Err)
		}

		if resp.CompactRevision > 0 {
			return fmt.Errorf("✗ watch canceled: revision %d has been compacted", resp.CompactRevision)
		}

		for _, event := range resp.Events {
			if err := printWatchEvent(event); err != nil {
				return err
			}
		}
	}

	return nil
}

func printWatchEvent(event client.WatchEvent) error {
	if outputFormat == output.FormatJSON.String() {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		fmt.Println(string(data))
	} else {
		// Simple format: print only the raw value (matches etcdctl)
		fmt.Println(event.Value)
	}
	return nil
}
