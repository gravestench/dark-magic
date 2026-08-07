// Package headlessshell composes the shared administration shell for targets
// that do not own a renderer.
package headlessshell

import (
	"context"
	"errors"
	"io"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/gravestench/dark-magic/internal/logging"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/luashell"
	"github.com/gravestench/dark-magic/internal/shell/tui"
)

// Run starts a renderer-free Lua runtime and its modal Charm administration UI.
func Run(ctx context.Context, target string, policy shell.Policy, level slog.Level, input io.Reader, output io.Writer) error {
	logs := shell.NewLogBuffer(1000)
	handler := logging.NewObserverHandler(&slog.HandlerOptions{Level: level}, func(record logging.Record) {
		logs.Append(shell.LogEntry{
			At: record.At, Level: record.Level.String(), Message: record.Message, Attributes: record.Attributes,
		})
	})
	previous := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(previous)

	runtime := modruntime.New()
	if err := runtime.Start(ctx); err != nil {
		return err
	}
	defer runtime.Stop(context.Background())
	evaluator, err := luashell.NewForPolicy(runtime, policy)
	if err != nil {
		return err
	}
	session, err := shell.NewSession(target+"-local", target, policy, evaluator)
	if err != nil {
		return err
	}
	session.AttachLogs(logs)
	defer session.Close()
	slog.Info("administration shell ready", "target", target)
	if err := tui.Run(ctx, session, input, output); err != nil && !errors.Is(err, tea.ErrInterrupted) && !errors.Is(err, tea.ErrProgramKilled) {
		return err
	}
	return nil
}
