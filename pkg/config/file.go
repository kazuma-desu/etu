package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var deprecationWarningOnce sync.Once

// ContextConfig represents configuration for a single context
type ContextConfig struct {
	Username              string   `yaml:"username,omitempty"`
	Password              string   `yaml:"password,omitempty"`
	CACert                string   `yaml:"cacert,omitempty"`
	Cert                  string   `yaml:"cert,omitempty"`
	Key                   string   `yaml:"key,omitempty"`
	Endpoints             []string `yaml:"endpoints"`
	InsecureSkipTLSVerify bool     `yaml:"insecure-skip-tls-verify,omitempty"`
}

// Config represents the entire configuration file
type Config struct {
	Contexts       map[string]*ContextConfig `yaml:"contexts"`
	CurrentContext string                    `yaml:"current-context,omitempty"`
	LogLevel       string                    `yaml:"log-level,omitempty"`
	DefaultFormat  string                    `yaml:"default-format,omitempty"`
	Strict         bool                      `yaml:"strict,omitempty"`
	NoValidate     bool                      `yaml:"no-validate,omitempty"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	if envPath := os.Getenv("ETUCONFIG"); envPath != "" {
		return envPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "etu")
	return filepath.Join(configDir, "config.yaml"), nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty config
	info, statErr := os.Stat(configPath)
	if os.IsNotExist(statErr) {
		return &Config{
			Contexts: make(map[string]*ContextConfig),
		}, nil
	}
	if statErr != nil {
		return nil, fmt.Errorf("failed to stat config file %s: %w", configPath, statErr)
	}

	// Check file permissions - warn if too open
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		fmt.Fprintf(os.Stderr, "Warning: Config file %s has permissions %o. Consider changing to 0600 for better security.\n",
			configPath, mode)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure contexts map is initialized
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]*ContextConfig)
	}

	if cfg.DefaultFormat == "etcdctl" || cfg.DefaultFormat == "json" {
		deprecationWarningOnce.Do(func() {
			fmt.Fprintf(os.Stderr, "Warning: '%s' format is deprecated. Consider migrating to YAML using 'etu convert'.\n", cfg.DefaultFormat)
		})
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if mkdirErr := os.MkdirAll(configDir, 0700); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restricted permissions (0600 = rw-------)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCurrentContext returns the current context configuration
func GetCurrentContext() (*ContextConfig, string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, "", err
	}

	if cfg.CurrentContext == "" {
		return nil, "", nil
	}

	ctxConfig, exists := cfg.Contexts[cfg.CurrentContext]
	if !exists {
		return nil, "", fmt.Errorf("current context %q not found", cfg.CurrentContext)
	}

	return ctxConfig, cfg.CurrentContext, nil
}

// SetContext adds or updates a context in the config
func SetContext(name string, ctxConfig *ContextConfig, makeCurrent bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	cfg.Contexts[name] = ctxConfig

	if makeCurrent || cfg.CurrentContext == "" {
		cfg.CurrentContext = name
	}

	return SaveConfig(cfg)
}

// DeleteContext removes a context from the config
func DeleteContext(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if _, exists := cfg.Contexts[name]; !exists {
		return fmt.Errorf("context %q not found", name)
	}

	delete(cfg.Contexts, name)

	// If we deleted the current context, clear it
	if cfg.CurrentContext == name {
		cfg.CurrentContext = ""
		// Optionally set to first available context
		for ctxName := range cfg.Contexts {
			cfg.CurrentContext = ctxName
			break
		}
	}

	return SaveConfig(cfg)
}

// UseContext switches the current context
func UseContext(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if _, exists := cfg.Contexts[name]; !exists {
		return fmt.Errorf("context %q not found", name)
	}

	cfg.CurrentContext = name
	return SaveConfig(cfg)
}
