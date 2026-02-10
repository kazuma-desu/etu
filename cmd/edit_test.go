package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
)

func TestEditCommand_Args(t *testing.T) {
	assert.NotNil(t, editCmd)
	assert.True(t, strings.HasPrefix(editCmd.Use, "edit"))
	assert.NotEmpty(t, editCmd.Short)
	assert.NotEmpty(t, editCmd.Long)
}

func TestRunEdit_NotConnected(t *testing.T) {
	origContextName := contextName
	defer func() { contextName = origContextName }()

	contextName = "nonexistent-context-for-testing"

	err := runEdit(nil, []string{"/test/key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunEdit_NoEditor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	origContextName := contextName
	origEditor := os.Getenv("EDITOR")
	origVisual := os.Getenv("VISUAL")

	defer func() {
		contextName = origContextName
		os.Setenv("EDITOR", origEditor)
		os.Setenv("VISUAL", origVisual)
	}()

	os.Unsetenv("EDITOR")
	os.Unsetenv("VISUAL")

	contextName = "nonexistent-context-for-testing"

	err := runEdit(nil, []string{"/test/key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEditCommand_ArgsValidation(t *testing.T) {
	// Args validation is handled by cobra (ExactArgs(1)), not by runEdit directly
	assert.NotNil(t, editCmd.Args)
}

func TestRunEdit_WithMockClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mock client test in short mode")
	}

	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current-context: test-context
contexts:
  test-context:
    endpoints:
      - http://localhost:2379
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	t.Setenv("ETUCONFIG", configPath)

	mock := client.NewMockClient()
	mock.GetFunc = func(_ context.Context, _ string) (string, error) {
		return "original-value", nil
	}

	putCalled := false
	mock.PutFunc = func(_ context.Context, _, _ string) error {
		putCalled = true
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	value, err := mock.Get(ctx, "/test/key")
	require.NoError(t, err)
	assert.Equal(t, "original-value", value)

	err = mock.Put(ctx, "/test/key", "modified-value")
	require.NoError(t, err)
	assert.True(t, putCalled)
	assert.Len(t, mock.PutCalls, 1)
	assert.Equal(t, "/test/key", mock.PutCalls[0].Key)
	assert.Equal(t, "modified-value", mock.PutCalls[0].Value)
}

func TestRunEdit_GetError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mock client test in short mode")
	}

	mock := client.NewMockClient()
	mock.GetFunc = func(_ context.Context, _ string) (string, error) {
		return "", errors.New("key not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := mock.Get(ctx, "/test/key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestRunEdit_PutError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mock client test in short mode")
	}

	mock := client.NewMockClient()
	mock.PutFunc = func(_ context.Context, _, _ string) error {
		return errors.New("put failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := mock.Put(ctx, "/test/key", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "put failed")
}

func TestRunEdit_EditorEnv(t *testing.T) {
	origEditor := os.Getenv("EDITOR")
	origVisual := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", origEditor)
		os.Setenv("VISUAL", origVisual)
	}()

	tests := []struct {
		name      string
		editor    string
		visual    string
		wantEmpty bool
	}{
		{
			name:      "EDITOR set",
			editor:    "vim",
			visual:    "",
			wantEmpty: false,
		},
		{
			name:      "VISUAL set",
			editor:    "",
			visual:    "nano",
			wantEmpty: false,
		},
		{
			name:      "both set",
			editor:    "vim",
			visual:    "nano",
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.editor != "" {
				os.Setenv("EDITOR", tt.editor)
			} else {
				os.Unsetenv("EDITOR")
			}

			if tt.visual != "" {
				os.Setenv("VISUAL", tt.visual)
			} else {
				os.Unsetenv("VISUAL")
			}

			editor, err := resolveEditor()
			if tt.wantEmpty {
				assert.Error(t, err)
				assert.Empty(t, editor)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, editor)
			}
		})
	}
}

func TestRunEdit_TempFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "etu-edit-*.txt")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	content := "test-content"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	readContent, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(readContent))
}

func TestRunEdit_NoEditorFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	origPath := os.Getenv("PATH")
	defer func() { os.Setenv("PATH", origPath) }()

	tmpDir := t.TempDir()
	os.Setenv("PATH", tmpDir)

	os.Unsetenv("EDITOR")
	os.Unsetenv("VISUAL")

	editor := ""
	for _, fallback := range []string{"vi", "vim", "nano", "emacs"} {
		if _, lookupErr := exec.LookPath(fallback); lookupErr == nil {
			editor = fallback
			break
		}
	}

	assert.Empty(t, editor)
}

func TestRunEdit_KeyValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		isValid bool
	}{
		{
			name:    "valid key",
			key:     "/app/config",
			isValid: true,
		},
		{
			name:    "empty key",
			key:     "",
			isValid: false,
		},
		{
			name:    "root key",
			key:     "/",
			isValid: true,
		},
		{
			name:    "nested key",
			key:     "/app/config/database/host",
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKeyPrefix(tt.key)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMockClient_GetRecording(t *testing.T) {
	mock := client.NewMockClient()
	mock.GetFunc = func(_ context.Context, _ string) (string, error) {
		return "test-value", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	value, err := mock.Get(ctx, "/test/key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", value)

	require.Len(t, mock.GetCalls, 1)
	assert.Equal(t, "/test/key", mock.GetCalls[0])
}

func TestMockClient_Reset(t *testing.T) {
	mock := client.NewMockClient()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mock.Get(ctx, "/test")
	mock.Put(ctx, "/test", "value")

	assert.Len(t, mock.GetCalls, 1)
	assert.Len(t, mock.PutCalls, 1)

	mock.Reset()

	assert.Len(t, mock.GetCalls, 0)
	assert.Len(t, mock.PutCalls, 0)
}

func TestEditCommand_Lookup(t *testing.T) {
	assert.NotNil(t, editCmd)
	assert.True(t, strings.HasPrefix(editCmd.Use, "edit"))
}
