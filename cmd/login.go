package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Add a new etcd cluster context",
	Long: `Interactive wizard to configure a new etcd cluster connection.

Configuration is saved to ~/.config/etu/config.yaml.

For automation, use flags:
  etu login --context-name prod --endpoints http://etcd:2379`,
	Example: `  etu login
  etu login --context-name prod --endpoints http://etcd:2379 --username admin --password secret`,
	Args: cobra.NoArgs,
	RunE: runLogin,
}

var (
	loginContextName string
	loginEndpoints   []string
	loginUsername    string
	loginPassword    string
	loginNoTest      bool
	loginNoAuth      bool
)

type loginForm struct {
	ContextName  string
	Endpoints    string
	Username     string
	Password     string
	RequiresAuth bool
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginContextName, "context-name", "", "Context name")
	loginCmd.Flags().StringSliceVar(&loginEndpoints, "endpoints", nil, "Etcd endpoints (comma-separated)")
	loginCmd.Flags().StringVar(&loginUsername, "username", "", "Etcd username")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "Etcd password")
	loginCmd.Flags().BoolVar(&loginNoAuth, "no-auth", false, "Skip authentication")
	loginCmd.Flags().BoolVar(&loginNoTest, "no-test", false, "Skip connection test")
}

func runLogin(_ *cobra.Command, _ []string) error {
	if hasLoginFlags() {
		return runLoginAutomated()
	}
	return runLoginInteractive()
}

func hasLoginFlags() bool {
	return loginContextName != "" || len(loginEndpoints) > 0 ||
		loginUsername != "" || loginPassword != "" || loginNoAuth || loginNoTest
}

func runLoginInteractive() error {
	form := &loginForm{}
	accessible := os.Getenv("ACCESSIBLE") != ""

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Context name").
				Description("A unique name to identify this cluster").
				Placeholder("dev | apisix-stg | k8s-etcd").
				Validate(validateContextName).
				Value(&form.ContextName),

			huh.NewInput().
				Title("Endpoints").
				Description("etcd server addresses (comma-separated)").
				Placeholder("http://localhost:2379").
				Validate(validateEndpoints).
				Value(&form.Endpoints),

			huh.NewConfirm().
				Title("Requires authentication?").
				Description("Enable if your cluster uses username/password").
				Affirmative("Yes").
				Negative("No").
				Value(&form.RequiresAuth),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Description("etcd authentication username").
				Placeholder("root").
				Value(&form.Username),

			huh.NewInput().
				Title("Password").
				Description("etcd authentication password").
				EchoMode(huh.EchoModePassword).
				Value(&form.Password),
		).WithHideFunc(func() bool { return !form.RequiresAuth }),
	).
		WithTheme(huh.ThemeCharm()).
		WithAccessible(accessible).
		Run()

	if err != nil {
		if err == huh.ErrUserAborted {
			return nil
		}
		return err
	}

	endpoints := parseEndpoints(form.Endpoints)
	username, password := "", ""
	if form.RequiresAuth {
		username = strings.TrimSpace(form.Username)
		password = form.Password
	}

	testPassed := testConnection(endpoints, username, password)

	if !testPassed {
		fmt.Print("Connection failed. Save anyway? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	ctxConfig := &config.ContextConfig{
		Endpoints: endpoints,
		Username:  username,
		Password:  password,
	}

	if err := config.SetContext(form.ContextName, ctxConfig, true); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	configPath, _ := config.GetConfigPath()

	if testPassed {
		output.Success("Connection verified")
	}
	output.Success(fmt.Sprintf("Saved to %s", configPath))
	output.Success(fmt.Sprintf("Context '%s' is now active", form.ContextName))

	return nil
}

