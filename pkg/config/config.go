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
		if err == nil && cfg.Contexts[contextName] != nil {
			ctx := cfg.Contexts[contextName]
			endpoints = ctx.Endpoints
			username = ctx.Username
			password = ctx.Password
			caCert = ctx.CACert
			cert = ctx.Cert
			key = ctx.Key
			insecureSkipTLSVerify = ctx.InsecureSkipTLSVerify
		}
	} else {
		ctxConfig, _, err := GetCurrentContext()
		if err == nil && ctxConfig != nil {
			endpoints = ctxConfig.Endpoints
			username = ctxConfig.Username
			password = ctxConfig.Password
			caCert = ctxConfig.CACert
			cert = ctxConfig.Cert
			key = ctxConfig.Key
			insecureSkipTLSVerify = ctxConfig.InsecureSkipTLSVerify
		}
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no etcd configuration found - use 'etu login' to configure a context")
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
