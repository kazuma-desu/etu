package parsers

import (
	"encoding/json"

	"github.com/kazuma-desu/etu/pkg/logger"
	"github.com/kazuma-desu/etu/pkg/models"
)

// FlattenMap recursively flattens a nested map into etcd key-value pairs.
// Keys are constructed as paths with "/" delimiter (e.g., /app/db/host).
// Arrays are serialized as JSON strings.
// Null values are skipped.
func FlattenMap(data map[string]any) []*models.ConfigPair {
	var pairs []*models.ConfigPair
	flattenRecursive("", data, &pairs)
	return pairs
}

func flattenRecursive(prefix string, data map[string]any, pairs *[]*models.ConfigPair) {
	for key, value := range data {
		fullKey := prefix + "/" + key
		flattenValue(fullKey, value, pairs)
	}
}

func flattenValue(key string, value any, pairs *[]*models.ConfigPair) {
	if value == nil {
		return
	}

	switch v := value.(type) {
	case map[string]any:
		flattenRecursive(key, v, pairs)

	case []any:
		if len(v) == 0 {
			return
		}
		serialized, err := json.Marshal(v)
		if err != nil {
			logger.Log.Warn("failed to marshal array", "key", key, "error", err)
			return
		}
		*pairs = append(*pairs, &models.ConfigPair{
			Key:   key,
			Value: string(serialized),
		})

	default:
		formatted := models.FormatValue(v)
		if formatted == "" {
			return
		}
		*pairs = append(*pairs, &models.ConfigPair{
			Key:   key,
			Value: formatted,
		})
	}
}
