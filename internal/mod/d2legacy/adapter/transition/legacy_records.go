package transition

import (
	"fmt"
	"sort"

	models "github.com/gravestench/dark-magic/internal/game/data/model"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

type levelLink struct {
	slot             int
	destinationLevel int
	warpID           int
}

// levelLinks interprets Levels.txt's eight Vis#/Warp# column pairs. Keeping
// this join here is important: the schema layer preserves columns, while the
// d2legacy adapter decides that they describe Diablo level transitions.
func levelLinks(level models.LevelData) []levelLink {
	destinations := [...]int{level.Vis0, level.Vis1, level.Vis2, level.Vis3, level.Vis4, level.Vis5, level.Vis6, level.Vis7}
	warps := [...]int{level.Warp0, level.Warp1, level.Warp2, level.Warp3, level.Warp4, level.Warp5, level.Warp6, level.Warp7}
	result := make([]levelLink, 0, len(destinations))
	for slot, destination := range destinations {
		if destination > 0 {
			result = append(result, levelLink{slot: slot, destinationLevel: destination, warpID: warps[slot]})
		}
	}
	return result
}

// LevelTransition joins one raw DS1 special tile to the paired Levels.txt and
// LvlWarp.txt records. Geometry remains in its authored units until each field's
// coordinate semantics has been verified against production maps.
type LevelTransition struct {
	Tile             gameworld.SpecialTile
	DestinationLevel int
	WarpID           int
	SelectX          int
	SelectY          int
	SelectDX         int
	SelectDY         int
	ExitWalkX        int
	ExitWalkY        int
	OffsetX          int
	OffsetY          int
	LitVersion       bool
	Tiles            int
	NoInteract       bool
	Direction        string
	UniqueID         int
}

// SubtilePoint is one authoritative point in the map's 5x5-per-tile gameplay
// coordinate system.
type SubtilePoint struct{ X, Y int }

// LocalSelectionBounds preserves LvlWarp's client-side selection rectangle in
// authored local units. Production values prove these are not world subtiles;
// conversion to screen hit-test space belongs to presentation.
type LocalSelectionBounds struct{ MinX, MinY, MaxX, MaxY int }

// WarpGeometry separates the four positions that are easy to accidentally
// collapse into one magic coordinate.
type WarpGeometry struct {
	CellOrigin     SubtilePoint
	EntityPosition SubtilePoint
	SelectionLocal LocalSelectionBounds
	Arrival        SubtilePoint
	ExitWalkTarget SubtilePoint
}

// Geometry applies the legacy subtile arithmetic recovered by Riiablo:
//
//   - a DS1 special cell begins at tile coordinate * 5;
//   - Offset moves the warp entity from that cell origin;
//   - Select remains an untransformed client-local rectangle;
//   - arrival starts at the destination entity and ExitWalk is the follow-up
//     automatic movement target.
func (transition LevelTransition) Geometry() WarpGeometry {
	origin := SubtilePoint{X: transition.Tile.X * gameworld.SubtilesPerTile, Y: transition.Tile.Y * gameworld.SubtilesPerTile}
	entity := SubtilePoint{X: origin.X + transition.OffsetX, Y: origin.Y + transition.OffsetY}
	selection := LocalSelectionBounds{
		MinX: transition.SelectX,
		MinY: transition.SelectY,
		MaxX: transition.SelectX + transition.SelectDX,
		MaxY: transition.SelectY + transition.SelectDY,
	}
	return WarpGeometry{
		CellOrigin: origin, EntityPosition: entity, SelectionLocal: selection, Arrival: entity,
		ExitWalkTarget: SubtilePoint{X: entity.X + transition.ExitWalkX, Y: entity.Y + transition.ExitWalkY},
	}
}

// ResolveLevelTransitions resolves ordinary visibility special tiles. Main
// indexes 0 through 7 select the matching Levels.txt Vis#/Warp# pair. Higher
// indexes are other authored markers (entry, corpse, portal, and level-specific
// facts), so they are intentionally left unresolved here.
func ResolveLevelTransitions(m *gameworld.Map, level models.LevelData, warps map[int]models.LevelWarp) ([]LevelTransition, error) {
	if m == nil {
		return nil, nil
	}
	links := make(map[int]levelLink, 8)
	for _, link := range levelLinks(level) {
		links[link.slot] = link
	}
	result := make([]LevelTransition, 0)
	for _, tile := range m.SpecialTiles {
		if tile.MainIndex < 0 || tile.MainIndex > 7 {
			continue
		}
		link, found := links[int(tile.MainIndex)]
		if !found || link.warpID < 0 {
			continue
		}
		warp, found := warps[link.warpID]
		if !found {
			return nil, fmt.Errorf("world: level %d visibility slot %d references missing LvlWarp %d", level.Id, link.slot, link.warpID)
		}
		result = append(result, LevelTransition{
			Tile: tile, DestinationLevel: link.destinationLevel, WarpID: link.warpID,
			SelectX: warp.SelectX, SelectY: warp.SelectY, SelectDX: warp.SelectDX, SelectDY: warp.SelectDY,
			ExitWalkX: warp.ExitWalkX, ExitWalkY: warp.ExitWalkY, OffsetX: warp.OffsetX, OffsetY: warp.OffsetY,
			LitVersion: warp.LitVersion != 0, Tiles: warp.Tiles, NoInteract: warp.NoInteract != 0,
			Direction: warp.Direction, UniqueID: warp.UniqueId,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Tile.Y != result[j].Tile.Y {
			return result[i].Tile.Y < result[j].Tile.Y
		}
		if result[i].Tile.X != result[j].Tile.X {
			return result[i].Tile.X < result[j].Tile.X
		}
		return result[i].Tile.MainIndex < result[j].Tile.MainIndex
	})
	return result, nil
}
