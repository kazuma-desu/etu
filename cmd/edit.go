package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/output"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit <key>",
		Short: "Edit a key's value in your $EDITOR",
		Long:  `Fetch a key's value, open in $EDITOR, and save changes back to etcd.`,
		Example: `  # Edit a key's value
  etu edit /config/app/database/host`,
		Args: cobra.ExactArgs(1),
		RunE: runEdit,
	}
)

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(_ *cobra.Command, args []string) error {
	key := args[0]

	logVerboseInfo("Connecting to etcd")
	etcdClient, cleanup, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer cleanup()

	logVerbose("Fetching current value", "key", key)
	getCtx, getCancel := getOperationContext()
	value, err := etcdClient.Get(getCtx, key)
	getCancel()
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
		return fmt.Errorf("✗ no editor found: set $EDITOR or $VISUAL environment variable")
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
	logVerbose("Opening editor", "editor", editor, "file", filepath.Base(tmpPath))
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
		logVerboseInfo("No changes detected, skipping update")
		output.Info("No changes made")
		return nil
	}

	// Read modified value
	modifiedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read modified file: %w", err)
	}
	newValue := strings.TrimRight(string(modifiedContent), "\r\n")

	logVerbose("Updating key in etcd", "key", key)
	putCtx, putCancel := getOperationContext()
	defer putCancel()
	if err := etcdClient.Put(putCtx, key, newValue); err != nil {
		return fmt.Errorf("failed to update key %q: %w", key, err)
	}

	logVerbose("Successfully updated key", "key", key)
	output.Success(fmt.Sprintf("Updated %s", key))

	return nil
}
