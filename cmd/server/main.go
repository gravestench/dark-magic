// Command server is the standalone game-session composition root.
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

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/logging"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

func main() {
	logLevel := flag.String("log-level", "info", "log verbosity: debug, info, warn, or error")
	sessionID := flag.String("session-id", "standalone", "stable allocated game-session ID")
	quicListen := flag.String("quic-listen", "", "serve authenticated game sessions on this UDP address")
	tlsCertificate := flag.String("tls-cert", "", "PEM server certificate for QUIC")
	tlsKey := flag.String("tls-key", "", "PEM private key for QUIC")
	admissionKey := flag.String("admission-key", "", "file containing the realm-shared admission HMAC key")
	playerProfile := flag.String("player-profile", "", "player-controlled profile for an explicitly self-hosted character")
	profilePlayer := flag.String("profile-player", "", "stable player ID for the selected self-hosted profile character")
	profileX := flag.Float64("profile-x", 0, "authoritative profile-character spawn X")
	profileY := flag.Float64("profile-y", 0, "authoritative profile-character spawn Y")
	profileWidth := flag.Float64("profile-world-width", 0, "authoritative profile-character world width")
	profileHeight := flag.Float64("profile-world-height", 0, "authoritative profile-character world height")
	profileAct := flag.Int64("profile-act", 0, "authoritative profile-character act")
	profileLevel := flag.Int64("profile-level", 0, "authoritative profile-character level ID")
	flag.Parse()
	level, err := logging.ParseLevel(*logLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	contentFS, err := content.FromEnvironment()
	if err != nil {
		slog.Error("mounting authoritative content", "error", err)
		return
	}
	records := recordstore.New(contentFS)
	host, err := gameserver.Start(ctx, contentFS, records, gameserver.Config{
		Mode: gameserver.ModeStandalone, SessionID: *sessionID, Prediction: gamesession.PredictionLimited,
	})
	if err != nil {
		slog.Error("starting authoritative game server", "error", err)
		return
	}
	defer host.Close(context.Background())
	profilePath, err := darkpaths.ExpandHost(*playerProfile)
	if err != nil {
		slog.Error("expanding player profile path", "error", err)
		return
	}
	var profileAdmission serverapp.ProfileAdmission
	if profilePath != "" {
		destination, destinationErr := playeradapter.NewDestination(*profileX, *profileY, *profileWidth, *profileHeight, *profileAct, *profileLevel)
		if destinationErr != nil {
			slog.Error("validating player-profile destination", "error", destinationErr)
			return
		}
		profileAdmission = serverapp.ProfileAdmission{Path: profilePath, PlayerID: *profilePlayer, Destination: destination}
		if err := serverapp.AdmitSelectedProfile(host, profileAdmission); err != nil {
			slog.Error("admitting selected player-profile character", "error", err)
			return
		}
	}
	quicServer, err := serverapp.StartQUIC(serverapp.QUICConfig{
		Address: *quicListen, CertificatePath: *tlsCertificate, PrivateKeyPath: *tlsKey,
		AdmissionKeyPath: *admissionKey, SessionID: *sessionID,
	}, host)
	if err != nil {
		slog.Error("starting QUIC game-session transport", "error", err)
		return
	}
	if quicServer != nil {
		defer quicServer.Close()
		slog.Info("serving authenticated game sessions", "address", quicServer.Addr())
	}
	sessionContext, stopSession := context.WithCancel(ctx)
	sessionErrors := make(chan error, 1)
	go func() { sessionErrors <- host.Session.Run(sessionContext) }()
	transportErrors := make(chan error, 1)
	if quicServer != nil {
		go func() { transportErrors <- quicServer.Serve(sessionContext) }()
	}
	policy := shell.Policy{Name: "local-server-admin", Mutable: true}
	shellErr := headlessshell.Run(ctx, "server", policy, level, os.Stdin, os.Stdout, modruntime.SessionModule(host.Session))
	stopSession()
	sessionErr := <-sessionErrors
	if errors.Is(sessionErr, context.Canceled) {
		sessionErr = nil
	}
	var transportErr error
	if quicServer != nil {
		transportErr = <-transportErrors
		if errors.Is(transportErr, context.Canceled) {
			transportErr = nil
		}
	}
	profileErr := serverapp.PersistSelectedProfile(host, profileAdmission)
	if err := errors.Join(shellErr, sessionErr, transportErr, profileErr); err != nil {
		slog.Error("running standalone server", "error", err)
		os.Exit(1)
	}
}
