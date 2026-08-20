package world

import (
	"testing"

	"github.com/gravestench/ds1"
)

// TestReplaceAuthoredCellChangesOnlyRequestedPreviewCell protects local preview invalidation from full-map rebuilds.
func TestReplaceAuthoredCellChangesOnlyRequestedPreviewCell(t *testing.T) {
	stamp := &ds1.DS1{Width: 2, Height: 1, Act: 1}
	stamp.Tiles = [][]ds1.TileRecord{{
		{Floors: []ds1.FloorShadowRecord{{Prop1: 1, Style: 1}}},
		{Floors: []ds1.FloorShadowRecord{{Prop1: 1, Style: 1}}},
	}}
	first := TileIdentity{MainIndex: 1}
	second := TileIdentity{MainIndex: 2}
	catalog := NewTileCatalog([]TileReference{
		{Identity: first, Path: "floor.dt1", Index: 1},
		{Identity: second, Path: "floor.dt1", Index: 2},
	})
	preview, err := Materialize(stamp, catalog)
	if err != nil {
		t.Fatal(err)
	}
	before := preview.PresentationSnapshot()
	preview.ReplaceAuthoredCell(catalog, 0, 0, ds1.TileRecord{
		Floors: []ds1.FloorShadowRecord{{Prop1: 1, Style: 2}},
	})
	if len(preview.Tiles) != 2 || preview.Tiles[0].Identity != second || preview.Tiles[1].Identity != first {
		t.Fatalf("updated placements = %#v", preview.Tiles)
	}
	if before.Tiles[0].Identity != first || before.Tiles[1].Identity != first {
		t.Fatalf("published snapshot changed with preview authority: %#v", before.Tiles)
	}
}
