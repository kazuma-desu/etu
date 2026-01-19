package parsers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	r.Register(models.FormatEtcdctl, &EtcdctlParser{})
	r.Register(models.FormatYAML, &YAMLParser{})
	r.Register(models.FormatJSON, &JSONParser{})

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

// DetectFormat detects file format from extension, falling back to content analysis.
// Priority: extension > content signature > default (etcdctl)
func (r *Registry) DetectFormat(path string) (models.FormatType, error) {
	// 1. Extension-based detection (fastest)
	format := r.detectByExtension(path)
	if format != models.FormatAuto {
		if _, err := r.GetParser(format); err == nil {
			return format, nil
		}
	}

	// 2. Content-based detection (for extensionless files or unregistered parsers)
	format = r.detectByContent(path)
	if format != models.FormatAuto {
		if _, err := r.GetParser(format); err == nil {
			return format, nil
		}
	}

	// 3. Default fallback
	return models.FormatEtcdctl, nil
}

// detectByExtension returns format based on file extension
func (r *Registry) detectByExtension(path string) models.FormatType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return models.FormatYAML
	case ".json":
		return models.FormatJSON
	case ".txt":
		return models.FormatEtcdctl
	default:
		return models.FormatAuto
	}
}

// detectByContent peeks at file content to detect format
func (r *Registry) detectByContent(path string) models.FormatType {
	file, err := os.Open(path)
	if err != nil {
		return models.FormatAuto
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle long lines (default 64K may be too small)
	// Use 1MB initial capacity, 10MB max token size
	const (
		initBufSize  = 1024 * 1024      // 1MB
		maxTokenSize = 10 * 1024 * 1024 // 10MB
	)
	scanner.Buffer(make([]byte, initBufSize), maxTokenSize)
	var firstNonEmptyLine string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			firstNonEmptyLine = line
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return models.FormatAuto
	}

	if firstNonEmptyLine == "" {
		return models.FormatEtcdctl
	}

	// JSON detection: starts with { or [
	if strings.HasPrefix(firstNonEmptyLine, "{") || strings.HasPrefix(firstNonEmptyLine, "[") {
		return models.FormatJSON
	}

	// YAML detection: document separator or key: value pattern (not starting with /)
	if strings.HasPrefix(firstNonEmptyLine, "---") {
		return models.FormatYAML
	}
	if strings.Contains(firstNonEmptyLine, ": ") && !strings.HasPrefix(firstNonEmptyLine, "/") {
		return models.FormatYAML
	}

	// etcdctl detection: starts with /
	if strings.HasPrefix(firstNonEmptyLine, "/") {
		return models.FormatEtcdctl
	}

	return models.FormatAuto
}
