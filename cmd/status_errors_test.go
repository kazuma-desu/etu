package cmd

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/testutil"
)

func TestPrintStatusJSON_Success(t *testing.T) {
	endpoints := []string{"http://localhost:2379"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version: "3.5.12",
			DbSize:  1024000,
		},
	}

	output, err := testutil.CaptureStdout(func() error {
		return printStatusJSON(endpoints, statuses, nil)
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
	assert.Contains(t, result, "endpoints")
	assert.Contains(t, result, "summary")
}

func TestPrintStatusYAML_Success(t *testing.T) {
	endpoints := []string{"http://localhost:2379"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version: "3.5.12",
			DbSize:  1024000,
		},
	}

	output, err := testutil.CaptureStdout(func() error {
		return printStatusYAML(endpoints, statuses, nil)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "endpoints:")
	assert.Contains(t, output, "summary:")
}

func TestBuildStatusData_WithFirstError(t *testing.T) {
	endpoints := []string{"http://localhost:2379", "http://localhost:2380"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version: "3.5.12",
			DbSize:  1024000,
		},
		"http://localhost:2380": nil,
	}
	firstError := errors.New("connection failed")

	result := buildStatusData(endpoints, statuses, firstError)

	assert.Contains(t, result, "endpoints")
	assert.Contains(t, result, "summary")
	assert.Contains(t, result, "warning")

	warning, ok := result["warning"].(string)
	assert.True(t, ok)
	assert.Equal(t, "some endpoints are unreachable", warning)

	summary, ok := result["summary"].(map[string]int)
	assert.True(t, ok)
	assert.Equal(t, 1, summary["healthy"])
	assert.Equal(t, 1, summary["unhealthy"])
}

func TestBuildStatusData_NoError(t *testing.T) {
	endpoints := []string{"http://localhost:2379"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version: "3.5.12",
			DbSize:  1024000,
		},
	}

	result := buildStatusData(endpoints, statuses, nil)

	assert.Contains(t, result, "endpoints")
	assert.Contains(t, result, "summary")
	assert.NotContains(t, result, "warning")
}
