package entryworld

import (
	"math"
	"sort"
	"strconv"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// PopulationData converts prepared combat areas into deterministic room, point, and adjacency facts. Nil results
// signal that no complete wilderness exists, allowing composition roots to reject bootstrap before command admission.
func (world *Prepared) PopulationData(nearby int) map[string]any {
	zone, gameMap, found := world.wilderness()
	if !found {
		return nil
	}

	result := world.populationZone(zone, gameMap, nearby)
	additional := []any{}
	for _, levelID := range sortedLevelIDs(world.Worlds) {
		if levelID == world.Seam.Town.LevelID || levelID == zone.Request().LevelID {
			continue
		}
		other := world.Zones[levelID]
		if other == nil {
			return nil
		}
		additional = append(additional, world.populationZone(other, world.Worlds[levelID], 0))
	}
	if len(additional) > 0 {
		result["additional_zones"] = additional
	}
	return result
}

func (world *Prepared) populationZone(zone *worldgen.Zone, gameMap *gameworld.Map, nearby int) map[string]any {
	request := zone.Request()
	player := world.Spawns[request.LevelID]
	populated := populatedStamps(zone)
	var reachable map[gameworld.Point]bool
	if zone.Kind() == "maze" {
		reachable = reachablePopulationCells(gameMap, player)
	}

	rooms := make([]any, 0, len(zone.Rooms()))
	for _, room := range zone.Rooms() {
		anchors := populationAnchors(room, player, nearby)
		points := openPopulationPoints(gameMap, anchors)

		facts := populationRoom(room, populated[room.StampID], points)
		if reachable != nil {
			points, tiles := mazePopulationPoints(gameMap, room, reachable)
			facts["points"], facts["eligible_tiles"] = points, float64(tiles)
		}
		rooms = append(rooms, facts)
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

// reachablePopulationCells excludes isolated open pockets using the player's collision footprint.
func reachablePopulationCells(gameMap *gameworld.Map, spawn [2]float64) map[gameworld.Point]bool {
	start := gameworld.Point{X: float64(gameworld.CollisionCell(spawn[0])), Y: float64(gameworld.CollisionCell(spawn[1]))}
	reachable := map[gameworld.Point]bool{}
	if !gameMap.WalkableStep(start, start, 1) {
		return reachable
	}
	queue := []gameworld.Point{start}
	reachable[start] = true
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		for dy := -1.0; dy <= 1; dy++ {
			for dx := -1.0; dx <= 1; dx++ {
				next := gameworld.Point{X: current.X + dx, Y: current.Y + dy}
				if !reachable[next] && gameMap.WalkableStep(current, next, 1) {
					reachable[next] = true
					queue = append(queue, next)
				}
			}
		}
	}
	return reachable
}

// mazePopulationPoints counts usable authored tiles and separates pack members within their chamber.
func mazePopulationPoints(gameMap *gameworld.Map, room worldgen.Room, reachable map[gameworld.Point]bool) ([]any, int) {
	candidates := []gameworld.Point{}
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			point := gameworld.Point{X: float64(x*gameworld.SubtilesPerTile + 2), Y: float64(y*gameworld.SubtilesPerTile + 2)}
			if reachable[point] && gameMap.WalkableStep(point, point, 2) {
				candidates = append(candidates, point)
			}
		}
	}
	count := len(candidates)
	centerX, centerY := float64((room.X+room.Width/2)*gameworld.SubtilesPerTile), float64((room.Y+room.Height/2)*gameworld.SubtilesPerTile)
	sort.SliceStable(candidates, func(i, j int) bool {
		ix, iy := candidates[i].X-centerX, candidates[i].Y-centerY
		jx, jy := candidates[j].X-centerX, candidates[j].Y-centerY
		return ix*ix+iy*iy < jx*jx+jy*jy
	})
	selected := []gameworld.Point{}
	points := []any{}
	for _, candidate := range candidates {
		separated := true
		for _, previous := range selected {
			if math.Hypot(candidate.X-previous.X, candidate.Y-previous.Y) < 6 {
				separated = false
				break
			}
		}
		if !separated {
			continue
		}
		selected = append(selected, candidate)
		points = append(points, map[string]any{"x": candidate.X, "y": candidate.Y, "clearance": float64(2)})
		if len(points) == 4 {
			break
		}
	}
	return points, count
}
