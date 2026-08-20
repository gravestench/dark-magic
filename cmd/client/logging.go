package main

import (
	"log/slog"

	"github.com/gravestench/dark-magic/internal/logging"
	"github.com/gravestench/dark-magic/internal/shell"
)

// configureLogging installs one shared handler for terminal diagnostics and the
// in-game console. Using a single handler preserves record ordering across both views.
func configureLogging(level slog.Level) *shell.LogBuffer {
	logs := shell.NewLogBuffer(1000)
	handler := logging.NewHandlerWithObserver(
		&slog.HandlerOptions{Level: level},
		func(record logging.Record) {
			logs.Append(shell.LogEntry{
				At:         record.At,
				Level:      record.Level.String(),
				Message:    record.Message,
				Attributes: record.Attributes,
			})
		},
	)
	slog.SetDefault(slog.New(handler))

	return logs
}

// parseLogLevel validates the human-facing vocabulary at the command boundary
// so internal packages receive a typed level and never interpret user strings.
func parseLogLevel(value string) (slog.Level, error) {
	return logging.ParseLevel(value)
}
