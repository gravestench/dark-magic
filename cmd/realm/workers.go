package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/content"
)

// prepareWorkerAllocator creates the optional supervised game-worker allocator.
func prepareWorkerAllocator(
	directory string,
	config realmConfig,
) (realm.GameAllocator, *realm.ProcessAllocator, error) {
	if config.workerExecutable == "" {
		slog.Warn("Realm game creation is disabled; configure --worker-executable")
		return nil, nil, nil
	}

	assetSetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		return nil, nil, fmt.Errorf("identify worker game assets: %w", err)
	}

	slog.Info("validated worker game asset set", "asset_set_id", assetSetID)

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable:           config.workerExecutable,
		Arguments:            []string{"--log-level", config.logLevelName},
		StateDirectory:       filepath.Join(directory, "workers"),
		ControlListenAddress: config.workerControlListen,
		GameListenAddress:    config.workerGameListen,
		GameAdvertiseHost:    config.workerGameAdvertiseHost,
		LogWriter:            os.Stderr,
		ExpectedAssetSetID:   assetSetID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("configure Realm worker allocator: %w", err)
	}

	return allocator, allocator, nil
}
