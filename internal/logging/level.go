package logging

import (
	"fmt"
	"log/slog"
	"strings"
)

// ParseLevel converts the user-facing CLI/environment spelling into slog's
// ordered level. Empty input intentionally selects info so callers can pass an
// unset environment value without duplicating defaulting logic.
func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level %q: expected debug, info, warn, or error", value)
	}
}
