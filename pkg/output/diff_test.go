package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

func captureStdout(t *testing.T, f func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("captureStdout: failed to create pipe: %v", pipeErr)
	}

	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		outCh <- buf.String()
	}()

	os.Stdout = w

	var fErr error
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				fErr = fmt.Errorf("captureStdout: f() panicked: %v", rec)
			}
		}()
		fErr = f()
	}()

	w.Close()
	os.Stdout = old
	output := <-outCh

	return output, fErr
}

func TestPrintDiffResult_UnsupportedFormat(t *testing.T) {
	result := &DiffResult{Entries: []*DiffEntry{}}
	err := PrintDiffResult(result, "invalid_format", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
	assert.Contains(t, err.Error(), "yaml")
}

func TestPrintDiffResult_DispatchesToSimple(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/test/key", Status: DiffStatusAdded, NewValue: "value"},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return PrintDiffResult(result, "simple", false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Added")
	assert.Contains(t, output, "/test/key")
}

func TestPrintDiffResult_DispatchesToJSON(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/test/key", Status: DiffStatusAdded, NewValue: "value"},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return PrintDiffResult(result, "json", false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, `"added"`)
	assert.Contains(t, output, `"/test/key"`)
}

func TestPrintDiffResult_DispatchesToTable(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/test/key", Status: DiffStatusAdded, NewValue: "value"},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return PrintDiffResult(result, "table", false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "KEY")
}

func TestPrintDiffJSON_EmptyResult(t *testing.T) {
	result := &DiffResult{Entries: []*DiffEntry{}}

	output, err := captureStdout(t, func() error {
		return printDiffJSON(result, false)
	})
	require.NoError(t, err)

	var parsed struct {
		Entries   []interface{} `json:"entries"`
		Added     int           `json:"added"`
		Modified  int           `json:"modified"`
		Deleted   int           `json:"deleted"`
		Unchanged int           `json:"unchanged"`
	}
	err = json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)
	assert.Empty(t, parsed.Entries)
	assert.Equal(t, 0, parsed.Added)
}

func TestPrintDiffJSON_AllStatuses(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/added", Status: DiffStatusAdded, NewValue: "new"},
			{Key: "/modified", Status: DiffStatusModified, OldValue: "old", NewValue: "new"},
			{Key: "/deleted", Status: DiffStatusDeleted, OldValue: "old"},
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Added:     1,
		Modified:  1,
		Deleted:   1,
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffJSON(result, false)
	})
	require.NoError(t, err)

	var parsed struct {
		Entries []struct {
			Key      string `json:"key"`
			Status   string `json:"status"`
			OldValue string `json:"old_value,omitempty"`
			NewValue string `json:"new_value,omitempty"`
		} `json:"entries"`
		Added     int `json:"added"`
		Modified  int `json:"modified"`
		Deleted   int `json:"deleted"`
		Unchanged int `json:"unchanged"`
	}
	err = json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed.Entries, 3)
	assert.Equal(t, 1, parsed.Added)
	assert.Equal(t, 1, parsed.Modified)
	assert.Equal(t, 1, parsed.Deleted)
	assert.Equal(t, 1, parsed.Unchanged)

	statusMap := make(map[string]string)
	for _, e := range parsed.Entries {
		statusMap[e.Key] = e.Status
	}
	assert.Equal(t, "added", statusMap["/added"])
	assert.Equal(t, "modified", statusMap["/modified"])
	assert.Equal(t, "deleted", statusMap["/deleted"])
}

func TestPrintDiffJSON_ShowUnchanged(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffJSON(result, true)
	})
	require.NoError(t, err)

	var parsed struct {
		Entries []struct {
			Key    string `json:"key"`
			Status string `json:"status"`
		} `json:"entries"`
	}
	err = json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed.Entries, 1)
	assert.Equal(t, "unchanged", parsed.Entries[0].Status)
}

func TestPrintDiffJSON_ValidJSON(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/key", Status: DiffStatusAdded, NewValue: "value with \"quotes\" and\nnewlines"},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffJSON(result, false)
	})
	require.NoError(t, err)

	var raw interface{}
	err = json.Unmarshal([]byte(output), &raw)
	require.NoError(t, err, "Output should be valid JSON")
}

func TestPrintDiffResult_DispatchesToYAML(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/test/key", Status: DiffStatusAdded, NewValue: "value"},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return PrintDiffResult(result, "yaml", false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "added:")
	assert.Contains(t, output, "/test/key")
}

func TestPrintDiffYAML_EmptyResult(t *testing.T) {
	result := &DiffResult{Entries: []*DiffEntry{}}

	output, err := captureStdout(t, func() error {
		return printDiffYAML(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "added: 0")
	assert.Contains(t, output, "modified: 0")
	assert.Contains(t, output, "deleted: 0")
	assert.Contains(t, output, "entries: []")
}

func TestPrintDiffYAML_AllStatuses(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/added", Status: DiffStatusAdded, NewValue: "new"},
			{Key: "/modified", Status: DiffStatusModified, OldValue: "old", NewValue: "new"},
			{Key: "/deleted", Status: DiffStatusDeleted, OldValue: "old"},
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Added:     1,
		Modified:  1,
		Deleted:   1,
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffYAML(result, false)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "added: 1")
	assert.Contains(t, output, "modified: 1")
	assert.Contains(t, output, "deleted: 1")
	assert.Contains(t, output, "/added")
	assert.Contains(t, output, "/modified")
	assert.Contains(t, output, "/deleted")
	assert.NotContains(t, output, "/unchanged")
}

func TestPrintDiffYAML_ShowUnchanged(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffYAML(result, true)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "unchanged: 1")
	assert.Contains(t, output, "/unchanged")
}

