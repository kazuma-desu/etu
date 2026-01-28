package cmd

import (
	"github.com/spf13/cobra"
)

var optionsCmd = &cobra.Command{
	Use:   "options",
	Short: "Print the list of global flags inherited by all commands",
	Long:  `Print the list of global command-line options (flags) that can be passed to any command.`,
	Run:   runOptions,
}

func init() {
	rootCmd.AddCommand(optionsCmd)
}

func runOptions(cmd *cobra.Command, _ []string) {
	cmd.Print(`The following options can be passed to any command:

    --cacert='':
        Path to CA certificate for server verification

    --cert='':
        Path to client certificate for TLS

    --context='':
        The name of the context to use (overrides current-context)

    --insecure-skip-tls-verify=false:
        If true, the server's certificate will not be checked for validity.
        This will make your HTTPS connections insecure

    --key='':
        Path to client key for TLS

    --log-level='':
        Log level (debug, info, warn, error) - overrides config file

    -o, --output='simple':
        Output format (simple, json, table, tree)

    --password='':
        Password for etcd authentication (overrides context)

    --password-stdin=false:
        Read password from stdin (mutually exclusive with --password)

    --timeout=30s:
        Timeout for etcd operations (e.g., 30s, 1m, 2m30s)

    --username='':
        Username for etcd authentication (overrides context)
`)
}
