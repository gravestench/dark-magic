package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// startGameHost creates the sole authoritative simulation from pinned content,
// then installs every collision map before commands can advance the session.
func startGameHost(
	ctx context.Context,
	prepared *serverContent,
	config serverConfig,
) (*gameserver.Host, error) {
	mode := gameserver.ModeStandalone
	if config.realmWorker {
		mode = gameserver.ModeRealm
	}

	initialData := prepared.entryWorld.InitialData("", false)
	initialData["d2legacy.game_rules"] = gameRules(config)

	host, err := gameserver.Start(
		ctx,
		prepared.d2legacySource,
		prepared.records,
		gameserver.Config{
			Mode:        mode,
			SessionID:   config.sessionID,
			Prediction:  gamesession.PredictionLimited,
			Packages:    prepared.mods.Packages,
			Content:     prepared.contentFS,
			Mods:        &prepared.mods.Resolved,
			InitialData: initialData,
			AssetSetID:  prepared.assetSetID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("start authoritative game server: %w", err)
	}

	if err := prepared.entryWorld.InstallCollision(ctx, host.Authority.Runtime); err != nil {
		_ = host.Close(context.Background())
		return nil, fmt.Errorf("install authoritative d2legacy collision maps: %w", err)
	}

	return host, nil
}

// gameRules serializes process policy into deterministic runtime input. These
// values participate in behavior and must match the identity clients accept.
func gameRules(config serverConfig) map[string]any {
	return map[string]any{
		"target":          "lod-1.14d",
		"expansion":       true,
		"difficulty":      config.gameDifficulty,
		"hardcore":        config.gameHardcore,
		"ladder":          config.gameLadder,
		"maximum_players": config.gameMaximumPlayers,
	}
}

// restoreOrPopulateWorld chooses exactly one authority origin. A checkpoint must
// replace bootstrap population; applying both would duplicate players and world entities.
func restoreOrPopulateWorld(
	host *gameserver.Host,
	prepared *serverContent,
	config serverConfig,
) ([]string, error) {
	if config.restoreCheckpoint == "" {
		population, err := prepared.entryWorld.PopulationCommand(0)
		if err == nil {
			err = host.Session.Submit(population)
		}

		if err != nil {
			return nil, fmt.Errorf("queue authoritative d2legacy population: %w", err)
		}

		return nil, nil
	}

	return restoreCheckpoint(host, config.restoreCheckpoint)
}

// restoreCheckpoint resolves the host path and delegates compatibility checks to
// the host before returning restored player IDs for membership reconstruction.
func restoreCheckpoint(host *gameserver.Host, configuredPath string) ([]string, error) {
	path, err := darkpaths.ExpandHost(configuredPath)
	if err != nil {
		return nil, fmt.Errorf("expand Realm recovery checkpoint path: %w", err)
	}

	recovery, err := serverapp.ReadGameRecovery(path)
	if err == nil {
		err = host.Allocation.ValidateCheckpoint(recovery.Checkpoint.State, nil)
	}

	if err == nil {
		err = host.Session.RestoreRecoveryCheckpoint(recovery.Checkpoint)
	}

	if err != nil {
		return nil, fmt.Errorf("restore authoritative Realm worker checkpoint: %w", err)
	}

	playerIDs := append([]string(nil), recovery.PlayerIDs...)
	slog.Info(
		"restored authoritative Realm worker checkpoint",
		"tick", recovery.Checkpoint.State.Tick,
		"players", len(playerIDs),
	)

	return playerIDs, nil
}
