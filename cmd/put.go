package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/validator"
)

var (
	putOpts struct {
		dryRun   bool
		validate bool
	}

	putCmd = &cobra.Command{
		Use:   "put <key> [value]",
		Short: "Put a key-value pair into etcd",
		Long:  `Write a single key-value pair to etcd. Value can be provided as argument or via stdin using '-'.`,
		Example: `  # Put with inline value
  etu put /app/config/host "localhost"

  # Put from stdin
  echo "my-value" | etu put /app/config/name -

  # Multi-line value from stdin
  cat config.json | etu put /app/config/settings -

  # Preview without writing
  etu put /app/config/host "localhost" --dry-run

  # Validate before writing
  etu put /app/config/host "localhost" --validate`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runPut,
	}
)

func init() {
	rootCmd.AddCommand(putCmd)

	putCmd.Flags().BoolVar(&putOpts.dryRun, "dry-run", false,
		"preview the operation without writing to etcd")
	putCmd.Flags().BoolVar(&putOpts.validate, "validate", false,
		"validate key and value before writing")
}

func runPut(cmd *cobra.Command, args []string) error {
	ctx, cancel := getOperationContext()
	defer cancel()

	key := args[0]

	if !strings.HasPrefix(key, "/") {
		return fmt.Errorf("key must start with '/': %s", key)
	}

	value, err := resolveValue(args, os.Stdin)
	if err != nil {
		return err
	}

	if putOpts.validate {
		if err := validateKeyValue(key, value); err != nil {
			return err
		}
		if !isQuietOutput() {
			output.Success("Validation passed")
		}
	}

	etcdClient, cleanup, err := newEtcdClientOrDryRun(putOpts.dryRun)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := etcdClient.Put(ctx, key, value); err != nil {
		return wrapTimeoutError(fmt.Errorf("failed to put key: %w", err))
	}

	if putOpts.dryRun {
		output.Info(fmt.Sprintf("Would put: %s = %s", key, truncateForDisplay(value, 50)))
	} else {
		output.Success(fmt.Sprintf("Put: %s", key))
	}

	return nil
}

func resolveValue(args []string, stdin io.Reader) (string, error) {
	if len(args) < 2 || args[1] == "-" {
		return readValueFromStdin(stdin)
	}
	return args[1], nil
}

func readValueFromStdin(stdin io.Reader) (string, error) {
	if f, ok := stdin.(*os.File); ok {
		stat, err := f.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
			return "", fmt.Errorf("no value provided: use 'etu put <key> <value>' or pipe value via stdin")
		}
	}

	var builder strings.Builder
	scanner := bufio.NewScanner(stdin)
	for scanner.Scan() {
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}

	value := builder.String()
	if value == "" {
		return "", fmt.Errorf("empty value received from stdin")
	}

	return value, nil
}

func validateKeyValue(key, value string) error {
	pair := &models.ConfigPair{Key: key, Value: value}
	v := validator.NewValidator(false)
	result := v.Validate([]*models.ConfigPair{pair})

	if !result.Valid {
		var errMsgs []string
		for _, issue := range result.Issues {
			if issue.Level == "error" {
				errMsgs = append(errMsgs, issue.Message)
			}
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

func truncateForDisplay(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
