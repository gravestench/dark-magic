package world

import (
	"strings"
	"testing"

	models "github.com/gravestench/dark-magic/internal/game/data/model"
)

func TestResolveLevelTransitionsJoinsMatchingVisibilitySlot(t *testing.T) {
	m := &Map{SpecialTiles: []SpecialTile{
		{X: 8, Y: 4, Orientation: 10, MainIndex: 3, SubIndex: 22, Hidden: true},
		{X: 1, Y: 1, Orientation: 10, MainIndex: 30}, // town entry, not Vis#
		{X: 2, Y: 2, Orientation: 10, MainIndex: 1},  // unused Vis# slot
		{X: 3, Y: 3, Orientation: 10, MainIndex: 7},  // visible relationship without warp geometry
	}}
	level := models.LevelData{Id: 4, Vis3: 9, Warp3: 13, Vis7: 10, Warp7: -1}
	warp := models.LevelWarp{Id: 13, SelectX: -2, SelectY: 3, SelectDX: 10, SelectDY: 12, ExitWalkX: 4, ExitWalkY: 5, OffsetX: 6, OffsetY: 7, LitVersion: 1, Tiles: 2, NoInteract: 1, Direction: "b", UniqueId: 99}
	resolved, err := m.ResolveLevelTransitions(level, map[int]models.LevelWarp{13: warp})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 1 {
		t.Fatalf("resolved = %#v", resolved)
	}
	got := resolved[0]
	if got.DestinationLevel != 9 || got.WarpID != 13 || got.Tile.SubIndex != 22 || got.SelectX != -2 || got.ExitWalkY != 5 || got.OffsetX != 6 || !got.LitVersion || !got.NoInteract || got.Direction != "b" || got.UniqueID != 99 {
		t.Fatalf("transition = %#v", got)
	}
	geometry := got.Geometry()
	want := WarpGeometry{
		CellOrigin: SubtilePoint{X: 40, Y: 20}, EntityPosition: SubtilePoint{X: 46, Y: 27},
		SelectionLocal: LocalSelectionBounds{MinX: -2, MinY: 3, MaxX: 8, MaxY: 15},
		Arrival:        SubtilePoint{X: 46, Y: 27}, ExitWalkTarget: SubtilePoint{X: 50, Y: 32},
	}
	if geometry != want {
		t.Fatalf("geometry = %#v, want %#v", geometry, want)
	}
}

func TestResolveLevelTransitionsRejectsMissingTypedWarp(t *testing.T) {
	m := &Map{SpecialTiles: []SpecialTile{{Orientation: 11, MainIndex: 0}}}
	_, err := m.ResolveLevelTransitions(models.LevelData{Id: 4, Vis0: 5, Warp0: 12}, nil)
	if err == nil || !strings.Contains(err.Error(), "missing LvlWarp 12") {
		t.Fatalf("error = %v", err)
	}
}
