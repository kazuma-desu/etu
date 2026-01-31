package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/testutil"
)

func TestKeyValue(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple", "/app/name", "myapp"},
		{"empty value", "/app/empty", ""},
		{"with spaces", "/app/config", "hello world"},
		{"special chars", "/app/key", "value-with-dash_underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := testutil.CaptureStdoutFunc(func() {
				KeyValue(tt.key, tt.value)
			})
			require.NoError(t, err)

			assert.Contains(t, output, tt.key)
			assert.Contains(t, output, tt.value)
		})
	}
}

func TestKeyValueWithMetadata(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		metadata [][2]string
	}{
		{
			name:     "with metadata",
			key:      "/config/key",
			value:    "value",
			metadata: [][2]string{{"rev", "1"}, {"version", "2"}},
		},
		{
			name:     "empty metadata",
			key:      "/config/key",
			value:    "value",
			metadata: [][2]string{},
		},
		{
			name:     "single metadata",
			key:      "/config/key",
			value:    "value",
			metadata: [][2]string{{"CreateRevision", "123"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := testutil.CaptureStdoutFunc(func() {
				KeyValueWithMetadata(tt.key, tt.value, tt.metadata)
			})
			require.NoError(t, err)

			assert.Contains(t, output, tt.key)
			assert.Contains(t, output, tt.value)
			for _, kv := range tt.metadata {
				assert.Contains(t, output, kv[0])
				assert.Contains(t, output, kv[1])
			}
		})
	}
}
