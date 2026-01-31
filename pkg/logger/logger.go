package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var Log *log.Logger

func init() {
	Log = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
		Level:           log.WarnLevel,
	})
}

// SetLevel changes the log level atomically.
// Valid levels: debug, info, warn, error, fatal.
// Invalid levels default to warn.
func SetLevel(level string) {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		lvl = log.WarnLevel
	}
	Log.SetLevel(lvl)
}

// GetLevel returns the current log level as a string.
func GetLevel() string {
	return Log.GetLevel().String()
}

// SetFormatter allows switching between text/JSON/logfmt for CI environments.
// Available formatters: log.TextFormatter, log.JSONFormatter, log.LogfmtFormatter
func SetFormatter(f log.Formatter) {
	Log.SetFormatter(f)
}
