package entryworld

import (
	"strconv"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// PopulationData converts the prepared wilderness into deterministic room, point, and adjacency facts. Nil results
// signal that no complete wilderness exists, allowing composition roots to reject bootstrap before command admission.
func (world *Prepared) PopulationData(nearby int) map[string]any {
	zone, gameMap, found := world.wilderness()
	if !found {
		return nil
	}

	request := zone.Request()
	player := world.Spawns[world.Seam.Wilderness.LevelID]
	populated := populatedStamps(zone)

	rooms := make([]any, 0, len(zone.Rooms()))
	for _, room := range zone.Rooms() {
		anchors := populationAnchors(room, player, nearby)
		points := openPopulationPoints(gameMap, anchors)

		rooms = append(rooms, populationRoom(room, populated[room.StampID], points))
	}

	return map[string]any{
		"seed":       float64(request.Seed),
		"act":        float64(request.Act),
		"level_id":   float64(request.LevelID),
		"difficulty": float64(request.Difficulty),
		"rooms":      rooms,
		"links":      populationLinks(zone.Links()),
	}
}

// wilderness returns the paired generated and materialized wilderness chosen by the resolved seam. Looking up both by
// the same level ID prevents topology facts from being combined with collision from another level.
func (world *Prepared) wilderness() (*worldgen.Zone, *gameworld.Map, bool) {
	if world == nil {
		return nil, nil, false
	}

	levelID := world.Seam.Wilderness.LevelID
	zone := world.Zones[levelID]
	gameMap := world.Worlds[levelID]

	return zone, gameMap, zone != nil && gameMap != nil
}

// populatedStamps indexes recipe population policy once so each room can inherit its stamp's admitted decision.
func populatedStamps(zone *worldgen.Zone) map[uint32]bool {
	populated := make(map[uint32]bool)
	for _, stamp := range zone.Stamps() {
		populated[stamp.ID] = stamp.Populate
	}

	return populated
}

// populationAnchors selects candidate points near the player for development mode or around each room center for the
// normal deterministic bootstrap. The original ordering is preserved because it influences resident placement.
func populationAnchors(room worldgen.Room, player [2]float64, nearby int) [][2]float64 {
	if nearby > 0 {
		return [][2]float64{
			{player[0] + 10, player[1]},
			{player[0] + 7, player[1] + 7},
			{player[0], player[1] + 10},
			{player[0] - 7, player[1] + 7},
		}
	}

	centerX := float64((room.X+room.Width/2)*gameworld.SubtilesPerTile) + 2
	centerY := float64((room.Y+room.Height/2)*gameworld.SubtilesPerTile) + 2

	return [][2]float64{
		{centerX, centerY},
		{centerX + 1, centerY},
		{centerX, centerY + 1},
		{centerX - 1, centerY},
	}
}

// openPopulationPoints projects candidate anchors onto collision-safe subtiles. Candidates with no nearby open point
// are omitted rather than publishing an invalid spawn that downstream policy would need to second-guess.
func openPopulationPoints(gameMap *gameworld.Map, anchors [][2]float64) []any {
	points := make([]any, 0, len(anchors))
	for _, anchor := range anchors {
		if x, y, ok := gameMap.OpenPointNearSubtile(anchor[0], anchor[1]); ok {
			points = append(points, map[string]any{"x": x, "y": y})
		}
	}

	return points
}

// populationRoom emits tile bounds as subtiles because collision and entity placement operate in subtile space.
func populationRoom(room worldgen.Room, populate bool, points []any) map[string]any {
	return map[string]any{
		"id":       strconv.FormatUint(uint64(room.ID), 10),
		"populate": populate,
		"points":   points,
		"x":        float64(room.X * gameworld.SubtilesPerTile),
		"y":        float64(room.Y * gameworld.SubtilesPerTile),
		"width":    float64(room.Width * gameworld.SubtilesPerTile),
		"height":   float64(room.Height * gameworld.SubtilesPerTile),
	}
}

// populationLinks retains generator order while converting room IDs to the canonical decimal strings used in state.
func populationLinks(links []worldgen.Link) []any {
	result := make([]any, 0, len(links))
	for _, link := range links {
		result = append(result, map[string]any{
			"from": strconv.FormatUint(uint64(link.From), 10),
			"to":   strconv.FormatUint(uint64(link.To), 10),
		})
	}

	return result
}
