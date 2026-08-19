package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
)

// serverContent keeps mounted packages beside the records and maps derived from
// them. This ownership ensures derived authority cannot outlive its pinned content view.
type serverContent struct {
	mods           *distribution.ModSet
	contentFS      *content.FS
	d2legacySource fs.FS
	records        *recordstore.Store
	assetSetID     string
	entryWorld     *entryworld.Prepared
}

// prepareServerContent establishes one immutable content generation before the
// host exists. Any failure closes mounts immediately, avoiding half-owned assets.
func prepareServerContent(ctx context.Context, config serverConfig) (*serverContent, error) {
	mods, err := distribution.PrepareMods(config.mods)
	if err != nil {
		return nil, fmt.Errorf("prepare mod profile: %w", err)
	}

	prepared, err := buildServerContent(ctx, mods, config.gameDifficulty)
	if err != nil {
		_ = mods.Close()
		return nil, err
	}

	return prepared, nil
}

// buildServerContent derives records, runtime identity inputs, and entry maps
// from the same resolved package set so clients can verify a coherent generation.
func buildServerContent(
	ctx context.Context,
	mods *distribution.ModSet,
	difficulty int,
) (*serverContent, error) {
	contentFS, err := content.FromEnvironment(mods.Layers...)
	if err != nil {
		return nil, fmt.Errorf("mount authoritative content: %w", err)
	}

	assetSetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("identify external game assets: %w", err)
	}

	slog.Info("validated external game asset set", "asset_set_id", assetSetID)

	d2legacySource, err := fs.Sub(contentFS, "mods/d2legacy")
	if err != nil {
		return nil, fmt.Errorf("resolve d2legacy package: %w", err)
	}

	records := recordstore.New(contentFS)

	entryWorld, err := prepareEntryWorld(ctx, contentFS, d2legacySource, records, difficulty)
	if err != nil {
		return nil, err
	}

	return &serverContent{
		mods:           mods,
		contentFS:      contentFS,
		d2legacySource: d2legacySource,
		records:        records,
		assetSetID:     assetSetID,
		entryWorld:     entryWorld,
	}, nil
}

// prepareEntryWorld builds collision and admission geometry before session start;
// players must never join an authority whose destination maps are still partial.
func prepareEntryWorld(
	ctx context.Context,
	contentFS *content.FS,
	d2legacySource fs.FS,
	records *recordstore.Store,
	difficulty int,
) (*entryworld.Prepared, error) {
	recoveredData, err := recovered.New(contentFS).Snapshot()
	if err != nil {
		return nil, fmt.Errorf("load recovered d2legacy records: %w", err)
	}

	resolver, err := worldobjects.New(recoveredData, records)
	if err != nil {
		return nil, fmt.Errorf("build d2legacy world-object resolver: %w", err)
	}

	world, err := entryworld.Build(
		ctx,
		contentFS,
		d2legacySource,
		records,
		resolver,
		1,
		difficulty,
	)
	if err != nil {
		return nil, fmt.Errorf("prepare authoritative d2legacy entry world: %w", err)
	}

	return world, nil
}

// close releases the package mounts that back every prepared record and map.
// Callers must close the host first so no runtime reads from an invalidated view.
func (prepared *serverContent) close() error {
	return prepared.mods.Close()
}
