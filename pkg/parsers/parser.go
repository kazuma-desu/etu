package parsers

import (
	"fmt"

	"github.com/kazuma-desu/etu/pkg/models"
)

// Parser defines the interface that all configuration parsers must implement
// This allows for extensibility - new parsers can be added by implementing this interface
type Parser interface {
	// Parse reads a configuration file and returns a slice of ConfigPairs
	Parse(path string) ([]*models.ConfigPair, error)

	// FormatName returns the name of the format this parser handles
	FormatName() string
}

// Registry maintains a mapping of format types to their parsers
type Registry struct {
	parsers map[models.FormatType]Parser
}

// NewRegistry creates a new parser registry with default parsers
func NewRegistry() *Registry {
	r := &Registry{
		parsers: make(map[models.FormatType]Parser),
	}

	// Register built-in parsers
	r.Register(models.FormatEtcdctl, &EtcdctlParser{})

	return r
}

// Register adds a parser to the registry
func (r *Registry) Register(format models.FormatType, parser Parser) {
	r.parsers[format] = parser
}

// GetParser returns the parser for the specified format
func (r *Registry) GetParser(format models.FormatType) (Parser, error) {
	parser, ok := r.parsers[format]
	if !ok {
		return nil, fmt.Errorf("no parser registered for format: %s", format)
	}
	return parser, nil
}

// DetectFormat attempts to detect the file format based on file extension and content
func (r *Registry) DetectFormat(_ string) (models.FormatType, error) {
	// For now, we only support etcdctl format
	// This can be extended with more sophisticated detection logic
	return models.FormatEtcdctl, nil
}
