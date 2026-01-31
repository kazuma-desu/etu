package parsers

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kazuma-desu/etu/pkg/models"
)

var langLineRE = regexp.MustCompile(`^(?P<tag>[A-Za-z0-9_\-]{2,}):\s*(?P<val>.+?)\s*$`)

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
func (p *EtcdctlParser) parseValueLines(lines []string) any {
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

	// Check if all lines match the language tag pattern (key: value)
	langMap := make(map[string]any)
	allLangish := true

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			allLangish = false
			break
		}

		matches := langLineRE.FindStringSubmatch(trimmed)
		if matches == nil {
			allLangish = false
			break
		}

		tag := matches[1]
		val := stripWrappingQuotes(strings.TrimSpace(matches[2]))
		langMap[tag] = val
	}

	if allLangish {
		return langMap
	}

	// Fallback: join lines as single string
	return strings.Join(lines, "\n")
}

// parseScalar attempts to parse a scalar value with type inference
func (p *EtcdctlParser) parseScalar(s string) any {
	s = strings.TrimSpace(s)

	// Try int
	if matched, _ := regexp.MatchString(`^[+-]?\d+$`, s); matched {
		if val, err := strconv.ParseInt(s, 10, 64); err == nil {
			return val
		}
	}

	// Try float
	if matched, _ := regexp.MatchString(`^[+-]?(?:\d+\.\d*|\d*\.\d+|\d+)$`, s); matched {
		if val, err := strconv.ParseFloat(s, 64); err == nil {
			return val
		}
	}

	// Return as string, stripping quotes if present
	return stripWrappingQuotes(s)
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
