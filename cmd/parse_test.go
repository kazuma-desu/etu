package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommand(t *testing.T) {
	t.Run("Parse with valid etcdctl format", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.txt")

		content := `/app/name
myapp

/app/version
1.0.0

/app/port
8080
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		parseOpts.FilePath = configFile
		parseOpts.Format = "etcdctl"
		outputFormat = "simple"

		err = runParse(parseCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Parse with JSON output", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.txt")

		content := `/test/key
value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		parseOpts.FilePath = configFile
		parseOpts.Format = "etcdctl"
		outputFormat = "json"

		err = runParse(parseCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Parse with tree view", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.txt")

		content := `/app/db/host
localhost

/app/db/port
5432
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		parseOpts.FilePath = configFile
		parseOpts.Format = "etcdctl"
		outputFormat = "tree"

		err = runParse(parseCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Parse with auto-detect format", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "auto.txt")

		content := `/auto/key
value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		parseOpts.FilePath = configFile
		parseOpts.Format = "" // Auto-detect
		outputFormat = "simple"

		err = runParse(parseCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Parse with nonexistent file", func(t *testing.T) {
		parseOpts.FilePath = "/nonexistent/file.txt"
		parseOpts.Format = "etcdctl"
		outputFormat = "simple"

		err := runParse(parseCmd, []string{})
		assert.Error(t, err)
	})

	t.Run("Parse with empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "empty.txt")

		err := os.WriteFile(configFile, []byte(""), 0644)
		require.NoError(t, err)

		parseOpts.FilePath = configFile
		parseOpts.Format = "etcdctl"
		outputFormat = "simple"

		err = runParse(parseCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Parse with complex nested structure", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "nested.txt")

		content := `/service/api/endpoint
https://api.example.com

/service/api/timeout
30

/service/db/host
localhost

/service/db/port
5432

/service/cache/enabled
true
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		parseOpts.FilePath = configFile
		parseOpts.Format = "etcdctl"
		outputFormat = "json"

		err = runParse(parseCmd, []string{})
		assert.NoError(t, err)
	})
}
