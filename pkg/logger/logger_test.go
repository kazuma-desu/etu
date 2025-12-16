package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"default level", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.level)
			if Log == nil {
				t.Error("Expected logger to be initialized")
			}
		})
	}
}

func TestLoggerInitialization(t *testing.T) {
	if Log == nil {
		t.Error("Expected global logger to be initialized")
	}
}

func TestLoggerOutput(_ *testing.T) {
	SetLevel("info")
	Log.Info("test message")
	Log.Infow("test with fields", "key", "value")
	Log.Debug("debug message")
	Log.Debugw("debug with fields", "key", "value")
	Log.Warn("warn message")
	Log.Warnw("warn with fields", "key", "value")
	Log.Error("error message")
	Log.Errorw("error with fields", "key", "value")
}

func TestSetLevelCoverage(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}
	for _, level := range levels {
		SetLevel(level)
		if Log == nil {
			t.Errorf("Logger should be initialized after SetLevel(%s)", level)
		}
	}
}

func TestLoggerLevel(t *testing.T) {
	SetLevel("debug")
	if !Log.Desugar().Core().Enabled(zapcore.DebugLevel) {
		t.Error("Debug level should be enabled")
	}

	SetLevel("error")
	if Log.Desugar().Core().Enabled(zapcore.DebugLevel) {
		t.Error("Debug level should be disabled when level is error")
	}
}
