package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

// runServer assembles content, authority, admissions, and transports.
func runServer(ctx context.Context, config serverConfig) error {
	prepared, err := prepareServerContent(ctx, config)
	if err != nil {
		return err
	}
	defer func() {
		_ = prepared.close()
	}()

	host, err := startGameHost(ctx, prepared, config)
	if err != nil {
		return err
	}
	defer func() {
		_ = host.Close(context.Background())
	}()

	restoredPlayerIDs, err := restoreOrPopulateWorld(host, prepared, config)
	if err != nil {
		return err
	}

	admissions, err := prepareAdmissions(host, restoredPlayerIDs, config)
	if err != nil {
		return err
	}

	quicServer, err := startQUICTransport(host, prepared, admissions, config)
	if err != nil {
		return err
	}

	if quicServer != nil {
		defer func() {
			_ = quicServer.Close()
		}()
	}

	if config.workerConfigured() {
		return runRealmWorker(ctx, host, prepared, quicServer, admissions, config)
	}

	return runStandaloneServer(ctx, host, quicServer, admissions.localProfile, config)
}

// runRealmWorker starts private control, publishes readiness, and awaits a drain.
func runRealmWorker(
	ctx context.Context,
	host *gameserver.Host,
	prepared *serverContent,
	quicServer *sessionquic.Server,
	admissions admissionState,
	config serverConfig,
) error {
	destination, err := prepared.entryWorld.Destination(prepared.entryWorld.Seam.Town.LevelID)
	if err != nil {
		return fmt.Errorf("resolve authoritative d2legacy entry destination: %w", err)
	}

	control, drain, err := startWorkerControl(
		host,
		quicServer,
		destination,
		admissions,
		config,
	)
	if err != nil {
		return err
	}
	defer closeWorkerControl(control)

	readyPath, err := publishWorkerReadiness(control, quicServer, config)
	if err != nil {
		return err
	}

	defer func() {
		_ = os.Remove(readyPath)
	}()

	return serverapp.RunRealmWorker(ctx, host, quicServer, control, drain)
}

// runStandaloneServer coordinates the local session, transport, shell, and persistence.
func runStandaloneServer(
	ctx context.Context,
	host *gameserver.Host,
	quicServer *sessionquic.Server,
	profile serverapp.ProfileAdmission,
	config serverConfig,
) error {
	sessionContext, stopSession := context.WithCancel(ctx)

	sessionErrors := make(chan error, 1)
	go func() {
		sessionErrors <- host.Session.Run(sessionContext)
	}()

	transportErrors := startTransportServing(sessionContext, quicServer)
	shellErr := runAdminShell(ctx, host, config)

	stopSession()

	sessionErr := normalizeCancellation(<-sessionErrors)

	var transportErr error
	if transportErrors != nil {
		transportErr = normalizeCancellation(<-transportErrors)
	}

	profileErr := serverapp.PersistSelectedProfile(host, profile)

	return errors.Join(shellErr, sessionErr, transportErr, profileErr)
}

// startTransportServing starts QUIC alongside a standalone session when configured.
func startTransportServing(
	ctx context.Context,
	quicServer *sessionquic.Server,
) <-chan error {
	if quicServer == nil {
		return nil
	}

	errors := make(chan error, 1)
	go func() {
		errors <- quicServer.Serve(ctx)
	}()

	return errors
}

// runAdminShell runs the mutable local administration shell synchronously.
func runAdminShell(
	ctx context.Context,
	host *gameserver.Host,
	config serverConfig,
) error {
	policy := shell.Policy{Name: "local-server-admin", Mutable: true}

	return headlessshell.Run(
		ctx,
		"server",
		policy,
		config.logLevel,
		os.Stdin,
		os.Stdout,
		modruntime.SessionModule(host.Session),
	)
}

// normalizeCancellation treats coordinated session shutdown as success.
func normalizeCancellation(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
