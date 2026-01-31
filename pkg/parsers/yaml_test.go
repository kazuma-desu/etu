package parsers

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLParser_FormatName(t *testing.T) {
	parser := &YAMLParser{}
	assert.Equal(t, "yaml", parser.FormatName())
}

func TestYAMLParser_SimpleKeyValue(t *testing.T) {
	content := `name: myapp`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/name", "myapp")
}

func TestYAMLParser_NestedMaps(t *testing.T) {
	content := `
app:
  database:
    host: localhost
    port: 5432
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 2)
	assertYAMLPair(t, pairs, "/app/database/host", "localhost")
	assertYAMLPair(t, pairs, "/app/database/port", int64(5432))
}

func TestYAMLParser_DeepNesting(t *testing.T) {
	content := `
level1:
  level2:
    level3:
      level4:
        value: deep
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/level1/level2/level3/level4/value", "deep")
}

func TestYAMLParser_IntegerValue(t *testing.T) {
	content := `port: 8080`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/port", int64(8080))
}

func TestYAMLParser_FloatValue(t *testing.T) {
	content := `rate: 0.95`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/rate", 0.95)
}

func TestYAMLParser_BooleanValues(t *testing.T) {
	content := `
enabled: true
debug: false
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 2)
	assertYAMLPair(t, pairs, "/enabled", true)
	assertYAMLPair(t, pairs, "/debug", false)
}

func TestYAMLParser_NullValueSkipped(t *testing.T) {
	content := `
present: value
absent: null
also_absent: ~
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/present", "value")
}

func TestYAMLParser_ArrayOfScalars(t *testing.T) {
	content := `
tags:
  - dev
  - test
  - prod
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/tags", `["dev","test","prod"]`)
}

func TestYAMLParser_ArrayOfObjects(t *testing.T) {
	content := `
servers:
  - host: server1
    port: 8080
  - host: server2
    port: 8081
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/servers", `[{"host":"server1","port":8080},{"host":"server2","port":8081}]`)
}

func TestYAMLParser_InlineArray(t *testing.T) {
	content := `tags: [a, b, c]`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/tags", `["a","b","c"]`)
}

func TestYAMLParser_MixedTypes(t *testing.T) {
	content := `
app:
  name: myapp
  port: 8080
  rate: 0.5
  enabled: true
  tags:
    - web
    - api
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 5)
	assertYAMLPair(t, pairs, "/app/name", "myapp")
	assertYAMLPair(t, pairs, "/app/port", int64(8080))
	assertYAMLPair(t, pairs, "/app/rate", 0.5)
	assertYAMLPair(t, pairs, "/app/enabled", true)
	assertYAMLPair(t, pairs, "/app/tags", `["web","api"]`)
}

func TestYAMLParser_MultipleRoots(t *testing.T) {
	content := `
app:
  db:
    host: localhost
custom:
  db:
    host: remote
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 2)
	assertYAMLPair(t, pairs, "/app/db/host", "localhost")
	assertYAMLPair(t, pairs, "/custom/db/host", "remote")
}

func TestYAMLParser_EmptyFile(t *testing.T) {
	content := ``

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 0)
}

func TestYAMLParser_DocumentSeparator(t *testing.T) {
	content := `---
name: myapp
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/name", "myapp")
}

func TestYAMLParser_MultiDocument_ParsesFirstOnly(t *testing.T) {
	content := `---
first: doc1
---
second: doc2
`
	var stderr bytes.Buffer
	oldStderr := os.Stderr

	r, w, err := os.Pipe()
	require.NoError(t, err, "failed to create pipe")

	os.Stderr = w
	t.Cleanup(func() { os.Stderr = oldStderr })

	pairs := parseYAML(t, content)

	w.Close()
	stderr.ReadFrom(r)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/first", "doc1")
	assert.Contains(t, stderr.String(), "Warning")
	assert.Contains(t, stderr.String(), "multiple documents")
}

func TestYAMLParser_RootArrayError(t *testing.T) {
	content := `
- item1
- item2
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &YAMLParser{}
	_, err = parser.Parse(context.Background(), tmpFile)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRootNotMap)
}

