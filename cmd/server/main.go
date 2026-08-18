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

// main owns process exit semantics while runMain owns resource cleanup.
func main() {
	os.Exit(runMain())
}

// runMain loads process policy and runs the configured game server.
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
