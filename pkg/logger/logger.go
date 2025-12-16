package logger

import (
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

	logger, _ := config.Build()
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

	logger, _ := config.Build()
	Log = logger.Sugar()
}
