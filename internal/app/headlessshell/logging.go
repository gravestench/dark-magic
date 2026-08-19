package headlessshell

import (
	"log/slog"

	"github.com/gravestench/dark-magic/internal/logging"
	"github.com/gravestench/dark-magic/internal/shell"
)

// installLogCapture temporarily points slog at a buffer the terminal can show.
func installLogCapture(level slog.Level) (*shell.LogBuffer, func()) {
	logs := shell.NewLogBuffer(1000)
	handler := logging.NewObserverHandler(
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

	// Preserve the process logger because Run only owns logging for its lifetime.
	previous := slog.Default()
	slog.SetDefault(slog.New(handler))
	return logs, func() {
		slog.SetDefault(previous)
	}
}
