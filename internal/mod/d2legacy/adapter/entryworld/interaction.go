package entryworld

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// InteractionData combines the fixed Act I guide with selectable DS1 objects in deterministic level order. Attaching
// generated room identity here lets later resident and interaction systems refer to the same authoritative location.
func InteractionData(
	worlds map[int]*gameworld.Map,
	zones map[int]*worldgen.Zone,
	owner string,
	initial string,
) map[string]any {
	targets := []any{actOneGuideTarget()}

	for _, levelID := range sortedLevelIDs(worlds) {
		targets = appendSelectableTargets(targets, levelID, worlds[levelID], zones[levelID])
	}

	return map[string]any{
		"owner":          owner,
		"initial_target": initial,
		"targets":        targets,
	}
}

// actOneGuideTarget returns the compatibility bootstrap target that exists before DS1-derived residents are loaded.
func actOneGuideTarget() map[string]any {
	return map[string]any{
		"id":         "act1-akara",
		"npc":        "Akara",
		"vendor":     "Akara",
		"categories": "armo,misc,weap",
		"services":   "",
		"x":          float64(4096),
		"y":          float64(4096),
		"radius":     float64(160),
	}
}

// sortedLevelIDs makes target publication independent of Go map iteration, preserving stable initial-state snapshots
// and checksums across processes.
func sortedLevelIDs(worlds map[int]*gameworld.Map) []int {
	levels := make([]int, 0, len(worlds))
	for levelID := range worlds {
		levels = append(levels, levelID)
	}

	sort.Ints(levels)

	return levels
}

// appendSelectableTargets enriches renderer-neutral selectable points with recovered object names and generated room
// identity. Selectables without a usable name remain absent because Lua cannot present them meaningfully.
func appendSelectableTargets(
	targets []any,
	levelID int,
	gameMap *gameworld.Map,
	zone *worldgen.Zone,
) []any {
	objects := objectsBySelectableID(gameMap)

	for _, selected := range gameMap.Selectables() {
		object := objects[selected.ID]

		name := strings.TrimSpace(object.Description)
		if name == "" {
			name = strings.TrimSpace(object.Class)
		}

		if name == "" {
			continue
		}

		target := selectableTarget(selected, name)
		if roomID, found := RoomIDAt(zone, selected.X, selected.Y); found {
			target["resident_id"] = fmt.Sprintf("level:%d:%s", levelID, selected.ID)
			target["level_id"] = float64(levelID)
			target["room_id"] = roomID
		}

		targets = append(targets, target)
	}

	return targets
}

// objectsBySelectableID rebuilds the materializer's stable selectable identity so interaction metadata can recover the
// corresponding decoded object without coupling to renderer state.
func objectsBySelectableID(gameMap *gameworld.Map) map[string]gameworld.Object {
	objects := make(map[string]gameworld.Object, len(gameMap.Objects))
	for index, object := range gameMap.Objects {
		id := fmt.Sprintf("ds1-object:%d:%d:%d", object.Type, object.ID, index)
		objects[id] = object
	}

	return objects
}

// selectableTarget creates the stable interaction fields shared by all DS1-derived targets. Room-specific identity is
// added separately only when the point falls inside a generated room.
func selectableTarget(selected gameworld.Selectable, name string) map[string]any {
	return map[string]any{
		"id":         selected.ID,
		"npc":        name,
		"vendor":     "",
		"categories": "",
		"services":   "",
		"x":          selected.X,
		"y":          selected.Y,
		"radius":     float64(4),
	}
}

// RoomIDAt resolves one authoritative subtile point against a generated zone. Bounds are half-open so a point on a
// shared edge belongs to exactly one room, and the canonical decimal ID is shared by population and resident systems.
func RoomIDAt(zone *worldgen.Zone, x, y float64) (string, bool) {
	if zone == nil {
		return "", false
	}

	for _, room := range zone.Rooms() {
		left := float64(room.X * gameworld.SubtilesPerTile)
		top := float64(room.Y * gameworld.SubtilesPerTile)
		right := left + float64(room.Width*gameworld.SubtilesPerTile)
		bottom := top + float64(room.Height*gameworld.SubtilesPerTile)

		if x >= left && x < right && y >= top && y < bottom {
			return strconv.FormatUint(uint64(room.ID), 10), true
		}
	}

	return "", false
}
