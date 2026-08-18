package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gravestench/dark-magic/internal/app/headlessshell"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/shell"
)

// runRealm assembles owned resources and coordinates their shutdown.
func runRealm(ctx context.Context, config realmConfig) error {
	directory, err := realm.DataDirectory(config.dataDirectory)
	if err != nil {
		return err
	}
	postgres, err := openRealmRepositories(ctx, config)
	if err != nil {
		return err
	}
	defer postgres.Close()

	allocator, processAllocator, err := prepareWorkerAllocator(directory, config)
	if err != nil {
		return err
	}
	control, err := buildControlPlane(ctx, postgres, allocator, config)
	if err != nil {
		return err
	}
	if err := startMailWorker(ctx, postgres, control, config); err != nil {
		return err
	}
	assets, closeAssets, err := preparePortalAssets(directory)
	if err != nil {
		return err
	}
	defer closeAssets()

	servers, err := startRealmServers(control, assets, directory, config)
	if err != nil {
		return err
	}
	shellErrors := startAdminShell(ctx, config)
	startMaintenance(ctx, control, processAllocator != nil, config.workerHealthInterval)
	runErr := waitForRealmStop(ctx, servers, shellErrors)
	return errors.Join(runErr, shutdownRealm(servers, processAllocator))
}

// startAdminShell begins the optional mutable local administration shell.
func startAdminShell(ctx context.Context, config realmConfig) <-chan error {
	if !config.adminShell {
		return nil
	}
	errors := make(chan error, 1)
	policy := shell.Policy{Name: "local-realm-admin", Mutable: true}
	go func() {
		errors <- headlessshell.Run(
			ctx,
			"realm",
			policy,
			config.logLevel,
			os.Stdin,
			os.Stdout,
		)
	}()
	return errors
}

// startMaintenance begins session pruning and worker reconciliation.
func startMaintenance(
	ctx context.Context,
	control *realm.ControlPlane,
	hasProcessAllocator bool,
	interval time.Duration,
) {
	go realm.RunMaintenance(ctx, control, hasProcessAllocator, interval, logMaintenanceResult)
}

// logMaintenanceResult records failures and state-changing maintenance passes.
func logMaintenanceResult(result realm.MaintenanceResult) {
	attributes := []any{
		"pruned_sessions", result.PrunedSessions,
		"pruned_presence", result.PrunedPresence,
		"reconciled_games", result.ReconciledGames,
	}
	if result.Err != nil {
		slog.Warn("running Realm maintenance", append(attributes, "error", result.Err)...)
		return
	}
	if result.PrunedSessions > 0 || result.PrunedPresence > 0 || result.ReconciledGames > 0 {
		slog.Info("completed Realm maintenance", attributes...)
	}
}

// waitForRealmStop waits for cancellation or a component failure.
func waitForRealmStop(
	ctx context.Context,
	servers realmServers,
	shellErrors <-chan error,
) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-shellErrors:
		return err
	case err := <-servers.publicErrors:
		return normalizeServerError(err)
	case err := <-servers.operatorErrors:
		return normalizeServerError(err)
	}
}

// normalizeServerError treats an intentional HTTP shutdown as success.
func normalizeServerError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// shutdownRealm closes listeners and supervised workers within a bounded window.
func shutdownRealm(
	servers realmServers,
	processAllocator *realm.ProcessAllocator,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := servers.public.Shutdown(ctx)
	if servers.operator != nil {
		err = errors.Join(err, servers.operator.Shutdown(ctx))
	}
	if processAllocator != nil {
		err = errors.Join(err, processAllocator.Close(ctx))
	}
	return err
}