func runLoginAutomated() error {
	ctxName := strings.TrimSpace(loginContextName)
	if ctxName == "" {
		return fmt.Errorf("--context-name is required")
	}

	if err := validateContextNameFormat(ctxName); err != nil {
		return fmt.Errorf("invalid context name: %w", err)
	}

	endpoints := loginEndpoints
	if len(endpoints) == 0 {
		return fmt.Errorf("--endpoints is required")
	}

	// Normalize endpoints: trim whitespace and filter out empty strings
	endpointsNormalized := make([]string, 0, len(endpoints))
	for _, ep := range endpoints {
		trimmed := strings.TrimSpace(ep)
		if trimmed != "" {
			endpointsNormalized = append(endpointsNormalized, trimmed)
		}
	}

	if len(endpointsNormalized) == 0 {
		return fmt.Errorf("--endpoints cannot be empty")
	}

	for _, ep := range endpointsNormalized {
		if err := validateEndpointFormat(ep); err != nil {
			return fmt.Errorf("invalid endpoint: %w", err)
		}
	}

	username, password := loginUsername, loginPassword
	if loginNoAuth {
		username, password = "", ""
	}

	if !loginNoTest {
		output.Info("Testing connection...")
		if !testConnectionQuiet(endpointsNormalized, username, password) {
			return fmt.Errorf("connection failed - use --no-test to skip")
		}
	}

	ctxConfig := &config.ContextConfig{
		Endpoints: endpointsNormalized,
		Username:  username,
		Password:  password,
	}

	if err := config.SetContext(ctxName, ctxConfig, true); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	output.Success(fmt.Sprintf("Saved to %s", configPath))
	output.Success(fmt.Sprintf("Context '%s' is now active", ctxName))

	return nil
}

func testConnection(endpoints []string, username, password string) bool {
	cfg := &client.Config{
		Endpoints:   endpoints,
		Username:    username,
		Password:    password,
		DialTimeout: 5 * time.Second,
	}

	done := make(chan bool, 1)

	go func() {
		etcdClient, err := client.NewClient(cfg)
		if err != nil {
			done <- false
			return
		}
		defer etcdClient.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = etcdClient.Status(ctx, endpoints[0])
		done <- err == nil
	}()

	var result bool
	_ = spinner.New().
		Title("Testing connection...").
		Action(func() {
			result = <-done
		}).
		Run()

	return result
}

func testConnectionQuiet(endpoints []string, username, password string) bool {
	cfg := &client.Config{
		Endpoints:   endpoints,
		Username:    username,
		Password:    password,
		DialTimeout: 5 * time.Second,
	}

	etcdClient, err := client.NewClient(cfg)
	if err != nil {
		return false
	}
	defer etcdClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = etcdClient.Status(ctx, endpoints[0])
	return err == nil
}

func validateContextNameFormat(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("enter a context name")
	}
	if len(s) < 2 {
		return fmt.Errorf("at least 2 characters")
	}
	if len(s) > 63 {
		return fmt.Errorf("max 63 characters")
	}
	if strings.Contains(s, " ") {
		return fmt.Errorf("spaces not allowed, use dashes")
	}
	for _, r := range s {
		isLower := r >= 'a' && r <= 'z'
		isUpper := r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		isSpecial := r == '-' || r == '_'
		if !isLower && !isUpper && !isDigit && !isSpecial {
			return fmt.Errorf("invalid character '%c' — use letters, numbers, dash, underscore", r)
		}
	}
	return nil
}

func validateContextName(s string) error {
	if err := validateContextNameFormat(s); err != nil {
		return err
	}
	cfg, err := config.LoadConfig()
	if err == nil && cfg.Contexts[strings.TrimSpace(s)] != nil {
		return fmt.Errorf("context '%s' already exists", strings.TrimSpace(s))
	}
	return nil
}

func validateEndpointFormat(endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return fmt.Errorf("empty endpoint")
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		return fmt.Errorf("'%s' — must start with http:// or https://", truncate(endpoint, 20))
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("'%s' — invalid URL", truncate(endpoint, 20))
	}

	if parsed.Host == "" {
		return fmt.Errorf("'%s' — missing hostname", truncate(endpoint, 20))
	}

	return nil
}

func validateEndpoints(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("enter at least one endpoint")
	}

	validCount := 0
	for _, endpoint := range strings.Split(s, ",") {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint == "" {
			continue
		}

		validCount++

		if err := validateEndpointFormat(endpoint); err != nil {
			return err
		}
	}

	if validCount == 0 {
		return fmt.Errorf("enter at least one endpoint")
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func parseEndpoints(s string) []string {
	var endpoints []string
	for _, p := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			endpoints = append(endpoints, trimmed)
		}
	}
	return endpoints
}
