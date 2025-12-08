package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [context-name]",
	Short: "Save etcd connection configuration",
	Long: `Save etcd connection configuration for convenient reuse.

The login command allows you to save connection details for one or more etcd clusters.
You can manage multiple contexts (e.g., dev, staging, production) and switch between them.

Examples:
  # Interactive login
  etu login dev

  # Login with authentication
  etu login prod --endpoints http://prod:2379 --username admin --password secret

  # Login without authentication (etcd without auth enabled)
  etu login dev --endpoints http://localhost:2379 --no-auth

  # Login without storing password (more secure)
  etu login prod --endpoints http://prod:2379 --username admin

Security Note:
  Passwords are stored in plain text in ~/.config/etu/config.yaml (like Docker).
  For better security in production/CI environments:
  - Don't store passwords, provide via --password flag at runtime
  - Use environment variables (ETCD_PASSWORD)
  - Restrict file permissions (automatically set to 0600)`,
	Args: cobra.ExactArgs(1),
	Run:  runLogin,
}

var (
	loginEndpoints []string
	loginUsername  string
	loginPassword  string
	loginNoTest    bool
	loginNoAuth    bool
)

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringSliceVar(&loginEndpoints, "endpoints", nil, "Etcd endpoints (comma-separated)")
	loginCmd.Flags().StringVar(&loginUsername, "username", "", "Etcd username")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "Etcd password")
	loginCmd.Flags().BoolVar(&loginNoAuth, "no-auth", false, "Skip authentication (for etcd without auth)")
	loginCmd.Flags().BoolVar(&loginNoTest, "no-test", false, "Skip connection test")
}

func runLogin(_ *cobra.Command, args []string) {
	ctxName := args[0]

	// Get configuration interactively or from flags
	endpoints := loginEndpoints
	username := loginUsername
	password := loginPassword

	reader := bufio.NewReader(os.Stdin)

	// Prompt for endpoints if not provided
	if len(endpoints) == 0 {
		output.Prompt("Enter etcd endpoints (comma-separated): ")
		endpointsStr, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Failed to read input", "error", err)
		}
		endpointsStr = strings.TrimSpace(endpointsStr)
		if endpointsStr == "" {
			log.Fatal("Endpoints are required")
		}
		endpoints = parseEndpointsList(endpointsStr)
	}

	// Prompt for username if not provided (only in interactive mode and auth is needed)
	if !loginNoAuth && username == "" && len(loginEndpoints) == 0 {
		output.Prompt("Enter username (optional, press enter to skip): ")
		usernameInput, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Failed to read input", "error", err)
		}
		username = strings.TrimSpace(usernameInput)
	}

	// Prompt for password if not provided (only in interactive mode and auth is needed)
	if !loginNoAuth && password == "" && username != "" && len(loginEndpoints) == 0 {
		output.Prompt("Enter password (optional, press enter to skip): ")
		passwordInput, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Failed to read input", "error", err)
		}
		password = strings.TrimSpace(passwordInput)
	}

	// Test connection if not skipped
	testPassed := true
	if !loginNoTest {
		output.Info("Testing connection...")
		testPassed = testConnection(endpoints, username, password)

		if !testPassed {
			output.Prompt("Save configuration anyway? (y/N): ")
			response, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal("Failed to read input", "error", err)
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				output.Error("Configuration not saved")
				os.Exit(1)
			}
		}
	}

	// Save configuration
	ctxConfig := &config.ContextConfig{
		Endpoints: endpoints,
		Username:  username,
		Password:  password,
	}

	if err := config.SetContext(ctxName, ctxConfig, true); err != nil {
		log.Fatal("Failed to save configuration", "error", err)
	}

	configPath, _ := config.GetConfigPath()

	if !loginNoTest && testPassed {
		output.Success("Connected successfully")
	}
	output.Success(fmt.Sprintf("Configuration saved to %s", configPath))
	output.Success(fmt.Sprintf("Context '%s' is now active", ctxName))

	// Show security warning if password is stored
	if password != "" {
		output.PrintSecurityWarning()
	}
}

func testConnection(endpoints []string, username, password string) bool {
	cfg := &client.Config{
		Endpoints:   endpoints,
		Username:    username,
		Password:    password,
		DialTimeout: 5 * time.Second,
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to create client: %s", cleanError(err)))
		return false
	}
	defer etcdClient.Close()

	// Try to get cluster status
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = etcdClient.Status(ctx, endpoints[0])
	if err != nil {
		output.Error(fmt.Sprintf("Connection failed: %s", cleanError(err)))
		return false
	}

	return true
}

// cleanError extracts user-friendly error messages from verbose etcd errors
func cleanError(err error) string {
	errStr := err.Error()

	// Common error patterns and their user-friendly versions
	if strings.Contains(errStr, "missing port in address") {
		return "invalid endpoint format (use http://host:port or https://host:port)"
	}
	if strings.Contains(errStr, "connection refused") {
		return "connection refused - is etcd running?"
	}
	if strings.Contains(errStr, "context deadline exceeded") {
		return "connection timeout - check endpoint and network"
	}
	if strings.Contains(errStr, "no such host") {
		return "hostname not found - check endpoint address"
	}
	if strings.Contains(errStr, "authentication failed") {
		return "authentication failed - check username and password"
	}
	if strings.Contains(errStr, "permission denied") {
		return "permission denied - check credentials"
	}

	// Return simplified error without internal etcd details
	if idx := strings.Index(errStr, "rpc error:"); idx != -1 {
		// Extract the description part
		if descIdx := strings.Index(errStr, "desc = "); descIdx != -1 {
			desc := errStr[descIdx+7:]
			// Clean up quotes
			desc = strings.Trim(desc, "\"")
			return desc
		}
	}

	return errStr
}

func parseEndpointsList(s string) []string {
	parts := strings.Split(s, ",")
	var endpoints []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			endpoints = append(endpoints, trimmed)
		}
	}
	return endpoints
}
