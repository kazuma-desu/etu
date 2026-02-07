package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/testutil"
)

func TestConvertCommand(t *testing.T) {
	t.Run("Valid etcdctl file to YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		content := `/app/key
value
`
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		convertOpts.FilePath = testFile
		convertOpts.Format = ""

		output, err := testutil.CaptureStdout(func() error {
			return runConvert(nil, nil)
		})

		require.NoError(t, err)
		assert.Contains(t, output, "app:")
		assert.Contains(t, output, "key: value")
	})

	t.Run("No input file or stdin", func(t *testing.T) {
		convertOpts.FilePath = ""
		convertOpts.Format = ""

		err := runConvert(nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input file required")
	})

	t.Run("Nonexistent file", func(t *testing.T) {
		convertOpts.FilePath = "/nonexistent/file/path/that/does/not/exist"
		convertOpts.Format = ""

		err := runConvert(nil, nil)
		require.Error(t, err)
	})

	t.Run("Key collision error", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "collision.txt")

		content := `/a
value1

/a/b
value2
`
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		convertOpts.FilePath = testFile
		convertOpts.Format = ""

		err = runConvert(nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key collision")
	})
}
