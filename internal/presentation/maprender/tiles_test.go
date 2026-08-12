package maprender

import (
	"image"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/world"
)

func TestTileSetVisibleReturnsOnlyIntersectingDraws(t *testing.T) {
	set := &TileSet{Draws: []TileDraw{
		{Graphic: 0, Bounds: image.Rect(0, 0, 40, 40), Layer: world.LayerFloor},
		{Graphic: 0, Bounds: image.Rect(80, 0, 120, 40), Layer: world.LayerFloor},
		{Graphic: 1, Bounds: image.Rect(20, 20, 60, 60), Layer: world.LayerUpperWall},
	}}
	got := set.Visible(image.Rect(30, 30, 70, 70), []int{99})
	want := []int{99, 0, 2}
	if len(got) != len(want) {
		t.Fatalf("visible indexes = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("visible indexes = %v, want %v", got, want)
		}
	}
}

func TestWallPlacementUsesBlockMinimumInsteadOfHeaderHeight(t *testing.T) {
	reference := world.TileReference{Height: -160, YAdjust: -16}
	if got := referenceYAdjust(reference, world.LayerUpperWall); got != -16 {
		t.Fatalf("wall y adjustment = %d, want block-derived -16", got)
	}
	if oldGuess := -absolute(int(reference.Height)) + world.TilePixelHeight; oldGuess == referenceYAdjust(reference, world.LayerUpperWall) {
		t.Fatal("wall placement silently regressed to DT1 header-height guess")
	}
}

func TestTileSetVisibleKeepsSharedGraphicIdentity(t *testing.T) {
	set := &TileSet{
		Graphics: []*TileGraphic{{Path: "floor.dt1", Index: 7}},
		Draws: []TileDraw{
			{Graphic: 0, Bounds: image.Rect(0, 0, 10, 10)},
			{Graphic: 0, Bounds: image.Rect(20, 0, 30, 10)},
		},
	}
	visible := set.Visible(image.Rect(0, 0, 40, 20), nil)
	if len(visible) != 2 || set.Draws[visible[0]].Graphic != set.Draws[visible[1]].Graphic {
		t.Fatalf("repeated placements lost shared graphic identity: %v", visible)
	}
}

func TestTileSetVisibleDeduplicatesDrawSpanningBuckets(t *testing.T) {
	set := &TileSet{BucketSize: 16, Draws: []TileDraw{
		{Bounds: image.Rect(8, 8, 40, 40)}, // occupies nine spatial buckets
		{Bounds: image.Rect(48, 48, 56, 56)},
	}}
	got := set.Visible(image.Rect(0, 0, 48, 48), nil)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("visible indexes = %v, want [0] without bucket duplicates", got)
	}
	if len(set.Buckets) != 10 {
		t.Fatalf("bucket count = %d, want 10", len(set.Buckets))
	}
}

func BenchmarkTileSetVisibleSpatialIndex(b *testing.B) {
	set := &TileSet{BucketSize: 512}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			left, top := x*160, y*80
			set.Draws = append(set.Draws, TileDraw{Bounds: image.Rect(left, top, left+160, top+160)})
		}
	}
	set.buildBuckets()
	view := image.Rect(7000, 3500, 8000, 4300)
	visible := make([]int, 0, 256)
	b.ReportAllocs()
	for b.Loop() {
		visible = set.Visible(view, visible[:0])
		if len(visible) == 0 {
			b.Fatal("spatial query unexpectedly found no draws")
		}
	}
}
