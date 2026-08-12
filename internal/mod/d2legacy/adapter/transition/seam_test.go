package transition_test

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

func TestTownMoorSeamUsesAuthoredTownExitAndGeneratedMoorEdge(t *testing.T) {
	town, _ := gameworld.NewOpenMap(40, 40)
	town.SpecialTiles = []gameworld.SpecialTile{{X: 7, Y: 4, Orientation: 10}}
	moor, _ := gameworld.NewOpenMap(40, 40)
	seam, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{FirstLevel: 1, FirstDirection: "east", SecondLevel: 2, SecondDirection: "west", SecondTileX: 0, SecondTileY: 4}, town, moor)
	if err != nil {
		t.Fatal(err)
	}
	if seam.Town.LevelID != 1 || seam.Wilderness.LevelID != 2 || seam.Town.Direction != "east" || seam.Wilderness.Direction != "west" {
		t.Fatalf("seam = %#v", seam)
	}
}
