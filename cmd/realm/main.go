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

// main converts composition failure into process status; runMain remains free to
// return normally so deferred repository and listener cleanup can execute.
func main() {
	os.Exit(runMain())
}

// runMain resolves environment and flags before installing signal cancellation.
// This keeps configuration errors distinct from runtime shutdown and avoids partial startup.
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
