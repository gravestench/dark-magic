package headlessshell

import (
	"log/slog"

	"github.com/gravestench/dark-magic/internal/logging"
	"github.com/gravestench/dark-magic/internal/shell"
)

// installLogCapture redirects the process-global logger into the terminal session
// and returns a restoration closure so embedding commands do not retain shell logging.
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
