package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/config"
)

func TestRunCompletion(t *testing.T) {
	testCmd := &cobra.Command{Use: "test"}
	rootCmd.AddCommand(testCmd)
	defer rootCmd.RemoveCommand(testCmd)

	t.Run("bash completion", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"bash"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "bash completion")
	})

	t.Run("zsh completion", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"zsh"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "zsh completion")
	})

	t.Run("fish completion", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"fish"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "fish")
	})

	t.Run("powershell completion", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runCompletion(completionCmd, []string{"powershell"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		require.NoError(t, err)
		assert.Contains(t, output, "PowerShell")
	})

	t.Run("unknown shell returns nil", func(t *testing.T) {
		err := runCompletion(completionCmd, []string{"unknown"})
		assert.NoError(t, err)
	})
}

func TestCompleteConfigFiles(t *testing.T) {
	t.Run("returns file extensions", func(t *testing.T) {
		extensions, directive := completeConfigFiles(nil, nil, "")

		assert.Equal(t, []string{"txt", "yaml", "yml", "json"}, extensions)
		assert.Equal(t, cobra.ShellCompDirectiveFilterFileExt, directive)
	})
}

func TestCompleteContextNames(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	t.Run("returns empty for no contexts", func(t *testing.T) {
		cfg := &config.Config{
			Contexts: map[string]*config.ContextConfig{},
		}
		err := config.SaveConfig(cfg)
		require.NoError(t, err)

		contexts, directive := completeContextNames(nil, nil, "")

		assert.Empty(t, contexts)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("returns sorted context names", func(t *testing.T) {
		cfg := &config.Config{
			Contexts: map[string]*config.ContextConfig{
				"prod":    {Endpoints: []string{"http://prod:2379"}},
				"dev":     {Endpoints: []string{"http://dev:2379"}},
				"staging": {Endpoints: []string{"http://staging:2379"}},
			},
		}
		err := config.SaveConfig(cfg)
		require.NoError(t, err)

		contexts, directive := completeContextNames(nil, nil, "")

		assert.Equal(t, []string{"dev", "prod", "staging"}, contexts)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("returns single context", func(t *testing.T) {
		cfg := &config.Config{
			Contexts: map[string]*config.ContextConfig{
				"only": {Endpoints: []string{"http://only:2379"}},
			},
		}
		err := config.SaveConfig(cfg)
		require.NoError(t, err)

		contexts, directive := completeContextNames(nil, nil, "")

		assert.Equal(t, []string{"only"}, contexts)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})
}

func TestCompleteContextNamesForArg(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	cfg := &config.Config{
		Contexts: map[string]*config.ContextConfig{
			"ctx1": {Endpoints: []string{"http://ctx1:2379"}},
			"ctx2": {Endpoints: []string{"http://ctx2:2379"}},
		},
	}
	err := config.SaveConfig(cfg)
	require.NoError(t, err)

	t.Run("returns contexts when no args", func(t *testing.T) {
		contexts, directive := CompleteContextNamesForArg(nil, []string{}, "")

		assert.Equal(t, []string{"ctx1", "ctx2"}, contexts)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("returns nil when args already provided", func(t *testing.T) {
		contexts, directive := CompleteContextNamesForArg(nil, []string{"existing"}, "")

		assert.Nil(t, contexts)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})

	t.Run("returns nil when multiple args provided", func(t *testing.T) {
		contexts, directive := CompleteContextNamesForArg(nil, []string{"arg1", "arg2"}, "")

		assert.Nil(t, contexts)
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	})
}

func TestRegisterFileCompletion(t *testing.T) {
	t.Run("registers completion function", func(_ *testing.T) {
		testCmd := &cobra.Command{Use: "test"}
		testCmd.Flags().String("file", "", "test file flag")

		registerFileCompletion(testCmd, "file")
	})

	t.Run("handles non-existent flag gracefully", func(_ *testing.T) {
		testCmd := &cobra.Command{Use: "test"}

		registerFileCompletion(testCmd, "nonexistent")
	})
}

func TestCompleteContextNames_LoadError(t *testing.T) {
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	configDir := tempDir + "/.config/etu"
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(configDir+"/config.yaml", []byte("invalid: yaml: content: ["), 0o644)

	contexts, directive := completeContextNames(nil, nil, "")

	assert.Nil(t, contexts)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}
