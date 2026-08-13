// Command realm is the realm control-plane composition root.
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
	policy := shell.Policy{Name: "local-realm-admin", Mutable: true}
	if err := headlessshell.Run(ctx, "realm", policy, level, os.Stdin, os.Stdout); err != nil {
		slog.Error("running realm", "error", err)
		os.Exit(1)
	}
}
