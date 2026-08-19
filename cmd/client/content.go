package main

import (
	"fmt"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
)

// prepareContent establishes one pinned content view for the whole client run.
// Validation happens before the renderer or Lua runtime starts so a partial mod
// selection cannot fail later after expensive process resources are active.
func prepareContent(selection string) (*distribution.ModSet, *content.FS, error) {
	mods, err := distribution.PrepareMods(selection)
	if err != nil {
		return nil, nil, fmt.Errorf("prepare mod profile: %w", err)
	}

	contentFS, err := content.FromEnvironment(mods.Layers...)
	if err == nil {
		err = content.ValidateClientAssets(contentFS)
	}

	if err != nil {
		_ = mods.Close()
		return nil, nil, err
	}

	return mods, contentFS, nil
}

// closeContent releases package mounts owned by the command. Cleanup failures
// cannot change an already-returned exit code, so they are surfaced through logs.
func closeContent(mods *distribution.ModSet) {
	if err := mods.Close(); err != nil {
		slog.Error("closing mod packages", "error", err)
	}
}
