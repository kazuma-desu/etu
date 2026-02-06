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
		// Skip empty string values
		if strVal, ok := pair.Value.(string); ok && strVal == "" {
			continue
		}

		// Remove leading slash to handle absolute paths
		key := strings.TrimPrefix(pair.Key, "/")
		if key == "" {
			continue
		}

		parts := strings.Split(key, "/")
		current := result

		for i, part := range parts {
			isLeaf := i == len(parts)-1

			if isLeaf {
				// Leaf node: Attempt to set the value
				if existing, exists := current[part]; exists {
					// Collision check: Cannot overwrite a map (directory) with a value
					if _, isMap := existing.(map[string]any); isMap {
						return nil, fmt.Errorf("key collision: '%s' is implicitly a directory (has children), cannot set as value", pair.Key)
					}
					// Overwriting existing value is allowed (last write wins)
				}
				current[part] = pair.Value
			} else {
				// Intermediate node: Navigate or create map
				existing, exists := current[part]

				if !exists {
					// Create new level
					newMap := make(map[string]any)
					current[part] = newMap
					current = newMap
				} else {
					// Verify existing node is a map
					if asMap, ok := existing.(map[string]any); ok {
						current = asMap
					} else {
						// Collision check: Cannot treat a value as a directory
						// Construct the conflicting path for the error message
						conflictingPath := "/" + strings.Join(parts[:i+1], "/")
						return nil, fmt.Errorf("key collision: '%s' is already a value, cannot append '%s'", conflictingPath, parts[i+1])
					}
				}
			}
		}
	}

	return result, nil
}
