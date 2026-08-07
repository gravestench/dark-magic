package world

import (
	"testing"

	"github.com/gravestench/dt1"
)

type testObjectResolver struct{ calls int }

func (resolver *testObjectResolver) ResolveStaticObject(act, id int) (int, string, bool) {
	resolver.calls++
	if act == 1 && id == 7 {
		return 108, "Malus", true
	}
	return 0, "", false
}

func (resolver *testObjectResolver) ResolveDynamicObject(act, id int) (string, bool) {
	resolver.calls++
	if act == 1 && id == 7 {
		return "fallen", true
	}
	return "", false
}

func TestResolveObjectUsesRecoveredMappingOnlyForStaticRecords(t *testing.T) {
	resolver := &testObjectResolver{}
	static := resolveObject(1, ObjectTypeStatic, 7, 10, 20, 3, resolver)
	if !static.Resolved || static.ObjectID != 108 || static.Description != "Malus" {
		t.Fatalf("static object = %#v", static)
	}
	dynamic := resolveObject(1, ObjectTypeDynamic, 7, 10, 20, 3, resolver)
	if !dynamic.Resolved || dynamic.Class != "fallen" || resolver.calls != 2 {
		t.Fatalf("dynamic object = %#v, resolver calls = %d", dynamic, resolver.calls)
	}
}

func TestApplyCombinesSubtileFlags(t *testing.T) {
	m := &Map{WidthSubtiles: 10, HeightSubtiles: 5, flags: make([]Flags, 50)}
	first := &dt1.Tile{}
	first.SubTileFlags[0].BlockLOS = true
	first.SubTileFlags[24].BlockWalk = true
	second := &dt1.Tile{}
	second.SubTileFlags[0].BlockPlayerWalk = true

	m.apply(1, 0, first)
	m.apply(1, 0, second)

	got, ok := m.FlagsAt(5, 0)
	if !ok || !got.BlockLOS || !got.BlockPlayerWalk || !got.Blocked() {
		t.Fatalf("FlagsAt(5, 0) = %#v, %v", got, ok)
	}
	got, ok = m.FlagsAt(9, 4)
	if !ok || !got.BlockWalk || !got.Blocked() {
		t.Fatalf("FlagsAt(9, 4) = %#v, %v", got, ok)
	}
}

func TestFlagsAtRejectsOutsideMap(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	for _, point := range [][2]int{{-1, 0}, {0, -1}, {5, 0}, {0, 5}} {
		if _, ok := m.FlagsAt(point[0], point[1]); ok {
			t.Fatalf("FlagsAt(%d, %d) unexpectedly succeeded", point[0], point[1])
		}
	}
}

func TestSubtilePixelProjectionMatchesRendererAndRoundTrips(t *testing.T) {
	m := &Map{WidthTiles: 10, HeightTiles: 8}
	x, y := m.SubtileToPixel(5, 10)
	if x != 720 || y != 320 {
		t.Fatalf("projected = %v,%v", x, y)
	}
	for _, point := range [][2]float64{{0, 0}, {5, 10}, {13.25, 7.5}} {
		px, py := m.SubtileToPixel(point[0], point[1])
		sx, sy := m.PixelToSubtile(px, py)
		if sx != point[0] || sy != point[1] {
			t.Fatalf("round trip %v = %v,%v", point, sx, sy)
		}
	}
}

func TestChooseIsDeterministicAndHandlesZeroWeight(t *testing.T) {
	first := &dt1.Tile{RarityFrameIndex: 0}
	second := &dt1.Tile{RarityFrameIndex: 0}
	if got := choose([]*dt1.Tile{first, second}, 7, 11); got != first {
		t.Fatalf("zero-weight choose returned %p, want first tile %p", got, first)
	}
	first.RarityFrameIndex, second.RarityFrameIndex = 1, 10
	want := choose([]*dt1.Tile{first, second}, 7, 11)
	for range 10 {
		if got := choose([]*dt1.Tile{first, second}, 7, 11); got != want {
			t.Fatalf("choose was not deterministic: got %p, want %p", got, want)
		}
	}
}

func TestApplyIgnoresTileOutsideMap(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	tile := &dt1.Tile{}
	tile.SubTileFlags[0].BlockWalk = true
	m.apply(-1, 0, tile)
	m.apply(1, 0, tile)
}
