package world

import (
	"testing"

	"github.com/gravestench/ds1"
)

// TestMaterializeUsesInMemoryDS1Document proves unsaved edits can enter the world pipeline without a temporary file.
func TestMaterializeUsesInMemoryDS1Document(t *testing.T) {
	stamp := &ds1.DS1{
		Version:              ds1.LatestVersion,
		Width:                1,
		Height:               1,
		Act:                  1,
		NumberOfWalls:        1,
		NumberOfFloors:       1,
		NumberOfShadowLayers: 1,
		Tiles: [][]ds1.TileRecord{{{
			Walls:   []ds1.WallRecord{{}},
			Floors:  []ds1.FloorShadowRecord{{Prop1: 2, Style: 3, Sequence: 4}},
			Shadows: []ds1.FloorShadowRecord{{}},
		}}},
	}
	catalog := NewTileCatalog([]TileReference{{
		Identity: TileIdentity{MainIndex: 3, SubIndex: 4},
		Path:     "floor.dt1",
		Index:    0,
		Rarity:   1,
	}})
	world, err := Materialize(stamp, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(world.Tiles) != 1 || world.Tiles[0].Identity != (TileIdentity{MainIndex: 3, SubIndex: 4}) {
		t.Fatalf("placements = %#v", world.Tiles)
	}
}
