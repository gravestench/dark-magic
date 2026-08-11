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

	m.apply(1, 0, TileReference{SubTileFlags: first.SubTileFlags})
	m.apply(1, 0, TileReference{SubTileFlags: second.SubTileFlags})

	got, ok := m.FlagsAt(5, 4)
	if !ok || !got.BlockLOS || !got.BlockPlayerWalk || !got.Blocked() {
		t.Fatalf("FlagsAt(5, 4) = %#v, %v", got, ok)
	}
	got, ok = m.FlagsAt(9, 0)
	if !ok || !got.BlockWalk || !got.Blocked() {
		t.Fatalf("FlagsAt(9, 0) = %#v, %v", got, ok)
	}
}

func TestApplyUsesLegacyBottomToTopDT1SubtileRows(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	tile := &dt1.Tile{}
	for index := range tile.SubTileFlags {
		tile.SubTileFlags[index].BlockWalk = index == 0 || index == 24
	}
	m.apply(0, 0, TileReference{SubTileFlags: tile.SubTileFlags})
	for _, point := range [][2]int{{0, 4}, {4, 0}} {
		flags, ok := m.FlagsAt(point[0], point[1])
		if !ok || !flags.BlockWalk {
			t.Fatalf("legacy subtile %v = %#v, %v", point, flags, ok)
		}
	}
	for _, point := range [][2]int{{0, 0}, {4, 4}} {
		flags, _ := m.FlagsAt(point[0], point[1])
		if flags.BlockWalk {
			t.Fatalf("unrelated mirrored subtile %v is blocked", point)
		}
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

func TestApplyIgnoresTileOutsideMap(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	tile := &dt1.Tile{}
	tile.SubTileFlags[0].BlockWalk = true
	reference := TileReference{SubTileFlags: tile.SubTileFlags}
	m.apply(-1, 0, reference)
	m.apply(1, 0, reference)
}

func TestTileLayerNamesAreStablePresentationFacts(t *testing.T) {
	want := []string{"floor", "lower-wall", "shadow", "upper-wall", "roof"}
	for layer, name := range want {
		if got := TileLayer(layer).String(); got != name {
			t.Fatalf("layer %d = %q, want %q", layer, got, name)
		}
	}
	if got := TileLayer(255).String(); got != "unknown" {
		t.Fatalf("unknown layer = %q", got)
	}
}

func TestOpenPointNearCenterSkipsBlockedCenterDeterministically(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	m.flags[2*5+2].BlockWalk = true
	x, y, found := m.OpenPointNearCenter()
	if !found || x != 1 || y != 1 {
		t.Fatalf("open point = %v,%v,%v; want first perimeter point 1,1", x, y, found)
	}
}

func TestActOneTownEntryUsesAuthoredBonfireAndAvoidsItsCenter(t *testing.T) {
	m := &Map{WidthSubtiles: 20, HeightSubtiles: 20, Objects: []Object{{Type: ObjectTypeStatic, ID: 2, X: 10, Y: 10}}}
	m.flags = make([]Flags, 400)
	x, y, found := m.ActOneTownEntry()
	if !found {
		t.Fatal("expected an entry point")
	}
	if x != 6 || y != 6 {
		t.Fatalf("entry = (%v,%v), want deterministic first point at radius four", x, y)
	}
}

func TestActOneTownEntryRequiresAuthoredBonfire(t *testing.T) {
	m := &Map{WidthSubtiles: 20, HeightSubtiles: 20, flags: make([]Flags, 400)}
	if _, _, found := m.ActOneTownEntry(); found {
		t.Fatal("invented an entry without a bonfire")
	}
}
