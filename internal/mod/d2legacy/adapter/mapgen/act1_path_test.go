package mapgen

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	worldgen "github.com/gravestench/dark-magic/internal/game/worldgen"
	"github.com/gravestench/dt1"
)

func TestActOneDirtPathReplacesFloorAndCollisionTogether(t *testing.T) {
	oldFlags := [25]dt1.SubTileFlags{{BlockWalk: true}}
	old := gameworld.TileReference{Identity: gameworld.TileIdentity{MainIndex: 9, SubIndex: 9}, Path: "fill.dt1", Index: 1, SubTileFlags: oldFlags}
	// A three-cell horizontal path gives its center east+west neighbors. The
	// legacy lookup maps that mask to sequence 0x07.
	roadIdentity := gameworld.TileIdentity{MainIndex: 0, SubIndex: 0x07}
	road := gameworld.TileReference{Identity: roadIdentity, Path: "road.dt1", Index: 2}
	westEnd := gameworld.TileReference{Identity: gameworld.TileIdentity{MainIndex: 0, SubIndex: 0x0D}, Path: "road.dt1", Index: 3}
	eastEnd := gameworld.TileReference{Identity: gameworld.TileIdentity{MainIndex: 0, SubIndex: 0x10}, Path: "road.dt1", Index: 4}
	m, _ := gameworld.NewOpenMap(15, 5)
	m.Tiles = []gameworld.TilePlacement{{X: 1, Y: 0, Layer: gameworld.LayerFloor, Identity: old.Identity, Reference: old}}
	m.RebuildFlags()
	zone, _ := worldgen.NewZone(worldgen.Definition{Request: worldgen.Request{Version: 1, Act: 1, LevelID: 2}, Kind: "outdoor", Bounds: worldgen.Bounds{Width: 3, Height: 1}, Stamps: []worldgen.Stamp{{ID: 1, Width: 3, Height: 1, DS1Path: "fill.ds1"}}, Rooms: []worldgen.Room{{ID: 1, Width: 3, Height: 1, StampID: 1}}, Paths: []worldgen.PathTile{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}})
	if err := RealizeActOneDirtPath(m, zone, []*gameworld.TileCatalog{gameworld.NewTileCatalog([]gameworld.TileReference{road, westEnd, eastEnd})}); err != nil {
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
