package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/testutil"
)

func TestPrintStatusSimple(t *testing.T) {
	tests := []struct {
		name       string
		endpoints  []string
		statuses   map[string]*client.StatusResponse
		firstError error
		wantErr    bool
		contains   []string
	}{
		{
			name:      "single healthy endpoint",
			endpoints: []string{"http://localhost:2379"},
			statuses: map[string]*client.StatusResponse{
				"http://localhost:2379": {
					Version:          "3.5.12",
					DbSize:           1024000,
					Leader:           12345,
					RaftIndex:        100,
					RaftTerm:         5,
					RaftAppliedIndex: 100,
					IsLearner:        false,
					Errors:           []string{},
				},
			},
			firstError: nil,
			wantErr:    false,
			contains: []string{
				"Cluster Status",
				"http://localhost:2379",
				"HEALTHY",
				"3.5.12",
				"12345",
				"Summary",
				"Healthy:   1",
			},
		},
		{
			name:      "unhealthy endpoint",
			endpoints: []string{"http://localhost:2379"},
			statuses: map[string]*client.StatusResponse{
				"http://localhost:2379": nil,
			},
			firstError: errors.New("connection refused"),
			wantErr:    true,
			contains: []string{
				"Cluster Status",
				"http://localhost:2379",
				"UNHEALTHY",
				"Failed to connect",
				"Summary",
				"Healthy:   0",
				"Unhealthy: 1",
			},
		},
		{
			name:      "learner node",
			endpoints: []string{"http://localhost:2379"},
			statuses: map[string]*client.StatusResponse{
				"http://localhost:2379": {
					Version:          "3.5.12",
					DbSize:           1024000,
					Leader:           12345,
					RaftIndex:        100,
					RaftTerm:         5,
					RaftAppliedIndex: 100,
					IsLearner:        true,
					Errors:           []string{},
				},
			},
			firstError: nil,
			wantErr:    false,
			contains: []string{
				"HEALTHY",
				"Learner",
			},
		},
		{
			name:      "endpoint with errors",
			endpoints: []string{"http://localhost:2379"},
			statuses: map[string]*client.StatusResponse{
				"http://localhost:2379": {
					Version:          "3.5.12",
					DbSize:           1024000,
					Leader:           12345,
					RaftIndex:        100,
					RaftTerm:         5,
					RaftAppliedIndex: 100,
					IsLearner:        false,
					Errors:           []string{"error1", "error2"},
				},
			},
			firstError: nil,
			wantErr:    false,
			contains: []string{
				"HEALTHY",
				"Errors:",
				"error1",
				"error2",
			},
		},
		{
			name:      "multiple endpoints",
			endpoints: []string{"http://localhost:2379", "http://localhost:2380"},
			statuses: map[string]*client.StatusResponse{
				"http://localhost:2379": {
					Version:          "3.5.12",
					DbSize:           1024000,
					Leader:           12345,
					RaftIndex:        100,
					RaftTerm:         5,
					RaftAppliedIndex: 100,
				},
				"http://localhost:2380": nil,
			},
			firstError: errors.New("one endpoint down"),
			wantErr:    true,
			contains: []string{
				"http://localhost:2379",
				"http://localhost:2380",
				"HEALTHY",
				"UNHEALTHY",
				"Healthy:   1",
				"Unhealthy: 1",
				"Total:     2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := testutil.CaptureStdout(func() error {
				return printStatusSimple(tt.endpoints, tt.statuses, tt.firstError)
			})
			// printStatusSimple always returns nil now; error is handled by caller
			require.NoError(t, err)
			for _, want := range tt.contains {
				assert.Contains(t, output, want)
			}
		})
	}
}

func TestPrintStatusJSON(t *testing.T) {
	endpoints := []string{"http://localhost:2379", "http://localhost:2380"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version:          "3.5.12",
			DbSize:           1024000,
			Leader:           12345,
			RaftIndex:        100,
			RaftTerm:         5,
			RaftAppliedIndex: 100,
			IsLearner:        false,
			Errors:           []string{},
		},
		"http://localhost:2380": nil,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printStatusJSON(endpoints, statuses, errors.New("connection failed"))
	})
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	endpointsList, ok := result["endpoints"].([]any)
	require.True(t, ok)
	assert.Len(t, endpointsList, 2)

	ep1 := endpointsList[0].(map[string]any)
	assert.Equal(t, "http://localhost:2379", ep1["endpoint"])
	assert.Equal(t, true, ep1["healthy"])
	assert.Equal(t, "3.5.12", ep1["version"])
	assert.Equal(t, float64(1024000), ep1["dbSize"])
	assert.Equal(t, float64(12345), ep1["leader"])
	assert.Equal(t, float64(100), ep1["raftIndex"])
	assert.Equal(t, float64(5), ep1["raftTerm"])

	ep2 := endpointsList[1].(map[string]any)
	assert.Equal(t, "http://localhost:2380", ep2["endpoint"])
	assert.Equal(t, false, ep2["healthy"])
	assert.Empty(t, ep2["version"])

	summary, ok := result["summary"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(1), summary["healthy"])
	assert.Equal(t, float64(1), summary["unhealthy"])
	assert.Equal(t, float64(2), summary["total"])

	assert.Equal(t, "some endpoints are unreachable", result["warning"])
}

func TestPrintStatusYAML(t *testing.T) {
	endpoints := []string{"http://localhost:2379"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version:          "3.5.12",
			DbSize:           1024000,
			Leader:           12345,
			RaftIndex:        100,
			RaftTerm:         5,
			RaftAppliedIndex: 100,
			IsLearner:        false,
			Errors:           []string{},
		},
	}

	output, err := testutil.CaptureStdout(func() error {
		return printStatusYAML(endpoints, statuses, nil)
	})
	require.NoError(t, err)

	assert.Contains(t, output, "endpoints:")
	assert.Contains(t, output, "endpoint: http://localhost:2379")
	assert.Contains(t, output, "healthy: true")
	assert.Contains(t, output, "version: 3.5.12")
	assert.Contains(t, output, "summary:")
	assert.Contains(t, output, "healthy: 1")
	assert.Contains(t, output, "total: 1")
}

func TestPrintStatusYAMLWithWarning(t *testing.T) {
	endpoints := []string{"http://localhost:2379", "http://localhost:2380"}
	statuses := map[string]*client.StatusResponse{
		"http://localhost:2379": {
			Version:   "3.5.12",
			DbSize:    1024000,
			Leader:    12345,
			RaftIndex: 100,
			RaftTerm:  5,
		},
		"http://localhost:2380": nil,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printStatusYAML(endpoints, statuses, errors.New("connection failed"))
	})
	require.NoError(t, err)

	assert.Contains(t, output, "warning: some endpoints are unreachable")
	assert.Contains(t, output, "healthy: 1")
	assert.Contains(t, output, "unhealthy: 1")
}

func TestRunStatus_NotConnected(t *testing.T) {
	origContextName := contextName
	defer func() { contextName = origContextName }()

	contextName = "nonexistent-context-for-testing"

	err := runStatus(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
	assert.Contains(t, err.Error(), "etu login")
}

func TestRunStatus_InvalidOutputFormat(t *testing.T) {
	origFormat := outputFormat
	origContextName := contextName
	origConfig := os.Getenv("ETUCONFIG")
	defer func() {
		outputFormat = origFormat
		contextName = origContextName
		os.Setenv("ETUCONFIG", origConfig)
	}()

	// Create a temp config with a valid context so runStatus reaches format validation
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

	os.Setenv("ETUCONFIG", configPath)
	outputFormat = "invalid"
	contextName = "test-context"

	err = runStatus(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}
