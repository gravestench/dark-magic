// Package entryworld prepares the first authoritative d2legacy game world for
// every topology. Interactive, listen, dedicated, and Realm authorities must
// all consume this adapter so map geometry, initial state, and admission spawn
// cannot drift between composition roots.
package entryworld

import (
	"errors"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// Records is the narrow recovered-data boundary needed by map generation. Keeping the cache behind this interface lets
// every composition root share entry-world policy without depending on a particular storage implementation.
type Records interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

// Prepared owns the generated zones, materialized maps, trusted spawns, and seam as one consistent entry-world result.
// Consumers should not mix fields from different Prepared values because their geometry is seed-dependent.
type Prepared struct {
	Worlds     map[int]*gameworld.Map
	Zones      map[int]*worldgen.Zone
	Spawns     map[int][2]float64
	Seam       gametransition.Seam
	Difficulty int
}

// Destination converts one prepared level into the player's admission contract. Both generated and materialized forms
// must exist, and only the trusted spawn recorded during Build is eligible for player placement.
func (world *Prepared) Destination(levelID int) (playeradapter.Destination, error) {
	if world == nil || world.Worlds[levelID] == nil || world.Zones[levelID] == nil {
		return playeradapter.Destination{}, errors.New("d2legacy entry world: destination level is unavailable")
	}

	spawn, found := world.Spawns[levelID]
	if !found {
		return playeradapter.Destination{}, errors.New("d2legacy entry world: destination has no trusted spawn")
	}

	request := world.Zones[levelID].Request()
	gameMap := world.Worlds[levelID]

	return playeradapter.NewDestination(
		spawn[0],
		spawn[1],
		float64(gameMap.WidthSubtiles),
		float64(gameMap.HeightSubtiles),
		int64(request.Act),
		int64(request.LevelID),
	)
}
