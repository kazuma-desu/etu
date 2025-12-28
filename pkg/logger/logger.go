package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)

	logger, err := config.Build()
	if err != nil {
		logger = zap.NewNop()
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logger: %v\n", err)
	}
	Log = logger.Sugar()
}

func SetLevel(level string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.WarnLevel
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := config.Build()
	if err != nil {
		logger = zap.NewNop()
		fmt.Fprintf(os.Stderr, "Warning: failed to set log level: %v\n", err)
	}
	Log = logger.Sugar()
}
