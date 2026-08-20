package modruntime

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mapeditor"
)

// TestAutoDrawLearnsSameTilesetNeighborPattern keeps learned transitions local to the selected DT1 family.
func TestAutoDrawLearnsSameTilesetNeighborPattern(t *testing.T) {
	document, err := mapeditor.New(mapeditor.NewConfig{Width: 4, Height: 1})
	if err != nil {
		t.Fatal(err)
	}
	a := mapeditor.Brush{Identity: mapeditor.Identity{Style: 1}}
	b := mapeditor.Brush{Identity: mapeditor.Identity{Style: 2}}
	if err := document.BeginStroke(mapeditor.LayerFloor, 0, a); err != nil {
		t.Fatal(err)
	}
	for x, brush := range []mapeditor.Brush{a, b, a, b} {
		if _, err := document.Paint(mapeditor.Point{X: x}, brush); err != nil {
			t.Fatal(err)
		}
	}
	document.EndStroke()

	session := &mapEditorSession{
		document: document,
		catalog: world.NewTileCatalog([]world.TileReference{
			{Identity: world.TileIdentity{MainIndex: 1}, Path: "terrain.dt1"},
			{Identity: world.TileIdentity{MainIndex: 2}, Path: "terrain.dt1"},
		}),
	}
	model := newAutoDrawModel(session, mapeditor.LayerFloor, 0, "terrain.dt1")
	if model == nil {
		t.Fatal("auto draw model was not built")
	}
	got := model.choose(document, mapeditor.Point{X: 1}, a)
	if got.Identity.Style != 2 {
		t.Fatalf("learned choice style = %d, want 2", got.Identity.Style)
	}
}
