package entryworld

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	"github.com/gravestench/dt1"
)

type caveRecords map[string][]map[string]string

// Load supplies synthetic rows through the same records boundary as owned game data.
func (r caveRecords) Load(path string) ([]map[string]string, error) { return r[path], nil }

// Invalidate is inert because the fixture rows cannot change outside each test.
func (caveRecords) Invalidate(string) {}

// Loaded reports the synthetic records available without archive access.
func (caveRecords) Loaded(string) bool { return true }

// caveFixture pairs synthetic reciprocal rows with open maps carrying the exact authored marker identities.
func caveFixture(t *testing.T) (caveRecords, map[int]*worldgen.Zone, map[int]*gameworld.Map, d2mapgen.CaveWarpSpec) {
	t.Helper()
	records := caveRecords{
		"data/global/excel/Levels.txt": {{"Id": "2", "Vis3": "8", "Warp3": "0"}, {"Id": "8", "Vis0": "2", "Warp0": "4"}},
		"data/global/excel/LvlWarp.txt": {
			{"Id": "0", "OffsetX": "1", "OffsetY": "-1", "ExitWalkX": "5", "ExitWalkY": "0"},
			{"Id": "4", "OffsetX": "2", "OffsetY": "5", "ExitWalkX": "3", "ExitWalkY": "5"},
		},
	}
	zones := map[int]*worldgen.Zone{}
	worlds := map[int]*gameworld.Map{}
	for _, side := range []struct {
		level, target, id int
		role              string
		main, sub         int32
	}{{2, 8, 3, "den-entrance", 6, 24}, {8, 2, 1, "den-exit", 0, 21}} {
		zone, err := worldgen.NewZone(worldgen.Definition{
			Request: worldgen.Request{Version: worldgen.ContractVersion, Act: 1, LevelID: side.level},
			Kind:    "test", Bounds: worldgen.Bounds{Width: 24, Height: 24},
			Warps: []worldgen.Warp{{
				ID: uint32(side.id), Role: side.role, DestinationLevel: side.target,
				X: 1, Y: 1, Direction: "north",
			}},
		})
		if err != nil {
			t.Fatal(err)
		}
		zones[side.level] = zone
		m, err := gameworld.NewOpenMap(120, 120)
		if err != nil {
			t.Fatal(err)
		}
		m.SpecialTiles = []gameworld.SpecialTile{{X: 10, Y: 10, Orientation: 10, MainIndex: side.main, SubIndex: side.sub}}
		worlds[side.level] = m
	}
	spec := d2mapgen.CaveWarpSpec{
		FirstLevel: 2, SecondLevel: 8, FirstWarpRole: "den-entrance", SecondWarpRole: "den-exit",
		FirstZoneWarpID: 3, SecondZoneWarpID: 1, FirstVisSlot: 3, SecondVisSlot: 0,
		FirstLvlWarpID: 0, SecondLvlWarpID: 4,
	}
	return records, zones, worlds, spec
}

// Placeholder recipe coordinates cannot relocate the interaction anchor.
func TestDenCaveWarpUsesFullMarkerAndTypedOffsets(t *testing.T) {
	records, zones, worlds, spec := caveFixture(t)
	pair, err := resolveCaveWarp(records, zones, worlds, spec)
	if err != nil {
		t.Fatal(err)
	}
	if pair.First.ID == pair.Second.ID || pair.First.PairedID != pair.Second.ID || pair.Second.PairedID != pair.First.ID {
		t.Fatal("unpaired endpoint IDs")
	}
	if pair.First.LevelID != 2 || pair.First.TargetLevel != 8 || pair.Second.LevelID != 8 || pair.Second.TargetLevel != 2 {
		t.Fatal("wrong targets")
	}
	if pair.First.X != 51 || pair.First.Y != 49 || pair.Second.X != 52 || pair.Second.Y != 55 {
		t.Fatalf("source points did not use markers plus typed offsets: %#v", pair)
	}
	if pair.First.ArrivalX != 55 || pair.First.ArrivalY != 60 || pair.Second.ArrivalX != 56 || pair.Second.ArrivalY != 49 {
		t.Fatalf("arrival points did not use ExitWalk: %#v", pair)
	}
	if pair.First.Width != 120 || pair.First.Height != 120 {
		t.Fatal("wrong destination dimensions")
	}
	// Another orientation-10 meaning must not create an ambiguous warp.
	worlds[2].SpecialTiles = append(worlds[2].SpecialTiles, gameworld.SpecialTile{
		X: 2, Y: 2, Orientation: 10, MainIndex: 0, SubIndex: 24,
	})
	again, err := resolveCaveWarp(records, zones, worlds, spec)
	if err != nil || again != pair {
		t.Fatalf("unrelated marker changed pair: %v", err)
	}
}

// blockCaveTile makes a synthetic map cell fail player-sized cave approach checks.
func blockCaveTile(m *gameworld.Map, x, y int) {
	var flags [25]dt1.SubTileFlags
	for i := range flags {
		flags[i].BlockWalk = true
	}
	m.Tiles = append(m.Tiles, gameworld.TilePlacement{
		X: x, Y: y, Layer: gameworld.LayerFloor,
		Reference: gameworld.TileReference{Path: "synthetic", SubTileFlags: flags},
	})
}

