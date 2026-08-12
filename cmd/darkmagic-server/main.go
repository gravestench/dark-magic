// Command darkmagic-server is the standalone game-session composition root.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	gamestate "github.com/gravestench/dark-magic/internal/game/state"
	"github.com/gravestench/dark-magic/internal/logging"
	d2legacymod "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
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
	engine := gameecs.New()
	authority, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		_ = engine.Close()
		slog.Error("creating authoritative session", "error", err)
		return
	}
	defer authority.Close()
	contentFS, err := content.FromEnvironment()
	if err != nil {
		slog.Error("mounting authoritative content", "error", err)
		return
	}
	records := recordstore.New(contentFS)
	mod, err := d2legacymod.Start(ctx, contentFS, records, engine, authority, 0)
	if err != nil {
		slog.Error("starting d2legacy authority", "error", err)
		return
	}
	defer mod.Stop(context.Background())
	if err := gamesession.RegisterMovement(authority); err != nil {
		slog.Error("registering authoritative movement commands", "error", err)
		return
	}
	if err := gamestate.Register(engine); err != nil {
		slog.Error("registering authoritative timed state engine", "error", err)
		return
	}
	sessionContext, stopSession := context.WithCancel(ctx)
	sessionErrors := make(chan error, 1)
	go func() { sessionErrors <- authority.Run(sessionContext) }()
	policy := shell.Policy{Name: "local-server-admin", Mutable: true}
	shellErr := headlessshell.Run(ctx, "server", policy, level, os.Stdin, os.Stdout, modruntime.SessionModule(authority))
	stopSession()
	sessionErr := <-sessionErrors
	if errors.Is(sessionErr, context.Canceled) {
		sessionErr = nil
	}
	if err := errors.Join(shellErr, sessionErr); err != nil {
		slog.Error("running standalone server", "error", err)
		os.Exit(1)
	}
}
