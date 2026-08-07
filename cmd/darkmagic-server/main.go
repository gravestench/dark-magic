// Command darkmagic-server is the standalone game-session composition root.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/logging"
	"github.com/gravestench/dark-magic/internal/shell"
)

func main() {
	logLevel := flag.String("log-level", "info", "log verbosity: debug, info, warn, or error")
	flag.Parse()
	level, err := logging.ParseLevel(*logLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	policy := shell.Policy{Name: "local-server-admin", Mutable: true}
	if err := headlessshell.Run(ctx, "server", policy, level, os.Stdin, os.Stdout); err != nil {
		slog.Error("running standalone server", "error", err)
		os.Exit(1)
	}
}
