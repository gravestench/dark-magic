package transition_test

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	"github.com/gravestench/dt1"
)

// TestTownMoorSeamUsesAuthoredTownExitAndGeneratedMoorEdge protects asymmetric endpoint source policy.
func TestTownMoorSeamUsesAuthoredTownExitAndGeneratedMoorEdge(t *testing.T) {
	town, _ := gameworld.NewOpenMap(40, 40)
	town.SpecialTiles = []gameworld.SpecialTile{{X: 7, Y: 4, Orientation: 10}}
	moor, _ := gameworld.NewOpenMap(40, 40)

	seam, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
		FirstLevel:      1,
		FirstDirection:  "east",
		SecondLevel:     2,
		SecondDirection: "west",
		SecondTileX:     0,
		SecondTileY:     4,
	}, town, moor)
	if err != nil {
		t.Fatal(err)
	}

	if seam.First.LevelID != 1 || seam.Second.LevelID != 2 ||
		seam.First.Direction != "east" || seam.Second.Direction != "west" {
		t.Fatalf("seam = %#v", seam)
	}
}

// The fixed east/west policy isolates collision behavior from zone selection in safety tests.
func generatedEastWestSpec() d2mapgen.SeamSpec {
	return d2mapgen.SeamSpec{
		FirstLevel: 2, SecondLevel: 3, FirstDirection: "east", SecondDirection: "west",
		FirstTileX: 7, FirstTileY: 4, SecondTileX: 0, SecondTileY: 4,
		FirstGenerated: true, TriggerRadius: 2, MaxArrivalDistance: 10, ArrivalInset: 6,
	}
}

// Full tile columns model a blocked band or an isolated edge strip after collision rebuild.
func blockTileColumns(world *gameworld.Map, first, last int) {
	var flags [25]dt1.SubTileFlags
	for index := range flags {
		flags[index].BlockWalk = true
	}
	for tileX := first; tileX <= last; tileX++ {
		for tileY := 0; tileY < 8; tileY++ {
			world.Tiles = append(world.Tiles, gameworld.TilePlacement{
				X: tileX, Y: tileY, Layer: gameworld.LayerFloor,
				Reference: gameworld.TileReference{Path: "wall.dt1", SubTileFlags: flags},
			})
		}
	}
	world.RebuildFlags()
}

// blockSubtileColumn leaves the arrival and edge individually open while testing whether collision separates them.
func blockSubtileColumn(world *gameworld.Map, x int) {
	var flags [25]dt1.SubTileFlags
	for index := range flags {
		if index%gameworld.SubtilesPerTile == x%gameworld.SubtilesPerTile {
			flags[index].BlockWalk = true
		}
	}
	for tileY := 0; tileY < 8; tileY++ {
		world.Tiles = append(world.Tiles, gameworld.TilePlacement{
			X: x / gameworld.SubtilesPerTile, Y: tileY, Layer: gameworld.LayerFloor,
			Reference: gameworld.TileReference{Path: "wall.dt1", SubTileFlags: flags},
		})
	}
	world.RebuildFlags()
}

// A narrow open edge strip cannot serve as the interior needed for an ordinary return trip.
func TestResolveGeneratedSeamRejectsEdgeStripWithoutInterior(t *testing.T) {
	moor, _ := gameworld.NewOpenMap(40, 40)
	plains, _ := gameworld.NewOpenMap(40, 40)
	blockTileColumns(moor, 0, 5)
	if _, err := gametransition.ResolveSeam(generatedEastWestSpec(), moor, plains); err == nil {
		t.Fatal("edge strip was accepted as a reachable interior")
	}
}

// A walkable arrival must also be able to reach the return trigger across the same collision map.
func TestResolveGeneratedSeamRejectsSeparatedReturnTrigger(t *testing.T) {
	moor, _ := gameworld.NewOpenMap(40, 40)
	plains, _ := gameworld.NewOpenMap(40, 40)
	blockSubtileColumn(moor, 35)
	if _, err := gametransition.ResolveSeam(generatedEastWestSpec(), moor, plains); err == nil {
		t.Fatal("unreachable return trigger was accepted")
	}
}

// Generated arrivals must stay beyond the return trigger, within the search bound, and connected inward.
func TestResolveGeneratedSeamRejectsUnsafeArrival(t *testing.T) {
	for _, test := range []struct {
		name                  string
		editSpec              func(*d2mapgen.SeamSpec)
		blockFirst, blockLast int
	}{
		{"immediate return", func(spec *d2mapgen.SeamSpec) { spec.ArrivalInset = 2 }, -1, -1},
		{"arrival too distant", nil, 6, 7},
		{"no interior path", nil, 4, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			moor, _ := gameworld.NewOpenMap(40, 40)
			plains, _ := gameworld.NewOpenMap(40, 40)
			if test.blockFirst >= 0 {
				blockTileColumns(moor, test.blockFirst, test.blockLast)
			}
			spec := generatedEastWestSpec()
			if test.editSpec != nil {
				test.editSpec(&spec)
			}
			if _, err := gametransition.ResolveSeam(spec, moor, plains); err == nil {
				t.Fatal("unsafe generated seam was accepted")
			}
		})
	}
}

// The same safe-arrival rule must work for all four opposing cardinal edge pairs.
func TestResolveGeneratedSeamAllDirections(t *testing.T) {
	for _, facing := range []struct {
		first, second                    string
		firstX, firstY, secondX, secondY int
	}{
		{"north", "south", 4, 0, 4, 7},
		{"east", "west", 7, 4, 0, 4},
		{"south", "north", 4, 7, 4, 0},
		{"west", "east", 0, 4, 7, 4},
	} {
		t.Run(facing.first, func(t *testing.T) {
			moor, _ := gameworld.NewOpenMap(40, 40)
			plains, _ := gameworld.NewOpenMap(40, 40)
			seam, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
				FirstLevel: 2, SecondLevel: 3,
				FirstDirection: facing.first, SecondDirection: facing.second,
				FirstTileX: facing.firstX, FirstTileY: facing.firstY,
				SecondTileX: facing.secondX, SecondTileY: facing.secondY,
				FirstGenerated: true, TriggerRadius: 2, MaxArrivalDistance: 10, ArrivalInset: 6,
			}, moor, plains)
			if err != nil {
				t.Fatal(err)
			}
			if seam.First.LevelID != 2 || seam.Second.LevelID != 3 || seam.TriggerRadius != 2 {
				t.Fatalf("resolved generated seam = %#v", seam)
			}
			if seam.First.Direction != facing.first || seam.Second.Direction != facing.second {
				t.Fatalf("resolved directions = %#v", seam)
			}
		})
	}
}
