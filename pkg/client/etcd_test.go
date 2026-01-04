package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kazuma-desu/etu/pkg/models"
)

func generateTestPairs(prefix string, count int) []*models.ConfigPair {
	pairs := make([]*models.ConfigPair, count)
	for i := range count {
		pairs[i] = &models.ConfigPair{
			Key:   fmt.Sprintf("%s/key%04d", prefix, i),
			Value: fmt.Sprintf("value%04d", i),
		}
	}
	return pairs
}

func TestBuildClientOptions(t *testing.T) {
	tests := []struct {
		name          string
		opts          *GetOptions
		expectError   bool
		errorMsg      string
		expectOptions int // Minimum number of options expected
	}{
		{
			name:          "nil options returns empty slice",
			opts:          nil,
			expectError:   false,
			expectOptions: 0,
		},
		{
			name:        "empty options",
			opts:        &GetOptions{},
			expectError: false,
		},
		{
			name: "prefix",
			opts: &GetOptions{Prefix: true},
			// Expect: WithPrefix
			expectOptions: 1,
		},
		{
			name: "from key",
			opts: &GetOptions{FromKey: true},
			// Expect: WithFromKey
			expectOptions: 1,
		},
		{
			name: "range end",
			opts: &GetOptions{RangeEnd: "\x00"},
			// Expect: WithRange
			expectOptions: 1,
		},
		{
			name: "limit",
			opts: &GetOptions{Limit: 100},
			// Expect: WithLimit
			expectOptions: 1,
		},
		{
			name: "revision",
			opts: &GetOptions{Revision: 123},
			// Expect: WithRev
			expectOptions: 1,
		},
		{
			name: "keys only",
			opts: &GetOptions{KeysOnly: true},
			// Expect: WithKeysOnly
			expectOptions: 1,
		},
		{
			name: "count only",
			opts: &GetOptions{CountOnly: true},
			// Expect: WithCountOnly
			expectOptions: 1,
		},
		{
			name: "min mod revision",
			opts: &GetOptions{MinModRev: 10},
			// Expect: WithMinModRev
			expectOptions: 1,
		},
		{
			name: "max mod revision",
			opts: &GetOptions{MaxModRev: 20},
			// Expect: WithMaxModRev
			expectOptions: 1,
		},
		{
			name: "min create revision",
			opts: &GetOptions{MinCreateRev: 10},
			// Expect: WithMinCreateRev
			expectOptions: 1,
		},
		{
			name: "max create revision",
			opts: &GetOptions{MaxCreateRev: 20},
			// Expect: WithMaxCreateRev
			expectOptions: 1,
		},
		{
			name: "sort ascend key",
			opts: &GetOptions{SortOrder: "ASCEND", SortTarget: "KEY"},
			// Expect: WithSort
			expectOptions: 1,
		},
		{
			name: "sort descend version",
			opts: &GetOptions{SortOrder: "DESCEND", SortTarget: "VERSION"},
			// Expect: WithSort
			expectOptions: 1,
		},
		{
			name: "sort create revision",
			opts: &GetOptions{SortOrder: "ASCEND", SortTarget: "CREATE"},
			// Expect: WithSort
			expectOptions: 1,
		},
		{
			name: "sort modify revision",
			opts: &GetOptions{SortOrder: "ASCEND", SortTarget: "MODIFY"},
			// Expect: WithSort
			expectOptions: 1,
		},
		{
			name: "sort value",
			opts: &GetOptions{SortOrder: "ASCEND", SortTarget: "VALUE"},
			// Expect: WithSort
			expectOptions: 1,
		},
		{
			name:        "invalid sort order",
			opts:        &GetOptions{SortOrder: "INVALID"},
			expectError: true,
			errorMsg:    "invalid sort order",
		},
		{
			name:        "invalid sort target",
			opts:        &GetOptions{SortTarget: "INVALID"},
			expectError: true,
			errorMsg:    "invalid sort target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := buildClientOptions(tt.opts)
			if tt.expectError {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(opts), tt.expectOptions)
			}
		})
	}
}

func TestValidateAndPrepareConfig(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *Config
		expectError     bool
		errorMsg        string
		expectedTimeout time.Duration
	}{
		{
			name:        "nil config",
			cfg:         nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "missing endpoints",
			cfg: &Config{
				Endpoints: []string{},
			},
			expectError: true,
			errorMsg:    "at least one endpoint is required",
		},
		{
			name: "valid config with explicit timeout",
			cfg: &Config{
				Endpoints:   []string{"localhost:2379"},
				DialTimeout: 10 * time.Second,
			},
			expectError:     false,
			expectedTimeout: 10 * time.Second,
		},
		{
			name: "valid config applies default timeout",
			cfg: &Config{
				Endpoints: []string{"localhost:2379"},
			},
			expectError:     false,
			expectedTimeout: 5 * time.Second,
		},
		{
			name: "valid config with auth credentials",
			cfg: &Config{
				Endpoints:   []string{"localhost:2379"},
				Username:    "user",
				Password:    "pass",
				DialTimeout: 1 * time.Second,
			},
			expectError:     false,
			expectedTimeout: 1 * time.Second,
		},
		{
			name: "multiple endpoints",
			cfg: &Config{
				Endpoints:   []string{"localhost:2379", "localhost:2380", "localhost:2381"},
				DialTimeout: 3 * time.Second,
			},
			expectError:     false,
			expectedTimeout: 3 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndPrepareConfig(tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTimeout, tt.cfg.DialTimeout)
			}
		})
	}
}

func TestFormatValueUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "test", "test"},
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"float64", 3.14, "3.140000"},
		{"bool", true, "true"},
		{"nil", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("map", func(t *testing.T) {
		mapVal := map[string]any{"key1": "value1", "key2": "value2"}
		result := formatValue(mapVal)
		assert.Contains(t, result, "key1: value1")
		assert.Contains(t, result, "key2: value2")
	})
}

func TestBatchConstants(t *testing.T) {
	t.Run("DefaultMaxOpsPerTxn matches etcd limit", func(t *testing.T) {
		assert.Equal(t, 128, DefaultMaxOpsPerTxn)
	})

	t.Run("WarnValueSize is 100KB", func(t *testing.T) {
		assert.Equal(t, 100*1024, WarnValueSize)
	})
}

func TestGenerateTestPairs(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		count  int
	}{
		{"empty", "/test", 0},
		{"single", "/app/config", 1},
		{"small batch", "/batch", 10},
		{"exact limit", "/limit", DefaultMaxOpsPerTxn},
		{"over limit", "/over", DefaultMaxOpsPerTxn + 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairs := generateTestPairs(tt.prefix, tt.count)

			assert.Len(t, pairs, tt.count)

			for i, pair := range pairs {
				expectedKey := fmt.Sprintf("%s/key%04d", tt.prefix, i)
				expectedValue := fmt.Sprintf("value%04d", i)

				assert.Equal(t, expectedKey, pair.Key)
				assert.Equal(t, expectedValue, pair.Value)
			}
		})
	}
}

func TestGenerateTestPairs_UniqueKeys(t *testing.T) {
	pairs := generateTestPairs("/unique", 100)

	seen := make(map[string]bool)
	for _, pair := range pairs {
		assert.False(t, seen[pair.Key], "duplicate key: %s", pair.Key)
		seen[pair.Key] = true
	}
}
