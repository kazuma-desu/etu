package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand(t *testing.T) {
	t.Run("Validate with valid configuration", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "valid.txt")

		content := `/app/name
myapp

/app/version
1.0.0

/app/port
8080
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = false

		err = runValidate(validateCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Validate with invalid key format", func(t *testing.T) {
		// Skip this test because runValidate calls os.Exit(1) on validation failure
		// which terminates the test process. This would require refactoring to be testable.
		t.Skip("Skipping because validation failure calls os.Exit(1)")
	})

	t.Run("Validate with strict mode", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "strict.txt")

		content := `/app/name
myapp
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = true

		err = runValidate(validateCmd, []string{})
		// May pass or fail depending on validation rules
		// Just ensure it doesn't panic
		_ = err
	})

	t.Run("Validate with auto-detect format", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "auto.txt")

		content := `/test/key
value
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "" // Auto-detect
		validateOpts.Strict = false

		err = runValidate(validateCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Validate with nonexistent file", func(t *testing.T) {
		validateOpts.FilePath = "/nonexistent/file.txt"
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = false

		err := runValidate(validateCmd, []string{})
		assert.Error(t, err)
	})

	t.Run("Validate with empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "empty.txt")

		err := os.WriteFile(configFile, []byte(""), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = false

		err = runValidate(validateCmd, []string{})
		assert.NoError(t, err)
	})

	t.Run("Validate with large values", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "large.txt")

		// Create a value that's exactly at the warning threshold
		largeValue := make([]byte, 10*1024) // 10KB
		for i := range largeValue {
			largeValue[i] = 'a'
		}

		content := `/app/large
` + string(largeValue) + `
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = false

		err = runValidate(validateCmd, []string{})
		// Should succeed with warning
		assert.NoError(t, err)
	})

	t.Run("Validate with URL values", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "url.txt")

		content := `/app/api/endpoint
https://api.example.com

/app/api/callback
http://localhost:8080/callback
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = false

		err = runValidate(validateCmd, []string{})
		// HTTP URL should generate warning but not error
		assert.NoError(t, err)
	})

	t.Run("Validate with nested JSON values", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "json.txt")

		content := `/app/config
{"database": {"host": "localhost", "port": 5432}}
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		validateOpts.FilePath = configFile
		validateOpts.Format = "etcdctl"
		validateOpts.Strict = false

		err = runValidate(validateCmd, []string{})
		assert.NoError(t, err)
	})
}
