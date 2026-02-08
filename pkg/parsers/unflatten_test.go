package parsers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/models"
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
		{Key: "/app/database/port", Value: "5432"},
	}

	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	app, ok := result["app"].(map[string]any)
	require.True(t, ok)

	db, ok := app["database"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "localhost", db["host"])
	assert.Equal(t, "5432", db["port"])
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

	// After flattening and unflattening, all values become strings
	expected := map[string]any{
		"app": map[string]any{
			"name": "myapp",
			"db": map[string]any{
				"host": "localhost",
				"port": "5432",
			},
			"tags": `["a","b"]`,
		},
	}

	expectedJSON, _ := json.Marshal(expected)
	outputJSON, _ := json.Marshal(output)

	assert.JSONEq(t, string(expectedJSON), string(outputJSON))
}

func TestUnflattenMap_MixedTypes_Collision(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/port", Value: "8080"},
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

func TestUnflattenMap_NilEntry(t *testing.T) {
	pairs := []*models.ConfigPair{
		nil,
		{Key: "/valid", Value: "value"},
		nil,
	}
	result, err := UnflattenMap(pairs)
	require.NoError(t, err)
	assert.Equal(t, "value", result["valid"])
}

func TestUnflattenMap_ConsecutiveSlashes(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/a//b/c", Value: "val"},
	}
	result, err := UnflattenMap(pairs)
	require.NoError(t, err)
	// Should produce {"a": {"b": {"c": "val"}}} not {"a": {"": {"b": {"c": "val"}}}}
	a := result["a"].(map[string]any)
	b := a["b"].(map[string]any)
	assert.Equal(t, "val", b["c"])
	_, hasEmpty := a[""]
	assert.False(t, hasEmpty, "should not have empty string key")
}

func TestUnflattenMap_NilValue(t *testing.T) {
	pairs := []*models.ConfigPair{
		{Key: "/key", Value: ""},
	}
	result, err := UnflattenMap(pairs)
	require.NoError(t, err)
	assert.NotContains(t, result, "key")
}

func TestUnflattenMap_RootOnlyKey(t *testing.T) {
	// Key "/" should be skipped (becomes empty after TrimPrefix)
	pairs := []*models.ConfigPair{
		{Key: "/", Value: "value"},
		{Key: "/valid", Value: "ok"},
	}
	result, err := UnflattenMap(pairs)
	require.NoError(t, err)

	// "/" should be skipped, only "valid" should exist
	assert.Len(t, result, 1)
	assert.Equal(t, "ok", result["valid"])
}

func TestUnflattenMap_OnlySlashes(t *testing.T) {
	// Key "///" should be skipped (all parts are empty)
	pairs := []*models.ConfigPair{
		{Key: "///", Value: "value"},
	}
	result, err := UnflattenMap(pairs)
	require.NoError(t, err)
	assert.Empty(t, result)
}
