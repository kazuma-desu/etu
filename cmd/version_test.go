package cmd

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/testutil"
)

func TestRunVersion_SimpleFormat(t *testing.T) {
	originalFormat := outputFormat
	originalVersion := Version
	originalCommit := Commit
	originalBuildDate := BuildDate

	defer func() {
		outputFormat = originalFormat
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalBuildDate
	}()

	outputFormat = output.FormatSimple.String()
	Version = "1.0.0-test"
	Commit = "abc123"
	BuildDate = "2024-01-01"

	output, err := testutil.CaptureStdout(func() error {
		return runVersion(nil, nil)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "etu version 1.0.0-test")
	assert.Contains(t, output, "commit:")
	assert.Contains(t, output, "abc123")
	assert.Contains(t, output, "built:")
	assert.Contains(t, output, "2024-01-01")
	assert.Contains(t, output, "go version:")
	assert.Contains(t, output, "platform:")
}

func TestRunVersion_JSONFormat(t *testing.T) {
	originalFormat := outputFormat
	originalVersion := Version
	originalCommit := Commit
	originalBuildDate := BuildDate

	defer func() {
		outputFormat = originalFormat
		Version = originalVersion
		Commit = originalCommit
		BuildDate = originalBuildDate
	}()

	outputFormat = output.FormatJSON.String()
	Version = "1.0.0-test"
	Commit = "abc123"
	BuildDate = "2024-01-01"

	output, err := testutil.CaptureStdout(func() error {
		return runVersion(nil, nil)
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0-test", result["version"])
	assert.Equal(t, "abc123", result["commit"])
	assert.Equal(t, "2024-01-01", result["buildDate"])
	assert.Equal(t, runtime.Version(), result["goVersion"])
	assert.NotEmpty(t, result["platform"])
}

func TestRunVersion_DevVersion(t *testing.T) {
	originalFormat := outputFormat
	originalVersion := Version
	originalCommit := Commit

	defer func() {
		outputFormat = originalFormat
		Version = originalVersion
		Commit = originalCommit
	}()

	outputFormat = output.FormatSimple.String()
	Version = "dev"
	Commit = "unknown"

	output, err := testutil.CaptureStdout(func() error {
		return runVersion(nil, nil)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "etu version dev")
}

func TestVersionInfo_Struct(t *testing.T) {
	info := versionInfo{
		Version:   "1.0.0",
		Commit:    "abc123",
		BuildDate: "2024-01-01",
		GoVersion: "go1.21",
		Platform:  "linux/amd64",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", result["version"])
	assert.Equal(t, "abc123", result["commit"])
	assert.Equal(t, "2024-01-01", result["buildDate"])
	assert.Equal(t, "go1.21", result["goVersion"])
	assert.Equal(t, "linux/amd64", result["platform"])
}

func TestVersionCommand_Exists(t *testing.T) {
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)
	assert.NotEmpty(t, versionCmd.Short)
	assert.NotEmpty(t, versionCmd.Long)
}

func TestVersion_Variables(t *testing.T) {
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, BuildDate)
}