// Invalid records, ambiguous markers, and disconnected geometry fail atomically.
func TestDenCaveWarpRejectsUntrustedPair(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(caveRecords, map[int]*worldgen.Zone, map[int]*gameworld.Map, *d2mapgen.CaveWarpSpec)
	}{
		{"missing marker", func(_ caveRecords, _ map[int]*worldgen.Zone, m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			m[2].SpecialTiles = nil
		}},
		{"duplicate marker", func(
			_ caveRecords, _ map[int]*worldgen.Zone,
			m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			m[8].SpecialTiles = append(m[8].SpecialTiles, m[8].SpecialTiles[0])
		}},
		{"wrong main", func(_ caveRecords, _ map[int]*worldgen.Zone, m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			m[2].SpecialTiles[0].MainIndex = 0
		}},
		{"wrong orientation", func(
			_ caveRecords, _ map[int]*worldgen.Zone,
			m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			m[8].SpecialTiles[0].Orientation = 11
		}},
		{"wrong sub", func(_ caveRecords, _ map[int]*worldgen.Zone, m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			m[8].SpecialTiles[0].SubIndex = 24
		}},
		{"same map", func(_ caveRecords, _ map[int]*worldgen.Zone, m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			m[8] = m[2]
		}},
		{"wrong slot", func(_ caveRecords, _ map[int]*worldgen.Zone, _ map[int]*gameworld.Map, s *d2mapgen.CaveWarpSpec) {
			s.FirstVisSlot = 0
		}},
		{"wrong destination", func(
			r caveRecords, _ map[int]*worldgen.Zone,
			_ map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			r["data/global/excel/Levels.txt"][1]["Vis0"] = "3"
		}},
		{"wrong LvlWarp", func(r caveRecords, _ map[int]*worldgen.Zone, _ map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			r["data/global/excel/Levels.txt"][0]["Warp3"] = "1"
		}},
		{"missing LvlWarp", func(
			r caveRecords, _ map[int]*worldgen.Zone,
			_ map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			r["data/global/excel/LvlWarp.txt"] = nil
		}},
		{"duplicate record", func(
			r caveRecords, _ map[int]*worldgen.Zone,
			_ map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			p := "data/global/excel/LvlWarp.txt"
			r[p] = append(r[p], r[p][0])
		}},
		{"malformed offset", func(
			r caveRecords, _ map[int]*worldgen.Zone,
			_ map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			r["data/global/excel/LvlWarp.txt"][0]["OffsetX"] = "nan"
		}},
		{"outside marker", func(_ caveRecords, _ map[int]*worldgen.Zone, m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			m[2].SpecialTiles[0].X = 999
		}},
		{"blocked marker", func(_ caveRecords, _ map[int]*worldgen.Zone, m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec) {
			for y := 5; y < 16; y++ {
				for x := 5; x < 16; x++ {
					blockCaveTile(m[8], x, y)
				}
			}
			m[8].RebuildFlags()
		}},
		{"disconnected interior", func(
			_ caveRecords, _ map[int]*worldgen.Zone,
			m map[int]*gameworld.Map, _ *d2mapgen.CaveWarpSpec,
		) {
			for y := 0; y < 24; y++ {
				blockCaveTile(m[2], 11, y)
			}
			m[2].RebuildFlags()
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, z, m, s := caveFixture(t)
			test.edit(r, z, m, &s)
			pair, err := resolveCaveWarp(r, z, m, s)
			if err == nil || pair != (CaveWarpPair{}) {
				t.Fatalf("invalid pair published: %#v %v", pair, err)
			}
		})
	}
}

// A disconnected pocket at the geometric center must not hide the reachable
// interior of an otherwise valid authored chamber (owned definition 85/File2).
func TestDenCaveWarpAcceptsReachableInteriorBesideCenterPocket(t *testing.T) {
	records, zones, worlds, spec := caveFixture(t)
	for y := 11; y <= 13; y++ {
		for x := 11; x <= 13; x++ {
			if x == 11 || x == 13 || y == 11 || y == 13 {
				blockCaveTile(worlds[8], x, y)
			}
		}
	}
	worlds[8].RebuildFlags()
	if _, err := resolveCaveWarp(records, zones, worlds, spec); err != nil {
		t.Fatal(err)
	}
}

// A safe pair confined to the corner cannot publish a usable indoor level.
func TestDenCaveWarpRejectsIsolatedCorner(t *testing.T) {
	records, zones, worlds, spec := caveFixture(t)
	worlds[8].SpecialTiles[0].X, worlds[8].SpecialTiles[0].Y = 2, 2
	for cell := 0; cell < 24; cell++ {
		blockCaveTile(worlds[8], 5, cell)
		blockCaveTile(worlds[8], cell, 5)
	}
	worlds[8].RebuildFlags()
	if _, err := resolveCaveWarp(records, zones, worlds, spec); err == nil {
		t.Fatal("corner pocket admitted")
	}
}

// Even a cave connected to the geometric center must reach the established
// outdoor spawn; an unrelated fenced component is not a valid approach.
func TestDenCaveApproachRejectsDisconnectedOutdoorSpawn(t *testing.T) {
	_, _, worlds, _ := caveFixture(t)
	for y := 0; y < 24; y++ {
		blockCaveTile(worlds[2], 15, y)
	}
	worlds[2].RebuildFlags()
	if err := validateCaveApproach(worlds[2], gameworld.Point{X: 51, Y: 49}, gameworld.Point{X: 90, Y: 90}); err == nil {
		t.Fatal("disconnected outdoor spawn admitted")
	}
}
