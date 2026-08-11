package world

import "testing"

func TestAuthoredExitAnchorsUseSemanticOrientations(t *testing.T) {
	m := &Map{WidthSubtiles: 20, HeightSubtiles: 20, flags: make([]Flags, 400), SpecialTiles: []SpecialTile{
		{X: 3, Y: 1, Orientation: 11}, {X: 1, Y: 3, Orientation: 10}, {Orientation: 2},
	}}
	anchors := m.AuthoredExitAnchors()
	if len(anchors) != 2 || anchors[0].Orientation != 11 || anchors[1].Orientation != 10 {
		t.Fatalf("anchors = %#v", anchors)
	}
	x, y, found := m.OpenPointNearExit(anchors[0])
	if !found || x != 18 || y != 8 {
		t.Fatalf("open point = %v,%v,%v", x, y, found)
	}
}
