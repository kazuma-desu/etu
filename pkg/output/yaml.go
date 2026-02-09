package output

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// SerializeYAML serializes data to YAML with automatic block scalar detection for multi-line strings.
func SerializeYAML(data map[string]any) ([]byte, error) {
	node, err := toNode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to YAML node: %w", err)
	}

	return yaml.Marshal(node)
}

func toNode(v any) (*yaml.Node, error) {
	switch val := v.(type) {
	case map[string]any:
		return mapToNode(val)
	case map[string]string:
		return mapStringToNode(val)
	case []any:
		return sliceToNode(val)
	case string:
		return stringToNode(val), nil
	case bool:
		return scalarNode("!!bool", strconv.FormatBool(val)), nil
	case int:
		return scalarNode("!!int", strconv.Itoa(val)), nil
	case float64:
		return scalarNode("!!float", strconv.FormatFloat(val, 'f', -1, 64)), nil
	case nil:
		return scalarNode("!!null", "null"), nil
	default:
		return fallbackToNode(val)
	}
}

func scalarNode(tag, value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
}

func mapStringToNode(val map[string]string) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}

	keys := make([]string, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}
		valNode := stringToNode(val[k])
		node.Content = append(node.Content, keyNode, valNode)
	}
	return node, nil
}

func mapToNode(val map[string]any) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}
		valNode, err := toNode(val[k])
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valNode)
	}
	return node, nil
}

func sliceToNode(val []any) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.SequenceNode}
	for _, item := range val {
		itemNode, err := toNode(item)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, itemNode)
	}
	return node, nil
}

var (
	intRe   = regexp.MustCompile(`^-?\d+$`)
	floatRe = regexp.MustCompile(`^-?\d+\.\d+([eE][+-]?\d+)?$|^-?\d+[eE][+-]?\d+$`)
)

func stringToNode(val string) *yaml.Node {
	// Heuristic: detect if the string looks like a number or bool
	// and emit it with the appropriate tag so YAML renders it unquoted

	// Special case: lowercase "true"/"false" render as !!bool (cosmetic preference)
	if val == "true" || val == "false" {
		return scalarNode("!!bool", val)
	}

	// Check for YAML special values that should be quoted to prevent misinterpretation
	// YAML 1.1 treats these case-insensitively, so normalize before comparison
	lower := strings.ToLower(val)
	switch lower {
	case "null", "~", "yes", "no", "on", "off", "true", "false":
		// Force quoted with !!str tag to prevent YAML parser from interpreting as special values
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: val,
			Style: yaml.DoubleQuotedStyle,
		}
	}

	// Integer-looking strings: render as !!int (unquoted)
	// Avoid emitting leading-zero numbers as !!int (YAML 1.1 octal ambiguity)
	if intRe.MatchString(val) {
		stripped := strings.TrimPrefix(val, "-")
		if !(len(stripped) > 1 && stripped[0] == '0') {
			return scalarNode("!!int", val)
		}
	}

	// Float-looking strings: render as !!float (unquoted)
	if floatRe.MatchString(val) {
		return scalarNode("!!float", val)
	}

	// Regular strings: use !!str tag (default quoted behavior)
	return scalarNode("!!str", val)
}

func fallbackToNode(val any) (*yaml.Node, error) {
	// Fallback to standard marshaling for other types (int, bool, nil, custom types)
	data, err := yaml.Marshal(val)
	if err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	// doc is a DocumentNode, the first child is the actual value
	if len(doc.Content) > 0 {
		return doc.Content[0], nil
	}

	return nil, fmt.Errorf("yaml.Unmarshal produced a document with no content nodes")
}
