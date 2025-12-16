package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit KEY",
		Short: "Edit a key's value in your $EDITOR",
		Long: `Edit a key's value from etcd in your preferred text editor.

The edit command fetches the current value of a key from etcd, opens it in
your $EDITOR (or vi/nano as fallback), and writes the modified value back
to etcd after you save and close the editor.`,
		Example: `  # Edit a key's value
  etu edit /config/app/database/host

  # Edit a key using a specific context
  etu edit /config/app/database/host --context production`,
		Args: cobra.ExactArgs(1),
		RunE: runEdit,
	}
)

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	key := args[0]

	// Get etcd config with context
	log.Info("Connecting to etcd")
	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return fmt.Errorf("failed to get etcd config: %w", err)
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer etcdClient.Close()

	// Fetch current value
	log.Info("Fetching current value", "key", key)
	value, err := etcdClient.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get key %q: %w", key, err)
	}

	// Determine editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Fallback to common editors
		for _, fallback := range []string{"vi", "vim", "nano", "emacs"} {
			if _, lookupErr := exec.LookPath(fallback); lookupErr == nil {
				editor = fallback
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found: set $EDITOR or $VISUAL environment variable")
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "etu-edit-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write current value to temp file
	if _, writeErr := tmpFile.WriteString(value); writeErr != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", writeErr)
	}
	tmpFile.Close()

	// Get initial file info for change detection
	initialStat, err := os.Stat(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to stat temporary file: %w", err)
	}
	initialModTime := initialStat.ModTime()

	// Open editor
	log.Info("Opening editor", "editor", editor, "file", filepath.Base(tmpPath))
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if runErr := editorCmd.Run(); runErr != nil {
		return fmt.Errorf("editor exited with error: %w", runErr)
	}

	// Check if file was modified
	finalStat, err := os.Stat(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to stat temporary file after editing: %w", err)
	}
	finalModTime := finalStat.ModTime()

	if finalModTime.Equal(initialModTime) {
		log.Info("No changes detected, skipping update")
		fmt.Println("No changes made.")
		return nil
	}

	// Read modified value
	modifiedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read modified file: %w", err)
	}
	newValue := string(modifiedContent)

	// Write back to etcd
	log.Info("Updating key in etcd", "key", key)
	if err := etcdClient.Put(ctx, key, newValue); err != nil {
		return fmt.Errorf("failed to update key %q: %w", key, err)
	}

	log.Info("Successfully updated key", "key", key)
	fmt.Printf("Successfully updated %s\n", key)

	return nil
}
