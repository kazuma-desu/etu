package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
)

// GetEtcdConfig retrieves etcd configuration from context, flags, or environment variables
// Priority: context (if contextName provided) > environment variables
func GetEtcdConfig() (*client.Config, error) {
	return GetEtcdConfigWithContext("")
}

// GetEtcdConfigWithContext retrieves etcd configuration with optional context override
// Priority: explicit context > current context > environment variables
func GetEtcdConfigWithContext(contextName string) (*client.Config, error) {
	var endpoints []string
	var username, password string

	// Try to load from context first
	if contextName != "" {
		// Use explicit context
		cfg, err := LoadConfig()
		if err == nil && cfg.Contexts[contextName] != nil {
			ctx := cfg.Contexts[contextName]
			endpoints = ctx.Endpoints
			username = ctx.Username
			password = ctx.Password
		}
	} else {
		// Try current context
		ctxConfig, _, err := GetCurrentContext()
		if err == nil && ctxConfig != nil {
			endpoints = ctxConfig.Endpoints
			username = ctxConfig.Username
			password = ctxConfig.Password
		}
	}

	// Fallback to environment variables
	if len(endpoints) == 0 {
		endpoints = getEndpoints()
	}
	if username == "" {
		username = os.Getenv("ETCD_USERNAME")
	}
	if password == "" {
		password = os.Getenv("ETCD_PASSWORD")
	}

	// Parse userpass format (user:pass) for backwards compatibility
	if password == "" {
		if userpass := os.Getenv("ETCD_USERPASS"); userpass != "" {
			parts := strings.SplitN(userpass, ":", 2)
			if len(parts) == 2 {
				username = parts[0]
				password = parts[1]
			}
		}
	}

	// Validate we have at least endpoints
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no etcd configuration found - use 'etu login' or set ETCD_ENDPOINTS environment variable")
	}

	return &client.Config{
		Endpoints:   endpoints,
		Username:    username,
		Password:    password,
		DialTimeout: 5 * time.Second,
	}, nil
}

// getEndpoints parses etcd endpoints from environment
func getEndpoints() []string {
	// Try ETCD_ENDPOINTS first (comma-separated)
	if endpoints := os.Getenv("ETCD_ENDPOINTS"); endpoints != "" {
		return parseEndpoints(endpoints)
	}

	// Fallback to ETCD_HOST for single endpoint (backwards compatibility)
	if host := os.Getenv("ETCD_HOST"); host != "" {
		return []string{host}
	}

	return nil
}

// parseEndpoints splits a comma-separated list of endpoints
func parseEndpoints(s string) []string {
	parts := strings.Split(s, ",")
	var endpoints []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			endpoints = append(endpoints, trimmed)
		}
	}
	return endpoints
}
