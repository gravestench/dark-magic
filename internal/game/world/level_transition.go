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
