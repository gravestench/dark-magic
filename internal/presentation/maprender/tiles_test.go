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
