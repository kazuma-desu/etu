package output

import (
	"fmt"
	"sort"
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

// toNode converts a Go value to a yaml.Node, applying specific formatting rules:
// - Strings with newlines are set to LiteralStyle (block scalar "|")
// - Maps keys are sorted
func toNode(v any) (*yaml.Node, error) {
	switch val := v.(type) {
	case map[string]any:
		return mapToNode(val)
	case []any:
		return sliceToNode(val)
	case string:
		return stringToNode(val), nil
	default:
		return fallbackToNode(val)
	}
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
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}
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

func stringToNode(val string) *yaml.Node {
	// Use default style (0) for single-line scalars - yaml.v3 handles quoting automatically
	// Use LiteralStyle for multi-line strings to render as block scalars (|)
	var style yaml.Style
	if strings.Contains(val, "\n") {
		style = yaml.LiteralStyle
	}
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: val,
		Style: style,
	}
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

	// Should be unreachable for valid YAML
	return &yaml.Node{Kind: yaml.ScalarNode, Value: ""}, nil
}