func TestYAMLParser_Comments(t *testing.T) {
	content := `
# This is a comment
app:
  # Another comment
  name: myapp  # Inline comment
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/app/name", "myapp")
}

func TestYAMLParser_QuotedStrings(t *testing.T) {
	content := `
single: 'hello world'
double: "hello world"
numeric_string: "12345"
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 3)
	assertYAMLPair(t, pairs, "/single", "hello world")
	assertYAMLPair(t, pairs, "/double", "hello world")
	assertYAMLPair(t, pairs, "/numeric_string", "12345")
}

func TestYAMLParser_MultilineString(t *testing.T) {
	content := `
description: |
  This is a
  multiline string
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/description", "This is a\nmultiline string\n")
}

func TestYAMLParser_FoldedString(t *testing.T) {
	content := `
description: >
  This is a
  folded string
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/description", "This is a folded string\n")
}

func TestYAMLParser_Anchors(t *testing.T) {
	content := `
defaults: &defaults
  host: localhost
  port: 5432

production:
  <<: *defaults
  host: prod-db
`

	pairs := parseYAML(t, content)

	assertYAMLPair(t, pairs, "/defaults/host", "localhost")
	assertYAMLPair(t, pairs, "/defaults/port", int64(5432))
	assertYAMLPair(t, pairs, "/production/host", "prod-db")
	assertYAMLPair(t, pairs, "/production/port", int64(5432))
}

func TestYAMLParser_SpecialCharactersInKeys(t *testing.T) {
	content := `
"key.with.dots": value1
key-with-dashes: value2
key_with_underscores: value3
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 3)
	assertYAMLPair(t, pairs, "/key.with.dots", "value1")
	assertYAMLPair(t, pairs, "/key-with-dashes", "value2")
	assertYAMLPair(t, pairs, "/key_with_underscores", "value3")
}

func TestYAMLParser_UnicodeValues(t *testing.T) {
	content := `
greeting: こんにちは
emoji: 🚀
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 2)
	assertYAMLPair(t, pairs, "/greeting", "こんにちは")
	assertYAMLPair(t, pairs, "/emoji", "🚀")
}

func TestYAMLParser_FileNotFound(t *testing.T) {
	parser := &YAMLParser{}
	_, err := parser.Parse(context.Background(), "/nonexistent/file.yaml")

	assert.Error(t, err)
}

func TestYAMLParser_InvalidYAML(t *testing.T) {
	content := `
invalid: [unclosed bracket
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &YAMLParser{}
	_, err = parser.Parse(context.Background(), tmpFile)

	assert.Error(t, err)
}

func TestYAMLParser_ScientificNotation(t *testing.T) {
	content := `
small: 1.5e-10
large: 1.5e10
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 2)
	assertYAMLPair(t, pairs, "/small", 1.5e-10)
	assertYAMLPair(t, pairs, "/large", int64(15000000000))
}

func TestYAMLParser_OctalNumbers(t *testing.T) {
	content := `
octal: 0o755
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/octal", int64(493))
}

func TestYAMLParser_HexNumbers(t *testing.T) {
	content := `
hex: 0xFF
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/hex", int64(255))
}

func TestYAMLParser_NegativeNumbers(t *testing.T) {
	content := `
negative_int: -42
negative_float: -3.14
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 2)
	assertYAMLPair(t, pairs, "/negative_int", int64(-42))
	assertYAMLPair(t, pairs, "/negative_float", -3.14)
}

func TestYAMLParser_EmptyNestedMap(t *testing.T) {
	content := `
parent:
  empty: {}
  present: value
`

	pairs := parseYAML(t, content)

	assert.Len(t, pairs, 1)
	assertYAMLPair(t, pairs, "/parent/present", "value")
}

func parseYAML(t *testing.T, content string) []*models.ConfigPair {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	parser := &YAMLParser{}
	pairs, err := parser.Parse(context.Background(), tmpFile)
	require.NoError(t, err)

	return pairs
}

func assertYAMLPair(t *testing.T, pairs []*models.ConfigPair, key string, expectedValue any) {
	t.Helper()
	for _, p := range pairs {
		if p.Key == key {
			assert.Equal(t, expectedValue, p.Value, "Value mismatch for key %s", key)
			return
		}
	}
	t.Errorf("Key %s not found in pairs", key)
}
