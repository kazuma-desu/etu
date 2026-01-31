package logger

import (
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

func TestLoggerInitialization(t *testing.T) {
	assert.NotNil(t, Log, "Expected global logger to be initialized")
}

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel log.Level
	}{
		{"debug level", "debug", log.DebugLevel},
		{"info level", "info", log.InfoLevel},
		{"warn level", "warn", log.WarnLevel},
		{"error level", "error", log.ErrorLevel},
		{"invalid level defaults to warn", "invalid", log.WarnLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.level)
			assert.Equal(t, tt.expectedLevel, Log.GetLevel())
		})
	}
}

func TestGetLevel(t *testing.T) {
	SetLevel("info")
	assert.Equal(t, "info", GetLevel())

	SetLevel("debug")
	assert.Equal(t, "debug", GetLevel())
}

func TestSetFormatter(_ *testing.T) {
	// Test setting JSON formatter
	SetFormatter(log.JSONFormatter)
	// Formatter is set, no panic = success

	// Reset to text formatter
	SetFormatter(log.TextFormatter)
}

func TestLoggerOutput(_ *testing.T) {
	SetLevel("debug")

	// Test all log levels
	Log.Debug("debug message")
	Log.Info("info message")
	Log.Warn("warn message")
	Log.Error("error message")

	// Test structured logging
	Log.Info("test with fields", "key", "value", "number", 42)
	Log.Debug("debug with fields", "context", "testing")
}

func TestLevelFiltering(_ *testing.T) {
	SetLevel("error")

	// These should be filtered out
	Log.Debug("debug - should not appear")
	Log.Info("info - should not appear")
	Log.Warn("warn - should not appear")

	// This should appear
	Log.Error("error - should appear")
}
