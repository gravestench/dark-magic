// Command server is the authoritative game-session composition root.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
)

// main converts composition failure into process status while runMain returns
// normally, allowing deferred content, transport, and authority cleanup to execute.
func main() {
	os.Exit(runMain())
}

// runMain resolves environment and flags before signal-scoped startup so invalid
// policy cannot create an authority or listener that immediately needs rollback.
func runMain() int {
	environment, err := envconfig.Bootstrap("server", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	config, err := parseServerConfig(environment.DefaultPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runServer(ctx, config); err != nil {
		slog.Error("running authoritative game server", "error", err)
		return 1
	}

	return 0
}
