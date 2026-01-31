package parsers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser_FormatName(t *testing.T) {
	parser := &JSONParser{}
	assert.Equal(t, "json", parser.FormatName())
}

func TestJSONParser_SimpleKeyValue(t *testing.T) {
	content := `{"name": "myapp"}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/name", "myapp")
}

func TestJSONParser_NestedMaps(t *testing.T) {
	content := `{
	"app": {
		"database": {
			"host": "localhost",
			"port": 5432
		}
	}
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 2)
	assertJSONPair(t, pairs, "/app/database/host", "localhost")
	assertJSONPair(t, pairs, "/app/database/port", int64(5432))
}

func TestJSONParser_DeepNesting(t *testing.T) {
	content := `{
	"level1": {
		"level2": {
			"level3": {
				"level4": {
					"value": "deep"
				}
			}
		}
	}
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/level1/level2/level3/level4/value", "deep")
}

func TestJSONParser_IntegerValue(t *testing.T) {
	content := `{"port": 8080}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/port", int64(8080))
}

func TestJSONParser_FloatValue(t *testing.T) {
	content := `{"rate": 0.95}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/rate", 0.95)
}

func TestJSONParser_WholeNumberFloat(t *testing.T) {
	content := `{"count": 42.0}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/count", int64(42))
}

func TestJSONParser_BooleanValues(t *testing.T) {
	content := `{
	"enabled": true,
	"debug": false
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 2)
	assertJSONPair(t, pairs, "/enabled", true)
	assertJSONPair(t, pairs, "/debug", false)
}

func TestJSONParser_NullValueSkipped(t *testing.T) {
	content := `{
	"present": "value",
	"absent": null
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/present", "value")
}

func TestJSONParser_EmptyStringSkipped(t *testing.T) {
	content := `{
	"present": "value",
	"empty": ""
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/present", "value")
}

func TestJSONParser_ArrayOfScalars(t *testing.T) {
	content := `{
	"tags": ["dev", "test", "prod"]
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/tags", `["dev","test","prod"]`)
}

func TestJSONParser_ArrayOfObjects(t *testing.T) {
	content := `{
	"servers": [
		{"host": "server1", "port": 8080},
		{"host": "server2", "port": 8081}
	]
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/servers", `[{"host":"server1","port":8080},{"host":"server2","port":8081}]`)
}

func TestJSONParser_EmptyArraySkipped(t *testing.T) {
	content := `{"items": []}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 0)
}

func TestJSONParser_MixedTypes(t *testing.T) {
	content := `{
	"app": {
		"name": "myapp",
		"port": 8080,
		"rate": 0.5,
		"enabled": true,
		"tags": ["web", "api"]
	}
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 5)
	assertJSONPair(t, pairs, "/app/name", "myapp")
	assertJSONPair(t, pairs, "/app/port", int64(8080))
	assertJSONPair(t, pairs, "/app/rate", 0.5)
	assertJSONPair(t, pairs, "/app/enabled", true)
	assertJSONPair(t, pairs, "/app/tags", `["web","api"]`)
}

func TestJSONParser_MultipleRoots(t *testing.T) {
	content := `{
	"app": {
		"db": {
			"host": "localhost"
		}
	},
	"custom": {
		"db": {
			"host": "remote"
		}
	}
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 2)
	assertJSONPair(t, pairs, "/app/db/host", "localhost")
	assertJSONPair(t, pairs, "/custom/db/host", "remote")
}

func TestJSONParser_EmptyObject(t *testing.T) {
	content := `{}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 0)
}

func TestJSONParser_EmptyFile(t *testing.T) {
	content := ``

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 0)
}

func TestJSONParser_RootArrayError(t *testing.T) {
	content := `["item1", "item2"]`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &JSONParser{}
	_, err = parser.Parse(context.Background(), tmpFile)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRootNotObject)
}

func TestJSONParser_RootScalarError(t *testing.T) {
	content := `"string value"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &JSONParser{}
	_, err = parser.Parse(context.Background(), tmpFile)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRootNotObject)
}

func TestJSONParser_UnicodeValues(t *testing.T) {
	content := `{
	"greeting": "こんにちは",
	"emoji": "🚀"
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 2)
	assertJSONPair(t, pairs, "/greeting", "こんにちは")
	assertJSONPair(t, pairs, "/emoji", "🚀")
}

func TestJSONParser_SpecialCharactersInKeys(t *testing.T) {
	content := `{
	"key.with.dots": "value1",
	"key-with-dashes": "value2",
	"key_with_underscores": "value3"
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 3)
	assertJSONPair(t, pairs, "/key.with.dots", "value1")
	assertJSONPair(t, pairs, "/key-with-dashes", "value2")
	assertJSONPair(t, pairs, "/key_with_underscores", "value3")
}

func TestJSONParser_FileNotFound(t *testing.T) {
	parser := &JSONParser{}
	_, err := parser.Parse(context.Background(), "/nonexistent/file.json")

	assert.Error(t, err)
}

func TestJSONParser_InvalidJSON(t *testing.T) {
	content := `{invalid: json}`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &JSONParser{}
	_, err = parser.Parse(context.Background(), tmpFile)

	assert.Error(t, err)
}

func TestJSONParser_ScientificNotation(t *testing.T) {
	content := `{
	"small": 1.5e-10,
	"large": 1.5e10
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 2)
	assertJSONPair(t, pairs, "/small", 1.5e-10)
	assertJSONPair(t, pairs, "/large", int64(15000000000))
}

func TestJSONParser_NegativeNumbers(t *testing.T) {
	content := `{
	"negative_int": -42,
	"negative_float": -3.14
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 2)
	assertJSONPair(t, pairs, "/negative_int", int64(-42))
	assertJSONPair(t, pairs, "/negative_float", -3.14)
}

func TestJSONParser_EmptyNestedMap(t *testing.T) {
	content := `{
	"parent": {
		"empty": {},
		"present": "value"
	}
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/parent/present", "value")
}

func TestJSONParser_WhitespaceHandling(t *testing.T) {
	content := `    {    "name"    :    "myapp"    }`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 1)
	assertJSONPair(t, pairs, "/name", "myapp")
}

func TestJSONParser_NestedEmptyMaps(t *testing.T) {
	content := `{
	"a": {
		"b": {
			"c": {}
		}
	}
}`

	pairs := parseJSON(t, content)

	assert.Len(t, pairs, 0)
}

func parseJSON(t *testing.T, content string) []*models.ConfigPair {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &JSONParser{}
	pairs, err := parser.Parse(context.Background(), tmpFile)
	require.NoError(t, err)

	return pairs
}

func assertJSONPair(t *testing.T, pairs []*models.ConfigPair, key string, expectedValue any) {
	t.Helper()
	for _, p := range pairs {
		if p.Key == key {
			assert.Equal(t, expectedValue, p.Value, "Value mismatch for key %s", key)
			return
		}
	}
	t.Errorf("Key %s not found in pairs", key)
}
