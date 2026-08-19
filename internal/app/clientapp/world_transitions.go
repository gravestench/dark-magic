package clientapp

import (
	"github.com/gravestench/akara"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
)

// syncActiveWorldFromPlayer observes the controlled player's committed location and updates map
// presentation only afterward. It never initiates or authorizes a transition itself.
func (app *application) syncActiveWorldFromPlayer() {
	engine := app.presentationSimulation()
	if engine == nil {
		return
	}

	levelID, found := controlledPlayerLevel(engine.World(), app.presentationPlayerID())
	if !found || levelID == app.activeWorldLevel {
		return
	}

	// Lua has already committed the transition. This call only swaps map caches
	// and navigation inputs to follow the authoritative location.
	app.activateWorld(levelID)
}

// presentationPlayerID uses the authority-issued connected player ID when available and the stable
// offline ID otherwise, allowing one lookup path across both ownership models.
func (app *application) presentationPlayerID() string {
	if app.network == nil {
		return "local-player"
	}

	playerID, _ := app.network.Status()["player_id"].(string)
	if playerID == "" {
		return "local-player"
	}

	return playerID
}

// controlledPlayerLevel joins control ownership to location on the same entity. Matching by player
// label or the first location would let a peer or stale entity drive the active world.
func controlledPlayerLevel(world *akara.World, playerID string) (int, bool) {
	controls, found := akara.GetDynamicStore(world, "d2legacy.world.player_control")
	if !found {
		return 0, false
	}

	locations, found := akara.GetDynamicStore(world, "d2legacy.world.location")
	if !found {
		return 0, false
	}

	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)

		owner, _ := control.Get("player")
		if owner != playerID {
			continue
		}

		location, found := locations.Get(entity)
		if !found {
			return 0, false
		}

		level, _ := location.Get("level_id")

		return int(level.(int64)), true
	}

	return 0, false
}

// transitionBootstrapData copies the generated seam into Lua's bootstrap shape. Go does not
// reinterpret coordinates, preserving the generator as the single geometry authority.
func (app *application) transitionBootstrapData() map[string]any {
	app.worldMu.RLock()
	seam := app.transitionSeam
	app.worldMu.RUnlock()

	return entryworld.TransitionData(seam)
}

// warpBootstrapData creates paired portals solely for the named acceptance scene. Production scenes
// receive the same empty schema shape and cannot accidentally enable fixture traversal.
func (app *application) warpBootstrapData() map[string]any {
	if app.options.StartScene != "warp_lab" {
		return emptyWarpBootstrap()
	}

	townLevel := app.transitionSeam.Town.LevelID
	wildernessLevel := app.transitionSeam.Wilderness.LevelID
	town := app.gameWorlds[townLevel]

	wilderness := app.gameWorlds[wildernessLevel]
	if town == nil || wilderness == nil {
		return emptyWarpBootstrap()
	}

	townSpawn := app.gameWorldSpawns[townLevel]
	wildernessSpawn := app.gameWorldSpawns[wildernessLevel]
	townPortal := openWarpPoint(town, townSpawn, 7, 0)
	wildernessPortal := openWarpPoint(wilderness, wildernessSpawn, 7, 0)
	townArrival := openWarpPoint(town, townPortal, -4, 1)
	wildernessArrival := openWarpPoint(wilderness, wildernessPortal, 4, 1)

	return map[string]any{
		"endpoints": []any{
			app.warpEndpointData(
				"warp-lab:town",
				"warp-lab:wilderness",
				"TP",
				"BLUE TOWN WARP",
				townLevel,
				townPortal,
				wildernessArrival,
				wildernessLevel,
				wilderness,
			),
			app.warpEndpointData(
				"warp-lab:wilderness",
				"warp-lab:town",
				"PP",
				"RED WILDERNESS WARP",
				wildernessLevel,
				wildernessPortal,
				townArrival,
				townLevel,
				town,
			),
		},
	}
}

// emptyWarpBootstrap preserves the Lua schema even when the fixture is disabled, eliminating nil and
// missing-field branches from gameplay bootstrap code.
func emptyWarpBootstrap() map[string]any {
	return map[string]any{"endpoints": []any{}}
}

// openWarpPoint prefers a nearby collision-safe fixture position but falls back to the already valid
// generated anchor when no requested offset can fit.
func openWarpPoint(
	worldMap *gameworld.Map,
	anchor [2]float64,
	offset float64,
	radius float64,
) [2]float64 {
	x, y, found := worldMap.OpenPointNearSubtileForRadius(
		anchor[0]+offset,
		anchor[1],
		radius,
	)
	if !found {
		return anchor
	}

	return [2]float64{x, y}
}

// warpEndpointData includes destination bounds from trusted maps and adds room residency only when
// the generated zone can resolve it. Lua receives no guessed room identity.
func (app *application) warpEndpointData(
	id string,
	pairID string,
	token string,
	label string,
	levelID int,
	position [2]float64,
	destination [2]float64,
	destinationLevel int,
	destinationMap *gameworld.Map,
) map[string]any {
	result := map[string]any{
		"id":                 id,
		"pair_id":            pairID,
		"token":              token,
		"label":              label,
		"level_id":           levelID,
		"x":                  position[0],
		"y":                  position[1],
		"radius":             float64(3.5),
		"destination_level":  destinationLevel,
		"destination_x":      destination[0],
		"destination_y":      destination[1],
		"destination_width":  float64(destinationMap.WidthSubtiles),
		"destination_height": float64(destinationMap.HeightSubtiles),
	}

	roomID, found := entryworld.RoomIDAt(
		app.gameWorldZones[levelID],
		position[0],
		position[1],
	)
	if found {
		result["resident_id"] = id
		result["room_id"] = roomID
	}

	return result
}
