package world

import (
	"testing"

	mapgen "github.com/gravestench/dark-magic/internal/game/worldgen"
	"github.com/gravestench/dt1"
)

func TestActOneDirtPathReplacesFloorAndCollisionTogether(t *testing.T) {
	oldFlags := [25]dt1.SubTileFlags{{BlockWalk: true}}
	old := TileReference{Identity: TileIdentity{MainIndex: 9, SubIndex: 9}, Path: "fill.dt1", Index: 1, SubTileFlags: oldFlags}
	// A three-cell horizontal path gives its center east+west neighbors. The
	// legacy lookup maps that mask to sequence 0x07.
	roadIdentity := TileIdentity{MainIndex: 0, SubIndex: 0x07}
	road := TileReference{Identity: roadIdentity, Path: "road.dt1", Index: 2}
	westEnd := TileReference{Identity: TileIdentity{MainIndex: 0, SubIndex: 0x0D}, Path: "road.dt1", Index: 3}
	eastEnd := TileReference{Identity: TileIdentity{MainIndex: 0, SubIndex: 0x10}, Path: "road.dt1", Index: 4}
	m := &Map{WidthTiles: 3, HeightTiles: 1, WidthSubtiles: 15, HeightSubtiles: 5, Tiles: []TilePlacement{
		{X: 1, Y: 0, Layer: LayerFloor, Identity: old.Identity, Reference: old},
	}}
	m.flags = make([]Flags, 75)
	m.apply(1, 0, old)
	paths := []mapgen.PathTile{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}
	if err := m.realizeActOneDirtPath(paths, []*TileCatalog{NewTileCatalog([]TileReference{road, westEnd, eastEnd})}); err != nil {
		t.Fatal(err)
	}
	if got := m.Tiles[0].Identity; got != roadIdentity {
		t.Fatalf("floor identity = %#v, want %#v", got, roadIdentity)
	}
	if flags, _ := m.FlagsAt(5, 4); flags.Blocked() {
		t.Fatal("collision retained the replaced fill floor")
	}
}

func TestActOneDirtPathNeighborMaskUsesLegacyDirectionOrder(t *testing.T) {
	path := map[[2]int]bool{{0, 0}: true, {1, 0}: true, {-1, 0}: true}
	if got := pathNeighborMask(path, 0, 0); got != 0b01000010 {
		t.Fatalf("east/west mask = %08b", got)
	}
	if got := actOneDirtPathSequence[pathNeighborMask(path, 0, 0)]; got != 0x07 {
		t.Fatalf("east/west dirt sequence = %#x, want 0x07", got)
	}
}
