package parsers

import (
	"fmt"
	"strings"

	"github.com/kazuma-desu/etu/pkg/models"
)

// UnflattenMap converts a list of ConfigPairs back into a nested map structure.
// It is the reverse operation of FlattenMap.
//
// Rules:
// 1. Keys are split by "/" to create nested structure
// 2. Empty string values are skipped
// 3. Numeric keys are preserved as strings (no array conversion)
// 4. Map values (JSON strings) are preserved as strings
// 5. Collisions between values and directories return an error
func UnflattenMap(pairs []*models.ConfigPair) (map[string]any, error) {
	result := make(map[string]any)

	for _, pair := range pairs {
		parts, shouldProcess := preparePair(pair)
		if !shouldProcess {
			continue
		}

		if err := insertPath(result, parts, pair); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// preparePair handles filtering and path splitting logic.
// It returns the parts of the key and a boolean indicating whether to proceed.
func preparePair(pair *models.ConfigPair) ([]string, bool) {
	if pair == nil {
		return nil, false
	}

	// Skip empty string values
	if strVal, ok := pair.Value.(string); ok && strVal == "" {
		return nil, false
	}

	// Remove leading slash to handle absolute paths
	key := strings.TrimPrefix(pair.Key, "/")
	if key == "" {
		return nil, false
	}

	parts := strings.Split(key, "/")

	// Filter empty parts (handles consecutive slashes like /a//b)
	filtered := parts[:0]
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		return nil, false
	}

	return filtered, true
}

// insertPath navigates the map structure and inserts the value at the leaf.
func insertPath(root map[string]any, parts []string, pair *models.ConfigPair) error {
	current := root
	for i, part := range parts {
		if i == len(parts)-1 {
			return setLeafValue(current, part, pair)
		}

		nextMap, err := navigateToNextMap(current, part, pair.Key, parts[i+1])
		if err != nil {
			return err
		}
		current = nextMap
	}
	return nil
}

func setLeafValue(current map[string]any, part string, pair *models.ConfigPair) error {
	if existing, exists := current[part]; exists {
		if _, isMap := existing.(map[string]any); isMap {
			return fmt.Errorf("key collision: '%s' is implicitly a directory (has children), cannot set as value", pair.Key)
		}
	}
	current[part] = pair.Value
	return nil
}

func navigateToNextMap(current map[string]any, part, originalKey, nextPart string) (map[string]any, error) {
	existing, exists := current[part]

	if !exists {
		newMap := make(map[string]any)
		current[part] = newMap
		return newMap, nil
	}

	if asMap, ok := existing.(map[string]any); ok {
		return asMap, nil
	}

	return nil, fmt.Errorf("key collision: '%s' is already a value, cannot append '%s'",
		originalKey, nextPart)
}
