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
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
	gamestate "github.com/gravestench/dark-magic/internal/game/state"
	"github.com/gravestench/dark-magic/internal/logging"
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
	if err := gameplayer.Register(authority); err != nil {
		slog.Error("registering authoritative player commands", "error", err)
		return
	}
	if err := gamesession.RegisterMovement(authority); err != nil {
		slog.Error("registering authoritative movement commands", "error", err)
		return
	}
	if err := gamesession.RegisterSkillAssignments(authority); err != nil {
		slog.Error("registering authoritative skill commands", "error", err)
		return
	}
	if err := gameskill.RegisterIntentConsumer(engine); err != nil {
		slog.Error("registering authoritative skill intent consumer", "error", err)
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
