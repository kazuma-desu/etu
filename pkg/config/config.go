package config

import (
	"fmt"
	"time"

	"github.com/kazuma-desu/etu/pkg/client"
)

// GetEtcdConfig retrieves etcd configuration from context
func GetEtcdConfig() (*client.Config, error) {
	return GetEtcdConfigWithContext("")
}

// GetEtcdConfigWithContext retrieves etcd configuration with optional context override
// Priority: explicit context > current context
func GetEtcdConfigWithContext(contextName string) (*client.Config, error) {
	var endpoints []string
	var username, password string
	var caCert, cert, key string
	var insecureSkipTLSVerify bool

	if contextName != "" {
		cfg, err := LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.Contexts[contextName] == nil {
			return nil, fmt.Errorf("context %q not found in config - use 'etu login' to add it", contextName)
		}
		ctx := cfg.Contexts[contextName]
		endpoints = ctx.Endpoints
		username = ctx.Username
		password = ctx.Password
		caCert = ctx.CACert
		cert = ctx.Cert
		key = ctx.Key
		insecureSkipTLSVerify = ctx.InsecureSkipTLSVerify
	} else {
		ctxConfig, _, err := GetCurrentContext()
		if err != nil {
			return nil, fmt.Errorf("failed to get current context: %w", err)
		}
		if ctxConfig == nil {
			return nil, fmt.Errorf("no current context set - use 'etu login' to configure a context or 'etu config use-context <name>'")
		}
		endpoints = ctxConfig.Endpoints
		username = ctxConfig.Username
		password = ctxConfig.Password
		caCert = ctxConfig.CACert
		cert = ctxConfig.Cert
		key = ctxConfig.Key
		insecureSkipTLSVerify = ctxConfig.InsecureSkipTLSVerify
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no etcd endpoints configured - use 'etu login' to add a context")
	}

	return &client.Config{
		Endpoints:             endpoints,
		Username:              username,
		Password:              password,
		DialTimeout:           5 * time.Second,
		CACert:                caCert,
		Cert:                  cert,
		Key:                   key,
		InsecureSkipTLSVerify: insecureSkipTLSVerify,
	}, nil
}
