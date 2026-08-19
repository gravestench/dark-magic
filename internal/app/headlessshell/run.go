// Package headlessshell composes the shared administration shell for targets
// that do not own a renderer.
package headlessshell

import (
	"context"
	"io"
	"log/slog"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

// Run starts a renderer-free Lua runtime and its modal administration UI.
func Run(
	ctx context.Context,
	target string,
	policy shell.Policy,
	level slog.Level,
	input io.Reader,
	output io.Writer,
	modules ...modruntime.Module,
) error {
	logs, restoreLogs := installLogCapture(level)
	defer restoreLogs()

	runtime, settings, policy, err := buildRuntime(ctx, policy, modules)
	if err != nil {
		return err
	}
	defer func() {
		// Runtime shutdown is best-effort once the terminal has already exited.
		_ = runtime.Stop(context.Background())
	}()

	session, err := buildSession(target, policy, runtime, settings, logs)
	if err != nil {
		return err
	}
	defer func() {
		// Session cleanup cannot change the terminal's reported result.
		_ = session.Close()
	}()

	slog.Info("administration shell ready", "target", target)

	return runTerminal(ctx, session, input, output)
}
