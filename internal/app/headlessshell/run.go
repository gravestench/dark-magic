// Package headlessshell composes the shared administration shell for targets
// that do not own a renderer.
package headlessshell

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/gravestench/dark-magic/internal/logging"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/luashell"
	"github.com/gravestench/dark-magic/internal/shell/tui"
)

// Run starts a renderer-free Lua runtime and its modal administration UI.
func Run(ctx context.Context, target string, policy shell.Policy, level slog.Level, input io.Reader, output io.Writer, modules ...modruntime.Module) error {
	logs, restoreLogs := installLogCapture(level)
	defer restoreLogs()

	runtime, settings, policy, err := buildRuntime(ctx, policy, modules)
	if err != nil {
		return err
	}
	defer runtime.Stop(context.Background())

	session, err := buildSession(target, policy, runtime, settings, logs)
	if err != nil {
		return err
	}
	defer session.Close()

	slog.Info("administration shell ready", "target", target)
	return runTerminal(ctx, session, input, output)
}

// installLogCapture temporarily points slog at a buffer the terminal can show.
func installLogCapture(level slog.Level) (*shell.LogBuffer, func()) {
	logs := shell.NewLogBuffer(1000)
	handler := logging.NewObserverHandler(&slog.HandlerOptions{Level: level}, func(record logging.Record) {
		logs.Append(shell.LogEntry{At: record.At, Level: record.Level.String(), Message: record.Message, Attributes: record.Attributes})
	})
	previous := slog.Default()
	slog.SetDefault(slog.New(handler))
	return logs, func() { slog.SetDefault(previous) }
}

func buildRuntime(ctx context.Context, policy shell.Policy, modules []modruntime.Module) (*modruntime.Runtime, *shell.Settings, shell.Policy, error) {
	runtime := modruntime.New()
	for _, module := range modules {
		if err := runtime.RegisterModule(module); err != nil {
			return nil, nil, policy, err
		}
		policy.Capabilities = append(policy.Capabilities, module.Name)
	}
	settings, err := loadSettings()
	if err != nil {
		return nil, nil, policy, err
	}
	if err := runtime.RegisterModule(modruntime.ShellModule(settings)); err != nil {
		return nil, nil, policy, err
	}
	policy.Capabilities = append(policy.Capabilities, "engine.shell/v1")
	if err := runtime.Start(ctx); err != nil {
		return nil, nil, policy, err
	}
	return runtime, settings, policy, nil
}

func loadSettings() (*shell.Settings, error) {
	path, err := darkpaths.ExpandHost(os.Getenv("DARK_MAGIC_SHELL_CONFIG"))
	if err != nil {
		return nil, err
	}
	return shell.NewSettings(path)
}

func buildSession(target string, policy shell.Policy, runtime *modruntime.Runtime, settings *shell.Settings, logs *shell.LogBuffer) (*shell.Session, error) {
	evaluator, err := luashell.NewForPolicy(runtime, policy)
	if err != nil {
		return nil, err
	}
	session, err := shell.NewSession(target+"-local", target, policy, evaluator)
	if err != nil {
		return nil, err
	}
	session.AttachLogs(logs)
	session.AttachSettings(settings)
	return session, nil
}

func runTerminal(ctx context.Context, session *shell.Session, input io.Reader, output io.Writer) error {
	err := tui.Run(ctx, session, input, output)
	if errors.Is(err, tea.ErrInterrupted) || errors.Is(err, tea.ErrProgramKilled) {
		return nil
	}
	return err
}
