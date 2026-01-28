package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEtcdConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "etu")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	t.Setenv("ETUCONFIG", filepath.Join(configDir, "config.yaml"))

	cfg := &Config{
		CurrentContext: "test",
		Contexts: map[string]*ContextConfig{
			"test": {
				Endpoints: []string{"http://localhost:2379"},
				Username:  "admin",
				Password:  "secret",
			},
		},
	}
	require.NoError(t, SaveConfig(cfg))

	etcdCfg, err := GetEtcdConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{"http://localhost:2379"}, etcdCfg.Endpoints)
	assert.Equal(t, "admin", etcdCfg.Username)
	assert.Equal(t, "secret", etcdCfg.Password)
}

func TestGetEtcdConfigWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "etu")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	t.Setenv("ETUCONFIG", filepath.Join(configDir, "config.yaml"))

	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://dev:2379"},
				Username:  "dev-user",
			},
			"prod": {
				Endpoints:             []string{"http://prod:2379"},
				Username:              "prod-user",
				CACert:                "/path/to/ca.crt",
				Cert:                  "/path/to/client.crt",
				Key:                   "/path/to/client.key",
				InsecureSkipTLSVerify: true,
			},
		},
	}
	require.NoError(t, SaveConfig(cfg))

	tests := []struct {
		name           string
		contextName    string
		wantEndpoints  []string
		wantUsername   string
		wantCACert     string
		wantInsecure   bool
		wantErr        bool
		wantErrContain string
	}{
		{
			name:          "uses current context when empty",
			contextName:   "",
			wantEndpoints: []string{"http://dev:2379"},
			wantUsername:  "dev-user",
		},
		{
			name:          "uses specified context",
			contextName:   "prod",
			wantEndpoints: []string{"http://prod:2379"},
			wantUsername:  "prod-user",
			wantCACert:    "/path/to/ca.crt",
			wantInsecure:  true,
		},
		{
			name:           "errors on nonexistent context",
			contextName:    "nonexistent",
			wantErr:        true,
			wantErrContain: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etcdCfg, err := GetEtcdConfigWithContext(tt.contextName)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantEndpoints, etcdCfg.Endpoints)
			assert.Equal(t, tt.wantUsername, etcdCfg.Username)
			assert.Equal(t, tt.wantCACert, etcdCfg.CACert)
			assert.Equal(t, tt.wantInsecure, etcdCfg.InsecureSkipTLSVerify)
		})
	}
}

func TestGetEtcdConfigWithContext_NoCurrentContext(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "etu")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	t.Setenv("ETUCONFIG", filepath.Join(configDir, "config.yaml"))

	cfg := &Config{
		CurrentContext: "",
		Contexts: map[string]*ContextConfig{
			"dev": {
				Endpoints: []string{"http://dev:2379"},
			},
		},
	}
	require.NoError(t, SaveConfig(cfg))

	_, err := GetEtcdConfigWithContext("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no current context set")
}

func TestGetEtcdConfigWithContext_NoEndpoints(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "etu")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	t.Setenv("ETUCONFIG", filepath.Join(configDir, "config.yaml"))

	cfg := &Config{
		CurrentContext: "empty",
		Contexts: map[string]*ContextConfig{
			"empty": {
				Endpoints: []string{},
			},
		},
	}
	require.NoError(t, SaveConfig(cfg))

	_, err := GetEtcdConfigWithContext("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no etcd endpoints configured")
}

func TestGetEtcdConfigWithContext_AllTLSFields(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "etu")
	require.NoError(t, os.MkdirAll(configDir, 0700))

	t.Setenv("ETUCONFIG", filepath.Join(configDir, "config.yaml"))

	cfg := &Config{
		CurrentContext: "tls",
		Contexts: map[string]*ContextConfig{
			"tls": {
				Endpoints:             []string{"https://secure:2379"},
				Username:              "user",
				Password:              "pass",
				CACert:                "/ca.crt",
				Cert:                  "/client.crt",
				Key:                   "/client.key",
				InsecureSkipTLSVerify: false,
			},
		},
	}
	require.NoError(t, SaveConfig(cfg))

	etcdCfg, err := GetEtcdConfigWithContext("")
	require.NoError(t, err)
	assert.Equal(t, []string{"https://secure:2379"}, etcdCfg.Endpoints)
	assert.Equal(t, "user", etcdCfg.Username)
	assert.Equal(t, "pass", etcdCfg.Password)
	assert.Equal(t, "/ca.crt", etcdCfg.CACert)
	assert.Equal(t, "/client.crt", etcdCfg.Cert)
	assert.Equal(t, "/client.key", etcdCfg.Key)
	assert.False(t, etcdCfg.InsecureSkipTLSVerify)
}
