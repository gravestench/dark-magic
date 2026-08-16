// Command server is the standalone game-session composition root.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/logging"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

func main() {
	environment, err := envconfig.Bootstrap("server", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = flag.String("env-file", environment.DefaultPath, "environment file (overrides the default server.env selection)")
	logLevel := flag.String("log-level", environmentDefault("DARK_MAGIC_SERVER_LOG_LEVEL", "info"), "log verbosity: trace, debug, info, warn, or error")
	sessionID := flag.String("session-id", "standalone", "stable allocated game-session ID")
	allocationID := flag.String("allocation-id", "", "Realm-owned durable allocation generation")
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
	remoteProfileKey := flag.String("remote-profile-key", "", "protected file containing the self-host profile admission credential")
	realmWorker := flag.Bool("realm-worker", false, "run as a Realm-supervised game worker without an interactive shell")
	workerControlListen := flag.String("worker-control-listen", "", "serve the private Realm worker-control API on this TCP address")
	workerControlToken := flag.String("worker-control-token", "", "owner-only file containing the Realm worker-control bearer token")
	workerReadyFile := flag.String("worker-ready-file", "", "owner-only readiness rendezvous written after worker transports start")
	restoreCheckpoint := flag.String("restore-checkpoint", "", "owner-only Realm recovery checkpoint used before worker startup")
	modsFlag := flag.String("mods", os.Getenv("DARK_MAGIC_MODS"), "temporary comma-separated mod IDs from the installed profile")
	flag.Parse()
	level, err := logging.ParseLevel(*logLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	workerConfigured := *realmWorker || *workerControlListen != "" || *workerControlToken != "" || *workerReadyFile != ""
	if workerConfigured && (!*realmWorker || *allocationID == "" || *workerControlListen == "" || *workerControlToken == "" || *workerReadyFile == "" ||
		*quicListen == "" || *tlsCertificate == "" || *tlsKey == "" || *admissionKey == "") {
		slog.Error("configuring Realm worker", "error", "realm-worker, worker control, readiness, QUIC, TLS, and admission flags must be set together")
		return
	}
	if workerConfigured && (*playerProfile != "" || *remoteProfileKey != "") {
		slog.Error("configuring Realm worker", "error", "Realm workers cannot admit player-controlled profiles")
		return
	}
	if *restoreCheckpoint != "" && !*realmWorker {
		slog.Error("configuring recovery checkpoint", "error", "restore-checkpoint is valid only for Realm workers")
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	mods, err := distribution.PrepareMods(*modsFlag)
	if err != nil {
		slog.Error("preparing mod profile", "error", err)
		return
	}
	defer mods.Close()
	contentFS, err := content.FromEnvironment(mods.Layers...)
	if err != nil {
		slog.Error("mounting authoritative content", "error", err)
		return
	}
	assetSetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		slog.Error("identifying external game assets", "error", err)
		return
	}
	slog.Info("validated external game asset set", "asset_set_id", assetSetID)
	records := recordstore.New(contentFS)
	d2legacySource, err := fs.Sub(contentFS, "mods/d2legacy")
	if err != nil {
		slog.Error("resolving d2legacy package", "error", err)
		return
	}
	recoveredData, err := recovered.New(contentFS).Snapshot()
	if err != nil {
		slog.Error("loading recovered d2legacy records", "error", err)
		return
	}
	objectResolver, err := worldobjects.New(recoveredData, records)
	if err != nil {
		slog.Error("building d2legacy world-object resolver", "error", err)
		return
	}
	preparedWorld, err := entryworld.Build(ctx, contentFS, d2legacySource, records, objectResolver, 1)
	if err != nil {
		slog.Error("preparing authoritative d2legacy entry world", "error", err)
		return
	}
	entryDestination, err := preparedWorld.Destination(preparedWorld.Seam.Town.LevelID)
	if err != nil {
		slog.Error("resolving authoritative d2legacy entry destination", "error", err)
		return
	}
	mode := gameserver.ModeStandalone
	if *realmWorker {
		mode = gameserver.ModeRealm
	}
	host, err := gameserver.Start(ctx, d2legacySource, records, gameserver.Config{
		Mode: mode, SessionID: *sessionID, Prediction: gamesession.PredictionLimited, Packages: mods.Packages,
		Content: contentFS, Mods: &mods.Resolved, InitialData: preparedWorld.InitialData("", false), AssetSetID: assetSetID,
	})
	if err != nil {
		slog.Error("starting authoritative game server", "error", err)
		return
	}
	defer host.Close(context.Background())
	if err := preparedWorld.InstallCollision(ctx, host.Authority.Runtime); err != nil {
		slog.Error("installing authoritative d2legacy collision maps", "error", err)
		return
	}
	restored := false
	var restoredPlayerIDs []string
	if *restoreCheckpoint != "" {
		recoveryPath, pathErr := darkpaths.ExpandHost(*restoreCheckpoint)
		if pathErr != nil {
			slog.Error("expanding Realm recovery checkpoint path", "error", pathErr)
			return
		}
		recovery, recoveryErr := serverapp.ReadGameRecovery(recoveryPath)
		if recoveryErr == nil {
			recoveryErr = host.Allocation.ValidateCheckpoint(recovery.Checkpoint.State, nil)
		}
		if recoveryErr == nil {
			recoveryErr = host.Session.RestoreRecoveryCheckpoint(recovery.Checkpoint)
		}
		if recoveryErr != nil {
			slog.Error("restoring authoritative Realm worker checkpoint", "error", recoveryErr)
			return
		}
		restored = true
		restoredPlayerIDs = append([]string(nil), recovery.PlayerIDs...)
		slog.Info("restored authoritative Realm worker checkpoint", "tick", recovery.Checkpoint.State.Tick,
			"players", len(restoredPlayerIDs))
	}
	if !restored {
		population, populationErr := preparedWorld.PopulationCommand(0)
		if populationErr == nil {
			populationErr = host.Session.Submit(population)
		}
		if populationErr != nil {
			slog.Error("queuing authoritative d2legacy population", "error", populationErr)
			return
		}
	}
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
	var remoteProfileConfig *serverapp.RemoteProfileConfig
	if *remoteProfileKey != "" {
		if profilePath != "" {
			slog.Error("configuring player profiles", "error", "local and remote profile admission are mutually exclusive")
			return
		}
		credential, credentialErr := serverapp.ReadAdmissionKey(*remoteProfileKey)
		if credentialErr != nil {
			slog.Error("reading remote profile key", "error", credentialErr)
			return
		}
		destination, destinationErr := playeradapter.NewDestination(*profileX, *profileY, *profileWidth, *profileHeight, *profileAct, *profileLevel)
		if destinationErr != nil {
			slog.Error("validating remote profile destination", "error", destinationErr)
			return
		}
		remoteProfileConfig = &serverapp.RemoteProfileConfig{Credential: string(credential), PrincipalID: "self-host:remote-user",
			PlayerID: *profilePlayer, Destination: destination, Lifetime: 30 * time.Second}
	}
	var workerTickets *gameserver.TicketAuthority
	var workerMemberships *realm.WorkerMemberships
	if workerConfigured {
		secret, secretErr := serverapp.ReadAdmissionKey(*admissionKey)
		if secretErr != nil {
			slog.Error("reading Realm worker admission key", "error", secretErr)
			return
		}
		workerTickets, err = gameserver.NewTicketAuthority(secret, *sessionID)
		if err != nil {
			slog.Error("creating Realm worker ticket authority", "error", err)
			return
		}
		workerMemberships = realm.NewWorkerMemberships()
		for _, playerID := range restoredPlayerIDs {
			workerMemberships.Admit(playerID, time.Time{})
		}
	}
	quicServer, err := serverapp.StartQUIC(serverapp.QUICConfig{
		Address: *quicListen, CertificatePath: *tlsCertificate, PrivateKeyPath: *tlsKey,
		AdmissionKeyPath: *admissionKey, SessionID: *sessionID, RemoteProfile: remoteProfileConfig,
		ModCache: mods.Cache, Tickets: workerTickets, RealmMemberships: workerMemberships,
	}, host)
	if err != nil {
		slog.Error("starting QUIC game-session transport", "error", err)
		return
	}
	if quicServer != nil {
		defer quicServer.Close()
		slog.Info("serving authenticated game sessions", "address", quicServer.Addr())
	}
	if workerConfigured {
		drain := make(chan struct{}, 1)
		controlServer, controlErr := serverapp.StartWorkerControl(serverapp.WorkerControlConfig{
			Address: *workerControlListen, CertificatePath: *tlsCertificate, PrivateKeyPath: *tlsKey,
			TokenPath: *workerControlToken, Tickets: workerTickets, Destination: entryDestination,
			Memberships: workerMemberships,
			Drain: func() {
				// Tell the run coordinator this is an intentional shutdown before
				// closing QUIC. The private HTTP request is acknowledged only after
				// this callback returns, so public traffic is still fenced first.
				select {
				case drain <- struct{}{}:
				default:
				}
				_ = quicServer.Close()
			},
		}, host)
		if controlErr != nil {
			slog.Error("starting Realm worker control", "error", controlErr)
			return
		}
		defer controlServer.Close(context.Background())
		readyPath, pathErr := darkpaths.ExpandHost(*workerReadyFile)
		if pathErr != nil {
			slog.Error("expanding Realm worker readiness path", "error", pathErr)
			return
		}
		if err := realm.WriteWorkerProcessReady(readyPath, realm.WorkerProcessReady{
			GameID: *sessionID, AllocationID: *allocationID, ProcessID: os.Getpid(),
			ControlAddress: controlServer.Addr().String(), GameEndpoint: realm.GameEndpoint{
				Address: quicServer.Addr(), TLSFingerprint: controlServer.TLSFingerprint(),
			},
		}); err != nil {
			slog.Error("writing Realm worker readiness", "error", err)
			return
		}
		defer os.Remove(readyPath)
		slog.Info("Realm worker ready", "control_address", controlServer.Addr(), "game_address", quicServer.Addr(),
			"tls_fingerprint", controlServer.TLSFingerprint())
		if err := serverapp.RunRealmWorker(ctx, host, quicServer, controlServer, drain); err != nil {
			slog.Error("running Realm worker", "error", err)
			os.Exit(1)
		}
		return
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

func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
