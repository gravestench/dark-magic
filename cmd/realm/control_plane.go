package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

// openRealmRepositories creates the single durable repository set shared by API,
// mail, maintenance, and allocation. Centralizing it prevents subsystems from
// observing different database pools or public account URL policy.
func openRealmRepositories(
	ctx context.Context,
	config realmConfig,
) (*realm.Postgres, error) {
	if strings.TrimSpace(config.postgresURL) == "" {
		return nil, errors.New("realm requires --postgres-url or DARK_MAGIC_REALM_POSTGRES_URL")
	}

	postgres, err := realm.OpenPostgres(ctx, config.postgresURL, 0)
	if err != nil {
		return nil, fmt.Errorf("open Realm PostgreSQL repositories: %w", err)
	}

	if err := postgres.Accounts.SetAccountBaseURL(config.accountBaseURL); err != nil {
		postgres.Close()
		return nil, fmt.Errorf("configure Realm account URL: %w", err)
	}

	slog.Info("using PostgreSQL Realm repositories")

	return postgres, nil
}

// buildControlPlane converts process configuration into one authoritative Realm
// service. Recovery runs before the service is returned so stale allocations
// cannot become visible as healthy games after a process restart.
func buildControlPlane(
	ctx context.Context,
	postgres *realm.Postgres,
	allocator realm.GameAllocator,
	config realmConfig,
) (*realm.ControlPlane, error) {
	audit := realm.AuditSink(realm.NewSlogAuditSink(nil))
	audit = realm.NewAuditFanout(audit, postgres.Audit)

	control, err := realm.NewControlPlane(realm.ControlPlaneConfig{
		Accounts:           postgres.Accounts,
		Characters:         postgres.Characters,
		Games:              postgres.Games,
		Allocations:        postgres.Allocations,
		Memberships:        postgres.Memberships,
		Checkpoints:        postgres.Checkpoints,
		Audit:              audit,
		Allocator:          allocator,
		CheckpointInterval: config.checkpointInterval,
		PresenceTimeout:    config.presenceTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize Realm control plane: %w", err)
	}

	if err := recoverInterruptedGames(ctx, control); err != nil {
		return nil, err
	}

	return control, nil
}

// recoverInterruptedGames fails closed allocations whose supervising process no
// longer exists. Advertising them would hand players tickets to orphaned workers.
func recoverInterruptedGames(ctx context.Context, control *realm.ControlPlane) error {
	recovered, err := control.RecoverInterruptedGames(ctx)
	if err != nil {
		return fmt.Errorf("recover interrupted Realm games after %d recoveries: %w", recovered, err)
	}

	if recovered > 0 {
		slog.Warn("failed closed interrupted Realm games", "recovered_games", recovered)
	}

	return nil
}
