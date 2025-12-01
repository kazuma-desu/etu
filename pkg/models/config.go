package models

import "fmt"

// ConfigPair represents a single etcd key-value configuration pair
type ConfigPair struct {
	Key   string
	Value any
}

// String returns a string representation of the config pair
func (c *ConfigPair) String() string {
	return fmt.Sprintf("%s: %v", c.Key, c.Value)
}

// FormatType represents the type of configuration file format
type FormatType string

const (
	FormatAuto    FormatType = "auto"
	FormatEtcdctl FormatType = "etcdctl"
	// FormatHelmValues FormatType = "helm-values" // Reserved for future use
)

// IsValid checks if the format type is valid
func (f FormatType) IsValid() bool {
	switch f {
	case FormatAuto, FormatEtcdctl:
		return true
	default:
		return false
	}
}

// ApplyOptions contains options for applying configuration
type ApplyOptions struct {
	FilePath   string
	Format     FormatType
	DryRun     bool
	NoValidate bool
	Strict     bool
}

// ValidateOptions contains options for validation
type ValidateOptions struct {
	FilePath string
	Format   FormatType
	Strict   bool
}

// ParseOptions contains options for parsing
type ParseOptions struct {
	FilePath   string
	Format     FormatType
	JSONOutput bool
}
