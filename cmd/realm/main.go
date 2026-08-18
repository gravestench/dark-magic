// Command realm is the Realm control-plane composition root.
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

// runMain loads process policy and runs the Realm until shutdown.
func runMain() int {
	environment, err := envconfig.Bootstrap("realm", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	config, err := parseRealmConfig(environment.DefaultPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runRealm(ctx, config); err != nil {
		slog.Error("running Realm", "error", err)
		return 1
	}
	return 0
}
