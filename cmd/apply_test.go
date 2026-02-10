package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/models"
)

func TestApplyCommand_Stdin(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		format      models.FormatType
		expectError bool
	}{
		{
			name: "valid etcdctl format from stdin",
			content: `/test/key1
value1

/test/key2
value2
`,
			format:      models.FormatEtcdctl,
			expectError: false,
		},
		{
			name: "valid YAML format from stdin",
			content: `test:
  key1: value1
  key2: value2
`,
			format:      models.FormatYAML,
			expectError: false,
		},
		{
			name:        "valid JSON format from stdin",
			content:     `{"test": {"key1": "value1", "key2": "value2"}}`,
			format:      models.FormatJSON,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetApplyFlags()

			tempDir := t.TempDir()
			stdinFile := filepath.Join(tempDir, "stdin.txt")
			err := os.WriteFile(stdinFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			f, err := os.Open(stdinFile)
			require.NoError(t, err)
			defer f.Close()
			os.Stdin = f

			applyOpts.FilePath = "-"
			applyOpts.Format = tt.format
			applyOpts.DryRun = true
			applyOpts.NoValidate = false
			applyOpts.Strict = false

			err = runApply(applyCmd, []string{})
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func resetApplyFlags() {
	applyOpts.FilePath = ""
	applyOpts.Format = ""
	applyOpts.DryRun = false
	applyOpts.NoValidate = false
	applyOpts.Strict = false
}
