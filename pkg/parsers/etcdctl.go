package parsers

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/kazuma-desu/etu/pkg/models"
)

// EtcdctlParser parses etcdctl get output format
// Format:
//
//	/key/path
//	value line 1
//	value line 2
//
//	/another/key
//	single value
type EtcdctlParser struct{}

// FormatName returns the name of this format
func (p *EtcdctlParser) FormatName() string {
	return "etcdctl"
}

// Parse reads and parses an etcdctl format file with cancellation support
func (p *EtcdctlParser) Parse(ctx context.Context, path string) ([]*models.ConfigPair, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pairs []*models.ConfigPair
	var currentKey string
	var currentValueLines []string

	flushCurrent := func() {
		if currentKey != "" {
			value := p.parseValueLines(currentValueLines)
			pairs = append(pairs, &models.ConfigPair{
				Key:   currentKey,
				Value: value,
			})
		}
		currentKey = ""
		currentValueLines = nil
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Empty line
		if trimmed == "" {
			if currentKey != "" {
				currentValueLines = append(currentValueLines, "")
			}
			continue
		}

		// New key (starts with /)
		if strings.HasPrefix(trimmed, "/") {
			flushCurrent()
			currentKey = trimmed
			continue
		}

		// Value line
		if currentKey != "" {
			currentValueLines = append(currentValueLines, line)
		}
	}

	// Flush last pair
	flushCurrent()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return pairs, nil
}

// parseValueLines parses one or more value lines
func (p *EtcdctlParser) parseValueLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// Trim trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return ""
	}

	if len(lines) == 1 {
		return p.parseScalar(lines[0])
	}

	// Join lines as raw multi-line string
	return strings.Join(lines, "\n")
}

// parseScalar parses a scalar value as a string
func (p *EtcdctlParser) parseScalar(s string) string {
	return stripWrappingQuotes(strings.TrimSpace(s))
}

// stripWrappingQuotes removes surrounding quotes from a string
func stripWrappingQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
