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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	runtime := modruntime.New()
	if err := runtime.Start(ctx); err != nil {
		fatal(err)
	}
	defer runtime.Stop(context.Background())

	evaluator, err := luashell.New(runtime)
	if err != nil {
		fatal(err)
	}
	session, err := shell.NewSession("local-terminal", "development", shell.Policy{
		Name:         "local-developer",
		Capabilities: runtime.ModuleNames(),
		Mutable:      true,
	}, evaluator)
	if err != nil {
		fatal(err)
	}
	defer session.Close()

	if err := tui.Run(ctx, session, os.Stdin, os.Stdout); err != nil && !errors.Is(err, tea.ErrInterrupted) {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "darkmagic shell:", err)
	os.Exit(1)
}
