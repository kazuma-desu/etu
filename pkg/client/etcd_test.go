package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		{"float64", 3.14, "3.14"},
		{"bool", true, "true"},
		{"nil", nil, ""},
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

func TestDefaultBatchOptions(t *testing.T) {
	opts := DefaultBatchOptions()

	assert.Equal(t, 3, opts.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, opts.InitialBackoff)
	assert.Equal(t, 5*time.Second, opts.MaxBackoff)
	assert.True(t, opts.FallbackToSingleKeys)
	assert.Nil(t, opts.Logger)
}

func TestPutAllResult_FailedKey(t *testing.T) {
	t.Run("returns empty string when no failed keys", func(t *testing.T) {
		result := &PutAllResult{
			FailedKeys: nil,
			Succeeded:  10,
			Total:      10,
		}
		assert.Empty(t, result.FailedKey())
	})

	t.Run("returns empty string for empty slice", func(t *testing.T) {
		result := &PutAllResult{
			FailedKeys: []string{},
			Succeeded:  10,
			Total:      10,
		}
		assert.Empty(t, result.FailedKey())
	})

	t.Run("returns first failed key", func(t *testing.T) {
		result := &PutAllResult{
			FailedKeys: []string{"/key1", "/key2", "/key3"},
			Failed:     3,
			Total:      10,
		}
		assert.Equal(t, "/key1", result.FailedKey())
	})
}

type testLogger struct {
	debugMsgs []string
	infoMsgs  []string
	warnMsgs  []string
	errorMsgs []string
}

func (l *testLogger) Debug(msg string, _ ...any) { l.debugMsgs = append(l.debugMsgs, msg) }
func (l *testLogger) Info(msg string, _ ...any)  { l.infoMsgs = append(l.infoMsgs, msg) }
func (l *testLogger) Warn(msg string, _ ...any)  { l.warnMsgs = append(l.warnMsgs, msg) }
func (l *testLogger) Error(msg string, _ ...any) { l.errorMsgs = append(l.errorMsgs, msg) }

func TestLoggerInterface(t *testing.T) {
	var logger Logger = &testLogger{}

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	tl := logger.(*testLogger)
	assert.Equal(t, []string{"debug message"}, tl.debugMsgs)
	assert.Equal(t, []string{"info message"}, tl.infoMsgs)
	assert.Equal(t, []string{"warn message"}, tl.warnMsgs)
	assert.Equal(t, []string{"error message"}, tl.errorMsgs)
}

func TestWarnLargeValues(t *testing.T) {
	largeValue := strings.Repeat("x", WarnValueSize+1)
	smallValue := "small"

	t.Run("nil logger does not panic", func(t *testing.T) {
		pairs := []*models.ConfigPair{{Key: "/big", Value: largeValue}}
		assert.NotPanics(t, func() { warnLargeValues(nil, pairs) })
	})

	t.Run("no warning for values under threshold", func(t *testing.T) {
		log := &testLogger{}
		pairs := []*models.ConfigPair{
			{Key: "/a", Value: smallValue},
			{Key: "/b", Value: smallValue},
		}
		warnLargeValues(log, pairs)
		assert.Empty(t, log.warnMsgs)
	})

	t.Run("warns for values exceeding threshold", func(t *testing.T) {
		log := &testLogger{}
		pairs := []*models.ConfigPair{
			{Key: "/small", Value: smallValue},
			{Key: "/big", Value: largeValue},
		}
		warnLargeValues(log, pairs)
		require.Len(t, log.warnMsgs, 1)
		assert.Equal(t, "large value may impact performance", log.warnMsgs[0])
	})

	t.Run("warns for each large value", func(t *testing.T) {
		log := &testLogger{}
		pairs := []*models.ConfigPair{
			{Key: "/big1", Value: largeValue},
			{Key: "/big2", Value: largeValue},
		}
		warnLargeValues(log, pairs)
		assert.Len(t, log.warnMsgs, 2)
	})

	t.Run("exact threshold does not warn", func(t *testing.T) {
		log := &testLogger{}
		pairs := []*models.ConfigPair{
			{Key: "/exact", Value: strings.Repeat("x", WarnValueSize)},
		}
		warnLargeValues(log, pairs)
		assert.Empty(t, log.warnMsgs)
	})
}

