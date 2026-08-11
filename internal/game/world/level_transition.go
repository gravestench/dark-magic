package world

import (
	"fmt"
	"sort"

	models "github.com/gravestench/dark-magic/internal/game/data/model"
)

// LevelTransition joins one raw DS1 special tile to the paired Levels.txt and
// LvlWarp.txt records. Geometry remains in its authored units until each field's
// coordinate semantics has been verified against production maps.
type LevelTransition struct {
	Tile             SpecialTile
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
	origin := SubtilePoint{X: transition.Tile.X * SubtilesPerTile, Y: transition.Tile.Y * SubtilesPerTile}
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
func (m *Map) ResolveLevelTransitions(level models.LevelData, warps map[int]models.LevelWarp) ([]LevelTransition, error) {
	if m == nil {
		return nil, nil
	}
	links := make(map[int]models.LevelLink, 8)
	for _, link := range level.Links() {
		links[link.Slot] = link
	}
	result := make([]LevelTransition, 0)
	for _, tile := range m.SpecialTiles {
		if tile.MainIndex < 0 || tile.MainIndex > 7 {
			continue
		}
		link, found := links[int(tile.MainIndex)]
		if !found || link.WarpID < 0 {
			continue
		}
		warp, found := warps[link.WarpID]
		if !found {
			return nil, fmt.Errorf("world: level %d visibility slot %d references missing LvlWarp %d", level.Id, link.Slot, link.WarpID)
		}
		result = append(result, LevelTransition{
			Tile: tile, DestinationLevel: link.DestinationLevel, WarpID: link.WarpID,
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
