package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffKeyValues_AllAdded(t *testing.T) {
	fileMap := map[string]string{
		"/app/config/key1": "value1",
		"/app/config/key2": "value2",
	}
	etcdMap := map[string]string{}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Equal(t, 2, result.Added)
	assert.Equal(t, 0, result.Modified)
	assert.Equal(t, 0, result.Deleted)
	assert.Equal(t, 0, result.Unchanged)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, DiffStatusAdded, result.Entries[0].Status)
	assert.Equal(t, DiffStatusAdded, result.Entries[1].Status)
}

func TestDiffKeyValues_AllDeleted(t *testing.T) {
	fileMap := map[string]string{}
	etcdMap := map[string]string{
		"/app/config/key1": "value1",
		"/app/config/key2": "value2",
	}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 0, result.Modified)
	assert.Equal(t, 2, result.Deleted)
	assert.Equal(t, 0, result.Unchanged)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, DiffStatusDeleted, result.Entries[0].Status)
	assert.Equal(t, DiffStatusDeleted, result.Entries[1].Status)
}

func TestDiffKeyValues_AllUnchanged(t *testing.T) {
	fileMap := map[string]string{
		"/app/config/key1": "value1",
		"/app/config/key2": "value2",
	}
	etcdMap := map[string]string{
		"/app/config/key1": "value1",
		"/app/config/key2": "value2",
	}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 0, result.Modified)
	assert.Equal(t, 0, result.Deleted)
	assert.Equal(t, 2, result.Unchanged)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, DiffStatusUnchanged, result.Entries[0].Status)
	assert.Equal(t, DiffStatusUnchanged, result.Entries[1].Status)
}

func TestDiffKeyValues_AllModified(t *testing.T) {
	fileMap := map[string]string{
		"/app/config/key1": "new_value1",
		"/app/config/key2": "new_value2",
	}
	etcdMap := map[string]string{
		"/app/config/key1": "old_value1",
		"/app/config/key2": "old_value2",
	}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 2, result.Modified)
	assert.Equal(t, 0, result.Deleted)
	assert.Equal(t, 0, result.Unchanged)
	assert.Len(t, result.Entries, 2)
	assert.Equal(t, DiffStatusModified, result.Entries[0].Status)
	assert.Equal(t, "old_value1", result.Entries[0].OldValue)
	assert.Equal(t, "new_value1", result.Entries[0].NewValue)
}

func TestDiffKeyValues_Mixed(t *testing.T) {
	fileMap := map[string]string{
		"/app/config/new_key":  "new_value",
		"/app/config/existing": "updated_value",
	}
	etcdMap := map[string]string{
		"/app/config/existing":  "old_value",
		"/app/config/to_delete": "deleted_value",
	}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Modified)
	assert.Equal(t, 1, result.Deleted)
	assert.Equal(t, 0, result.Unchanged)
	assert.Len(t, result.Entries, 3)

	// Find each entry
	var added, modified, deleted *DiffEntry
	for _, e := range result.Entries {
		switch e.Status {
		case DiffStatusAdded:
			added = e
		case DiffStatusModified:
			modified = e
		case DiffStatusDeleted:
			deleted = e
		}
	}

	assert.NotNil(t, added)
	assert.Equal(t, "/app/config/new_key", added.Key)
	assert.Equal(t, "new_value", added.NewValue)

	assert.NotNil(t, modified)
	assert.Equal(t, "/app/config/existing", modified.Key)
	assert.Equal(t, "old_value", modified.OldValue)
	assert.Equal(t, "updated_value", modified.NewValue)

	assert.NotNil(t, deleted)
	assert.Equal(t, "/app/config/to_delete", deleted.Key)
	assert.Equal(t, "deleted_value", deleted.OldValue)
}

func TestDiffKeyValues_EmptyMaps(t *testing.T) {
	fileMap := map[string]string{}
	etcdMap := map[string]string{}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Equal(t, 0, result.Added)
	assert.Equal(t, 0, result.Modified)
	assert.Equal(t, 0, result.Deleted)
	assert.Equal(t, 0, result.Unchanged)
	assert.Len(t, result.Entries, 0)
}

func TestDiffKeyValues_KeysSorted(t *testing.T) {
	fileMap := map[string]string{
		"/z/last":   "value",
		"/a/first":  "value",
		"/m/middle": "value",
	}
	etcdMap := map[string]string{}

	result := DiffKeyValues(fileMap, etcdMap)

	assert.Len(t, result.Entries, 3)
	assert.Equal(t, "/a/first", result.Entries[0].Key)
	assert.Equal(t, "/m/middle", result.Entries[1].Key)
	assert.Equal(t, "/z/last", result.Entries[2].Key)
}

func TestDiffEntry_JSON(t *testing.T) {
	entry := &DiffEntry{
		Key:      "/app/config/test",
		Status:   DiffStatusModified,
		OldValue: "old",
		NewValue: "new",
	}

	data, err := json.Marshal(entry)
	assert.NoError(t, err)

	var decoded DiffEntry
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, entry.Key, decoded.Key)
	assert.Equal(t, entry.Status, decoded.Status)
	assert.Equal(t, entry.OldValue, decoded.OldValue)
	assert.Equal(t, entry.NewValue, decoded.NewValue)
}

func TestDiffStatus_Constants(t *testing.T) {
	assert.Equal(t, DiffStatus("added"), DiffStatusAdded)
	assert.Equal(t, DiffStatus("modified"), DiffStatusModified)
	assert.Equal(t, DiffStatus("deleted"), DiffStatusDeleted)
	assert.Equal(t, DiffStatus("unchanged"), DiffStatusUnchanged)
}
