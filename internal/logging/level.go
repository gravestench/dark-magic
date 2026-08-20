package logging

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// LevelTrace is intentionally below slog's built-in debug level. It is for
// high-frequency implementation diagnostics such as renderer and packet flow.
const LevelTrace slog.Level = slog.LevelDebug - 4

// Trace records high-volume diagnostics without forcing callers to understand the custom level value. Falling back to
// slog.Default mirrors the package-level slog helpers and keeps optional subsystem loggers safe to call.
func Trace(logger *slog.Logger, message string, args ...any) {
	if logger == nil {
		logger = slog.Default()
	}

	logger.Log(context.Background(), LevelTrace, message, args...)
}

// ParseLevel converts the user-facing CLI/environment spelling into slog's
// ordered level. Empty input intentionally selects info so callers can pass an
// unset environment value without duplicating defaulting logic.
func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "trace":
		return LevelTrace, nil
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level %q: expected trace, debug, info, warn, or error", value)
	}
}
