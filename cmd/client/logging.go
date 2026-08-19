package main

import (
	"log/slog"

	"github.com/gravestench/dark-magic/internal/logging"
	"github.com/gravestench/dark-magic/internal/shell"
)

// configureLogging routes structured records to stderr and the in-game console.
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

// parseLogLevel translates the public command value into the logger's level type.
func parseLogLevel(value string) (slog.Level, error) {
	return logging.ParseLevel(value)
}