func TestDryRunClient_PutAllWithOptions_LargeValueWarning(t *testing.T) {
	log := &testLogger{}
	client := NewDryRunClient()
	pairs := []*models.ConfigPair{
		{Key: "/big", Value: strings.Repeat("x", WarnValueSize+1)},
	}
	opts := &BatchOptions{Logger: log}

	result, err := client.PutAllWithOptions(context.Background(), pairs, nil, opts)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Succeeded)
	require.Len(t, log.warnMsgs, 1)
	assert.Equal(t, "large value may impact performance", log.warnMsgs[0])
}

func TestBuildTLSConfig(t *testing.T) {
	t.Run("no TLS config returns nil", func(t *testing.T) {
		cfg := &Config{}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.NoError(t, err)
		assert.Nil(t, tlsConfig)
	})

	t.Run("insecure skip only", func(t *testing.T) {
		cfg := &Config{InsecureSkipTLSVerify: true}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.NoError(t, err)
		require.NotNil(t, tlsConfig)
		assert.True(t, tlsConfig.InsecureSkipVerify)
	})

	t.Run("missing key with cert", func(t *testing.T) {
		cfg := &Config{Cert: "/path/to/cert.crt"}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.Error(t, err)
		assert.Nil(t, tlsConfig)
		assert.Contains(t, err.Error(), "both --cert and --key must be provided for mTLS")
	})

	t.Run("missing cert with key", func(t *testing.T) {
		cfg := &Config{Key: "/path/to/key.key"}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.Error(t, err)
		assert.Nil(t, tlsConfig)
		assert.Contains(t, err.Error(), "both --cert and --key must be provided for mTLS")
	})

	t.Run("invalid CA path", func(t *testing.T) {
		cfg := &Config{CACert: "/nonexistent/ca.crt"}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.Error(t, err)
		assert.Nil(t, tlsConfig)
		assert.Contains(t, err.Error(), "failed to read CA certificate")
	})

	t.Run("invalid CA content", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidCAFile := filepath.Join(tmpDir, "invalid-ca.crt")
		require.NoError(t, os.WriteFile(invalidCAFile, []byte("not a valid PEM"), 0644))

		cfg := &Config{CACert: invalidCAFile}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.Error(t, err)
		assert.Nil(t, tlsConfig)
		assert.Contains(t, err.Error(), "failed to parse CA certificate")
	})

	t.Run("valid CA cert", func(t *testing.T) {
		certDir := createTestCerts(t)
		cfg := &Config{CACert: filepath.Join(certDir, "ca.crt")}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.NoError(t, err)
		require.NotNil(t, tlsConfig)
		assert.NotNil(t, tlsConfig.RootCAs)
	})

	t.Run("valid mTLS config", func(t *testing.T) {
		certDir := createTestCerts(t)
		cfg := &Config{
			CACert: filepath.Join(certDir, "ca.crt"),
			Cert:   filepath.Join(certDir, "client.crt"),
			Key:    filepath.Join(certDir, "client.key"),
		}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.NoError(t, err)
		require.NotNil(t, tlsConfig)
		assert.NotNil(t, tlsConfig.RootCAs)
		assert.Len(t, tlsConfig.Certificates, 1)
	})

	t.Run("invalid cert/key pair", func(t *testing.T) {
		certDir := createTestCerts(t)
		wrongKeyFile := filepath.Join(certDir, "wrong.key")
		require.NoError(t, os.WriteFile(wrongKeyFile, []byte("not a valid key"), 0600))

		cfg := &Config{
			Cert: filepath.Join(certDir, "client.crt"),
			Key:  wrongKeyFile,
		}
		tlsConfig, err := buildTLSConfig(cfg)
		assert.Error(t, err)
		assert.Nil(t, tlsConfig)
		assert.Contains(t, err.Error(), "failed to load client certificate")
	})
}

func createTestCerts(t *testing.T) string {
	t.Helper()
	certDir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caCert, err := x509.ParseCertificate(caCertDER)
	require.NoError(t, err)

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "ca.crt"), caCertPEM, 0644))

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Client"},
		NotBefore:    time.Now().Add(-5 * time.Minute),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	require.NoError(t, err)

	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "client.crt"), clientCertPEM, 0644))

	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	require.NoError(t, err)
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER})
	require.NoError(t, os.WriteFile(filepath.Join(certDir, "client.key"), clientKeyPEM, 0600))

	return certDir
}
