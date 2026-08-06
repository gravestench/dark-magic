package acceptance

import (
	"context"

	"github.com/gravestench/dark-magic/internal/loadcore"
)

func acceptanceLoadingCoordinator() *loadcore.Coordinator {
	return acceptanceLoadingCoordinatorWithWorld(nil)
}

func acceptanceLoadingCoordinatorWithWorld(release <-chan struct{}) *loadcore.Coordinator {
	ready := func(context.Context) error { return nil }
	world := ready
	if release != nil {
		world = func(ctx context.Context) error {
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return loadcore.New(map[string]loadcore.Task{
		"selected_character": ready,
		"loading_assets":     ready,
		"world":              world,
	})
}
