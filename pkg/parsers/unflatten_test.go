package parsers

import (
	"encoding/json"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnflattenMap_SimpleKeyValue(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/name", Value: "myapp"},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	assert.Equal(t, "myapp", result["name"])
}

func TestUnflattenMap_NestedMaps(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/database/host", Value: "localhost"},
		{Key: "/app/database/port", Value: 5432},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	app, ok := result["app"].(map[string]any)
	require.True(t, ok)

	db, ok := app["database"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "localhost", db["host"])
	assert.Equal(t, 5432, db["port"])
}

func TestUnflattenMap_Collision_LeafAsParent(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/db", Value: "val"},
		{Key: "/app/db/host", Value: "val2"},
	}

	_, err := UnflattenMap(pairs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key collision")
	assert.Contains(t, err.Error(), "already a value")
}

func TestUnflattenMap_Collision_ParentAsLeaf(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/app/db/host", Value: "val2"},
		{Key: "/app/db", Value: "val"},
	}

	_, err := UnflattenMap(pairs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key collision")
	assert.Contains(t, err.Error(), "implicitly a directory")
}

func TestUnflattenMap_EmptyValueSkipped(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/present", Value: "value"},
		{Key: "/empty", Value: ""},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "value", result["present"])
	_, exists := result["empty"]
	assert.False(t, exists)
}

func TestUnflattenMap_MapValuePreserved(t *testing.T) {
	jsonStr := `{"foo":"bar"}`
	pairs := []*models.ConfigPair{
		{Key: "/config", Value: jsonStr},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	assert.Equal(t, jsonStr, result["config"])
}

func TestUnflattenMap_NumericKeys(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/items/0", Value: "a"},
		{Key: "/items/1", Value: "b"},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	items, ok := result["items"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "a", items["0"])
	assert.Equal(t, "b", items["1"])

	_, isArray := result["items"].([]any)
	assert.False(t, isArray)
}

func TestUnflattenMap_RoundTrip(t *testing.T) {
	input := map[string]any{
		"app": map[string]any{
			"name": "myapp",
			"db": map[string]any{
				"host": "localhost",
				"port": int64(5432),
			},
			"tags": `["a","b"]`,
		},
	}

	pairs := FlattenMap(input)

	output, err := UnflattenMap(pairs)
	require.NoError(t, err)

	inputJSON, _ := json.Marshal(input)
	outputJSON, _ := json.Marshal(output)

	assert.JSONEq(t, string(inputJSON), string(outputJSON))
}

func TestUnflattenMap_MixedTypes_Collision(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/port", Value: 8080},
		{Key: "/port/protocol", Value: "tcp"},
	}

	_, err := UnflattenMap(pairs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already a value")
}

func TestUnflattenMap_NoLeadingSlash(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "app/name", Value: "myapp"},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	app, ok := result["app"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "myapp", app["name"])
}
