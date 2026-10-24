package clientapp

import (
	"errors"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// preparedVelocityPaths publishes only prepared geometry, including legacy two-zone worlds.
func preparedVelocityPaths(worlds map[int]*gameworld.Map) (map[int64]gameworld.VelocityPathFinder, error) {
	paths := make(map[int64]gameworld.VelocityPathFinder, len(worlds))
	for level, world := range worlds {
		if world != nil {
			paths[int64(level)] = world
		}
	}
	if len(paths) == 0 {
		return nil, errors.New("register velocity movement: no prepared collision maps")
	}
	return paths, nil
}
