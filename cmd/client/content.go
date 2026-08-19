package main

import (
	"fmt"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
)

// prepareContent resolves the selected mods and validates the resulting client asset view.
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

// closeContent releases the mounted mod packages and reports cleanup failures.
func closeContent(mods *distribution.ModSet) {
	if err := mods.Close(); err != nil {
		slog.Error("closing mod packages", "error", err)
	}
}
