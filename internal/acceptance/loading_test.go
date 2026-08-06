package acceptance

import (
	"context"

	"github.com/gravestench/dark-magic/internal/loading"
)

func acceptanceLoadingCoordinator() *loading.Coordinator {
	return acceptanceLoadingCoordinatorWithWorld(nil)
}

func acceptanceLoadingCoordinatorWithWorld(release <-chan struct{}) *loading.Coordinator {
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
	return loading.New(map[string]loading.Task{
		"selected_character": ready,
		"loading_assets":     ready,
		"world":              world,
	})
}
