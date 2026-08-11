package world

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
)

func TestTownMoorSeamUsesAuthoredTownExitAndGeneratedMoorEdge(t *testing.T) {
	townZone, _ := mapgen.NewZone(mapgen.Definition{Request: mapgen.Request{Version: 1, Seed: 1, Act: 1, LevelID: 1}, Kind: mapgen.Preset, Bounds: mapgen.Bounds{Width: 8, Height: 8}, Stamps: []mapgen.Stamp{{ID: 1, Role: "act1-town:exit-east", Width: 8, Height: 8, DS1Path: "town.ds1"}}, Rooms: []mapgen.Room{{ID: 1, Width: 8, Height: 8, StampID: 1}}})
	moorZone, _ := mapgen.NewZone(mapgen.Definition{Request: mapgen.Request{Version: 1, Seed: 1, Act: 1, LevelID: 2}, Kind: mapgen.Outdoor, Bounds: mapgen.Bounds{Width: 8, Height: 8}, Stamps: []mapgen.Stamp{{ID: 1, Width: 8, Height: 8, DS1Path: "moor.ds1"}}, Rooms: []mapgen.Room{{ID: 1, Width: 8, Height: 8, StampID: 1}}, Warps: []mapgen.Warp{{ID: 1, Role: "town-entry", Direction: "west", X: 0, Y: 4, DestinationLevel: 1}, {ID: 2, Role: "next-level-exit", Direction: "east", X: 7, Y: 4, DestinationLevel: 3}}})
	town := &Map{WidthSubtiles: 40, HeightSubtiles: 40, flags: make([]Flags, 1600), Tiles: []TilePlacement{{X: 7, Y: 4, Identity: TileIdentity{Orientation: 10}}}}
	moor := &Map{WidthSubtiles: 40, HeightSubtiles: 40, flags: make([]Flags, 1600)}
	seam, err := NewActOneTownMoorSeam(townZone, town, moorZone, moor)
	if err != nil {
		t.Fatal(err)
	}
	if seam.Town.LevelID != 1 || seam.Wilderness.LevelID != 2 || seam.Town.Direction != "east" || seam.Wilderness.Direction != "west" {
		t.Fatalf("seam = %#v", seam)
	}
}
