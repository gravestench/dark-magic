// Command shell is the renderer-free development harness for the same shell
// session that client, game-server, and realm adapters consume.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	tea "charm.land/bubbletea/v2"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/luashell"
	"github.com/gravestench/dark-magic/internal/shell/tui"
)

// main owns process-scoped resources so normal shutdown closes the session before stopping its Lua runtime.
func main() {
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignals()

	runtime := startLuaRuntimeOrExit(ctx)
	// The signal-aware context is canceled on interrupt, so shutdown needs a fresh context to finish its cleanup.
	defer stopLuaRuntime(runtime, context.Background())

	session := openDevelopmentSessionOrExit(runtime)
	// LIFO cleanup closes the evaluator-backed session before stopping the runtime it depends on.
	defer closeDevelopmentSession(session)

	runTerminalOrExitOnFailure(ctx, session)
}

// startLuaRuntimeOrExit returns a ready runtime, keeping unusable partial startup out of the remaining workflow.
func startLuaRuntimeOrExit(ctx context.Context) *modruntime.Runtime {
	runtime := modruntime.New()
	if err := runtime.Start(ctx); err != nil {
		exitWithError(err)
	}

	return runtime
}

// stopLuaRuntime retains best-effort cleanup semantics, so a shutdown error cannot replace the terminal's exit status.
func stopLuaRuntime(runtime *modruntime.Runtime, ctx context.Context) {
	_ = runtime.Stop(ctx)
}

// openDevelopmentSessionOrExit binds a trusted local policy to the runtime-backed evaluator used by this harness.
func openDevelopmentSessionOrExit(runtime *modruntime.Runtime) *shell.Session {
	evaluator, err := luashell.New(runtime)
	if err != nil {
		exitWithError(err)
	}

	// This renderer-free harness intentionally exposes every registered module and permits local mutation.
	session, err := shell.NewSession("local-terminal", "development", shell.Policy{
		Name:         "local-developer",
		Capabilities: runtime.ModuleNames(),
		Mutable:      true,
	}, evaluator)
	if err != nil {
		exitWithError(err)
	}

	return session
}

// closeDevelopmentSession retains best-effort cleanup semantics while releasing the evaluator before its Lua runtime.
func closeDevelopmentSession(session *shell.Session) {
	_ = session.Close()
}

// runTerminalOrExitOnFailure treats user interruption as normal termination while surfacing every other TUI failure.
func runTerminalOrExitOnFailure(ctx context.Context, session *shell.Session) {
	if err := tui.Run(ctx, session, os.Stdin, os.Stdout); isUnexpectedTerminalExit(err) {
		exitWithError(err)
	}
}

// isUnexpectedTerminalExit preserves Bubble Tea's wrapped interruption sentinel as a successful user-requested exit.
func isUnexpectedTerminalExit(err error) bool {
	return err != nil && !errors.Is(err, tea.ErrInterrupted)
}

// exitWithError writes the stable harness prefix and terminates immediately, so callers cannot continue after failure.
func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "darkmagic shell:", err)
	os.Exit(1)
}
