package entryworld

import (
	"fmt"
	"strconv"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// PopulationData publishes geometry from each admitted wilderness in seam order.
// A missing keyed map or mismatched request aborts the entire bootstrap.
func (world *Prepared) PopulationData(nearby int) (map[string]any, error) {
	if world == nil || len(world.Seams) == 0 {
		return nil, fmt.Errorf("population requires prepared seams")
	}
	zones := make([]any, 0, len(world.Seams))
	seen := map[int]bool{}
	for _, seam := range world.Seams {
		levelID := seam.Second.LevelID
		zone, gameMap := world.Zones[levelID], world.Worlds[levelID]
		if seen[levelID] || zone == nil || gameMap == nil {
			return nil, fmt.Errorf("population level %d is duplicate or incomplete", levelID)
		}
		seen[levelID] = true
		request := zone.Request()
		if request.LevelID != levelID {
			return nil, fmt.Errorf("population keyed level %d differs from zone request %d", levelID, request.LevelID)
		}
		populated := populatedStamps(zone)
		rooms := make([]any, 0, len(zone.Rooms()))
		for _, room := range zone.Rooms() {
			var points []any
			if levelID == 3 || levelID == 4 {
				points = roomPopulationPoints(gameMap, room)
			} else {
				points = openPopulationPoints(gameMap, populationAnchors(room, world.Spawns[levelID], nearby))
			}
			rooms = append(rooms, populationRoom(room, populated[room.StampID], points))
		}
		zones = append(zones, map[string]any{
			"seed":       float64(request.Seed),
			"act":        float64(request.Act),
			"level_id":   float64(levelID),
			"difficulty": float64(request.Difficulty),
			"rooms":      rooms,
			"links":      populationLinks(zone.Links()),
		})
	}
	return map[string]any{"zones": zones}, nil
}

// roomPopulationPoints searches only this room, retaining distinct center cells
// in deterministic rings around its center. Lua checks the selected footprint.
func roomPopulationPoints(gameMap *gameworld.Map, room worldgen.Room) []any {
	points := make([]any, 0, 4)
	minX, minY := room.X*gameworld.SubtilesPerTile, room.Y*gameworld.SubtilesPerTile
	maxX, maxY := minX+room.Width*gameworld.SubtilesPerTile, minY+room.Height*gameworld.SubtilesPerTile
	centerX, centerY := (minX+maxX)/2, (minY+maxY)/2
	appendPoint := func(x, y int) {
		if len(points) >= 4 || x < minX || x >= maxX || y < minY || y >= maxY {
			return
		}
		if flags, ok := gameMap.FlagsAt(x, y); ok && !flags.Blocked() {
			points = append(points, map[string]any{"x": float64(x), "y": float64(y)})
		}
	}
	// Visit each ring's perimeter once, even when the whole room is blocked.
	limit := max(maxX-minX, maxY-minY)
	for radius := 0; radius <= limit && len(points) < 4; radius++ {
		left, right := centerX-radius, centerX+radius
		top, bottom := centerY-radius, centerY+radius
		for y := max(minY, top); y < min(maxY, bottom+1) && len(points) < 4; y++ {
			if y == top || y == bottom {
				for x := max(minX, left); x < min(maxX, right+1) && len(points) < 4; x++ {
					appendPoint(x, y)
				}
			} else {
				appendPoint(left, y)
				if right != left {
					appendPoint(right, y)
				}
			}
		}
	}
	return points
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
