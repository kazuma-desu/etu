package parsers

import (
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestFlattenMap_SimpleKeyValue(t *testing.T) {
	input := map[string]any{
		"name": "myapp",
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/name", "myapp")
}

func TestFlattenMap_NestedMaps(t *testing.T) {
	input := map[string]any{
		"app": map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 2)
	assertPair(t, pairs, "/app/database/host", "localhost")
	assertPair(t, pairs, "/app/database/port", int64(5432))
}

func TestFlattenMap_DeepNesting(t *testing.T) {
	input := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": "deep",
					},
				},
			},
		},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/level1/level2/level3/level4/value", "deep")
}

func TestFlattenMap_IntegerValue(t *testing.T) {
	input := map[string]any{
		"port": 8080,
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/port", int64(8080))
}

func TestFlattenMap_FloatValue(t *testing.T) {
	input := map[string]any{
		"rate": 0.95,
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/rate", 0.95)
}

func TestFlattenMap_WholeNumberFloat(t *testing.T) {
	input := map[string]any{
		"count": 42.0,
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/count", int64(42))
}

func TestFlattenMap_BooleanValue(t *testing.T) {
	input := map[string]any{
		"enabled": true,
		"debug":   false,
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 2)
	assertPair(t, pairs, "/enabled", true)
	assertPair(t, pairs, "/debug", false)
}

func TestFlattenMap_NullValueSkipped(t *testing.T) {
	input := map[string]any{
		"present": "value",
		"absent":  nil,
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/present", "value")
}

func TestFlattenMap_EmptyStringSkipped(t *testing.T) {
	input := map[string]any{
		"present": "value",
		"empty":   "",
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/present", "value")
}

func TestFlattenMap_ArrayOfScalars(t *testing.T) {
	input := map[string]any{
		"tags": []any{"dev", "test", "prod"},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/tags", `["dev","test","prod"]`)
}

func TestFlattenMap_ArrayOfObjects(t *testing.T) {
	input := map[string]any{
		"servers": []any{
			map[string]any{"host": "server1", "port": 8080},
			map[string]any{"host": "server2", "port": 8081},
		},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/servers", `[{"host":"server1","port":8080},{"host":"server2","port":8081}]`)
}

func TestFlattenMap_EmptyArraySkipped(t *testing.T) {
	input := map[string]any{
		"items": []any{},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 0)
}

func TestFlattenMap_MixedTypes(t *testing.T) {
	input := map[string]any{
		"app": map[string]any{
			"name":    "myapp",
			"port":    8080,
			"rate":    0.5,
			"enabled": true,
			"tags":    []any{"a", "b"},
		},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 5)
	assertPair(t, pairs, "/app/name", "myapp")
	assertPair(t, pairs, "/app/port", int64(8080))
	assertPair(t, pairs, "/app/rate", 0.5)
	assertPair(t, pairs, "/app/enabled", true)
	assertPair(t, pairs, "/app/tags", `["a","b"]`)
}

func TestFlattenMap_MultipleRoots(t *testing.T) {
	input := map[string]any{
		"app": map[string]any{
			"db": map[string]any{
				"host": "localhost",
			},
		},
		"custom": map[string]any{
			"db": map[string]any{
				"host": "remote",
			},
		},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 2)
	assertPair(t, pairs, "/app/db/host", "localhost")
	assertPair(t, pairs, "/custom/db/host", "remote")
}

func TestFlattenMap_EmptyMap(t *testing.T) {
	input := map[string]any{}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 0)
}

func TestFlattenMap_NestedEmptyMap(t *testing.T) {
	input := map[string]any{
		"empty": map[string]any{},
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 0)
}

func TestFlattenMap_Int64Value(t *testing.T) {
	input := map[string]any{
		"bignum": int64(9223372036854775807),
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 1)
	assertPair(t, pairs, "/bignum", int64(9223372036854775807))
}

func TestFlattenMap_SpecialCharactersInKeys(t *testing.T) {
	input := map[string]any{
		"key.with.dots":   "value1",
		"key-with-dashes": "value2",
		"key_with_under":  "value3",
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 3)
	assertPair(t, pairs, "/key.with.dots", "value1")
	assertPair(t, pairs, "/key-with-dashes", "value2")
	assertPair(t, pairs, "/key_with_under", "value3")
}

func TestFlattenMap_UnicodeValues(t *testing.T) {
	input := map[string]any{
		"greeting": "こんにちは",
		"emoji":    "🚀",
	}

	pairs := FlattenMap(input)

	assert.Len(t, pairs, 2)
	assertPair(t, pairs, "/greeting", "こんにちは")
	assertPair(t, pairs, "/emoji", "🚀")
}

func assertPair(t *testing.T, pairs []*models.ConfigPair, key string, expectedValue any) {
	t.Helper()
	for _, p := range pairs {
		if p.Key == key {
			assert.Equal(t, expectedValue, p.Value, "Value mismatch for key %s", key)
			return
		}
	}
	t.Errorf("Key %s not found in pairs", key)
}