func TestPrintDiffYAML_ValidYAML(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/key", Status: DiffStatusAdded, NewValue: "value with \"quotes\""},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffYAML(result, false)
	})
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err, "Output should be valid YAML")
}

func TestPrintDiffSimple_NoChanges(t *testing.T) {
	result := &DiffResult{Entries: []*DiffEntry{}}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "No changes detected")
}

func TestPrintDiffSimple_AddedEntries(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/app/config/key1", Status: DiffStatusAdded, NewValue: "value1"},
			{Key: "/app/config/key2", Status: DiffStatusAdded, NewValue: "value2"},
		},
		Added: 2,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Added (2)")
	assert.Contains(t, output, "/app/config/key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "+2 ~0 -0")
}

func TestPrintDiffSimple_ModifiedEntries(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/app/key", Status: DiffStatusModified, OldValue: "old", NewValue: "new"},
		},
		Modified: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Modified (1)")
	assert.Contains(t, output, "/app/key")
	assert.Contains(t, output, "old")
	assert.Contains(t, output, "new")
}

func TestPrintDiffSimple_DeletedEntries(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/app/deleted", Status: DiffStatusDeleted, OldValue: "deleted_value"},
		},
		Deleted: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Deleted (1)")
	assert.Contains(t, output, "/app/deleted")
	assert.Contains(t, output, "deleted_value")
}

func TestPrintDiffSimple_UnchangedNotShown(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "No changes detected")
	assert.NotContains(t, output, "Unchanged")
}

func TestPrintDiffSimple_UnchangedShown(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, true)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Unchanged (1)")
	assert.Contains(t, output, "/unchanged")
	assert.Contains(t, output, "=1")
}

func TestPrintDiffSimple_MixedStatuses(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/added", Status: DiffStatusAdded, NewValue: "new"},
			{Key: "/modified", Status: DiffStatusModified, OldValue: "old", NewValue: "new"},
			{Key: "/deleted", Status: DiffStatusDeleted, OldValue: "old"},
		},
		Added:    1,
		Modified: 1,
		Deleted:  1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Added (1)")
	assert.Contains(t, output, "Modified (1)")
	assert.Contains(t, output, "Deleted (1)")
	assert.Contains(t, output, "+1 ~1 -1")
	assert.Contains(t, output, "= 3 total")
}

func TestPrintDiffSimple_Summary(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/a", Status: DiffStatusAdded, NewValue: "v"},
			{Key: "/b", Status: DiffStatusModified, OldValue: "o", NewValue: "n"},
			{Key: "/c", Status: DiffStatusDeleted, OldValue: "o"},
			{Key: "/d", Status: DiffStatusUnchanged, OldValue: "s", NewValue: "s"},
		},
		Added:     1,
		Modified:  1,
		Deleted:   1,
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffSimple(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Summary: +1 ~1 -1")
	assert.NotContains(t, output, "=1")

	output, err = captureStdout(t, func() error {
		return printDiffSimple(result, true)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Summary: +1 ~1 -1 =1")
}

func TestPrintDiffTable_NoChanges(t *testing.T) {
	result := &DiffResult{Entries: []*DiffEntry{}}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "No changes detected")
}

func TestPrintDiffTable_Headers(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/key", Status: DiffStatusAdded, NewValue: "value"},
		},
		Added: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "KEY")
	assert.Contains(t, output, "OLD VALUE")
	assert.Contains(t, output, "NEW VALUE")
}

func TestPrintDiffTable_AllStatuses(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/added", Status: DiffStatusAdded, NewValue: "new"},
			{Key: "/modified", Status: DiffStatusModified, OldValue: "old", NewValue: "new"},
			{Key: "/deleted", Status: DiffStatusDeleted, OldValue: "old"},
		},
		Added:    1,
		Modified: 1,
		Deleted:  1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "/added")
	assert.Contains(t, output, "/modified")
	assert.Contains(t, output, "/deleted")
}

func TestPrintDiffTable_UnchangedFiltered(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "No changes detected")
}

func TestPrintDiffTable_UnchangedShown(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/unchanged", Status: DiffStatusUnchanged, OldValue: "same", NewValue: "same"},
		},
		Unchanged: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, true)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "/unchanged")
}

func TestPrintDiffTable_LongValuesTruncated(t *testing.T) {
	longValue := "this is a very long value that exceeds forty characters and should be truncated"
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/key", Status: DiffStatusModified, OldValue: longValue, NewValue: longValue},
		},
		Modified: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "...")
	assert.NotContains(t, output, longValue)
}

func TestPrintDiffTable_Summary(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			{Key: "/a", Status: DiffStatusAdded, NewValue: "v"},
			{Key: "/b", Status: DiffStatusDeleted, OldValue: "v"},
		},
		Added:   1,
		Deleted: 1,
	}

	output, err := captureStdout(t, func() error {
		return printDiffTable(result, false)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Summary: +1 ~0 -1")
	assert.Contains(t, output, "= 2 total")
}
