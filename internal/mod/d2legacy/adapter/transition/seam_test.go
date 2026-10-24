package transition_test

import (
	"strings"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	"github.com/gravestench/dt1"
)

// TestTownMoorSeamUsesAuthoredTownExitAndGeneratedMoorEdge protects asymmetric endpoint source policy.
func TestTownMoorSeamUsesAuthoredTownExitAndGeneratedMoorEdge(t *testing.T) {
	town, _ := gameworld.NewOpenMap(40, 40)
	town.SpecialTiles = []gameworld.SpecialTile{{X: 3, Y: 4, Orientation: 10}}
	town.Tiles = []gameworld.TilePlacement{{X: 7, Y: 4, Layer: gameworld.LayerFloor,
		Reference: gameworld.TileReference{Path: "floor.dt1"}}}
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
	if seam.First.X != 35 || seam.First.Y != 23 {
		t.Fatalf("town endpoint = (%v,%v), want visible edge (35,23)", seam.First.X, seam.First.Y)
	}
}

// The authored aperture bounds the exit when floor and collision continue under a neighboring fence.
func TestAuthoredTownTriggerExcludesFloorBeyondAperture(t *testing.T) {
	town, _ := gameworld.NewOpenMap(40, 40)
	town.SpecialTiles = []gameworld.SpecialTile{{X: 3, Y: 4, Orientation: 10}}
	for _, tileY := range []int{4, 5, 6} {
		town.Tiles = append(town.Tiles, gameworld.TilePlacement{X: 7, Y: tileY,
			Layer: gameworld.LayerFloor, Reference: gameworld.TileReference{Path: "floor.dt1"}})
	}
	town.Tiles = append(town.Tiles, gameworld.TilePlacement{X: 7, Y: 6,
		Layer: gameworld.LayerLowerWall, Reference: gameworld.TileReference{Path: "fence.dt1"}})
	moor, _ := gameworld.NewOpenMap(40, 40)
	seam, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
		FirstLevel: 1, FirstDirection: "east", SecondLevel: 2, SecondDirection: "west",
		SecondTileX: 0, SecondTileY: 4, TriggerRadius: 2, ArrivalInset: 6,
	}, town, moor)
	if err != nil {
		t.Fatal(err)
	}
	if seam.First.Y != 23 || seam.First.TriggerStrip == nil ||
		seam.First.TriggerStrip.LaneMin != 20 || seam.First.TriggerStrip.LaneMax != 28 {
		t.Fatalf("floor beyond authored aperture admitted or strip missing: %#v", seam.First)
	}
	if seam.Second.TriggerStrip != nil {
		t.Fatalf("generated reverse endpoint gained a town strip: %#v", seam.Second)
	}
}

// A hidden exit guide may be well inside the stamp. The named, authored floor corridor must determine its trigger.
func TestResolveAuthoredTownEdgeAllDirections(t *testing.T) {
	for _, test := range []struct {
		name, opposite string
		guide, floor   gameworld.SpecialTile
		extraFloor     []gameworld.TilePlacement
		wantX, wantY   float64
	}{
		{"east", "west", gameworld.SpecialTile{X: 3, Y: 4, Orientation: 10},
			gameworld.SpecialTile{X: 7, Y: 4}, nil, 35, 23},
		{"south", "north", gameworld.SpecialTile{X: 4, Y: 3, Orientation: 10},
			gameworld.SpecialTile{X: 4, Y: 7}, nil, 23, 35},
		{"west", "east", gameworld.SpecialTile{X: 4, Y: 4, Orientation: 10},
			gameworld.SpecialTile{X: 1, Y: 4}, []gameworld.TilePlacement{{X: 0, Y: 4, Layer: gameworld.LayerFloor,
				Reference: gameworld.TileReference{Path: "floor.dt1"}}}, 5, 23},
		{"north", "south", gameworld.SpecialTile{X: 4, Y: 4, Orientation: 10},
			gameworld.SpecialTile{X: 4, Y: 1}, []gameworld.TilePlacement{{X: 4, Y: 0, Layer: gameworld.LayerFloor,
				Reference: gameworld.TileReference{Path: "floor.dt1"}}}, 23, 5},
	} {
		t.Run(test.name, func(t *testing.T) {
			town, _ := gameworld.NewOpenMap(40, 40)
			town.SpecialTiles = []gameworld.SpecialTile{test.guide}
			town.Tiles = append(test.extraFloor, gameworld.TilePlacement{
				X: test.floor.X, Y: test.floor.Y, Layer: gameworld.LayerFloor,
				Reference: gameworld.TileReference{Path: "floor.dt1"},
			})
			moor, _ := gameworld.NewOpenMap(40, 40)
			seam, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
				FirstLevel: 1, FirstDirection: test.name, SecondLevel: 2, SecondDirection: test.opposite,
				SecondTileX: 4, SecondTileY: 4, TriggerRadius: 2, ArrivalInset: 6,
			}, town, moor)
			if err != nil {
				t.Fatal(err)
			}
			if seam.First.X != test.wantX || seam.First.Y != test.wantY {
				t.Fatalf("endpoint = (%v,%v), want (%v,%v)", seam.First.X, seam.First.Y, test.wantX, test.wantY)
			}
		})
	}
}

// An open collision edge does not become an exit without authored floor across the outward lane.
func TestResolveAuthoredTownEdgeRequiresFloorToBorder(t *testing.T) {
	for _, withInnerFloor := range []bool{false, true} {
		town, _ := gameworld.NewOpenMap(40, 40)
		town.SpecialTiles = []gameworld.SpecialTile{{X: 4, Y: 4, Orientation: 10}}
		if withInnerFloor {
			town.Tiles = append(town.Tiles, gameworld.TilePlacement{X: 1, Y: 4,
				Layer: gameworld.LayerFloor, Reference: gameworld.TileReference{Path: "floor.dt1"}})
		}
		moor, _ := gameworld.NewOpenMap(40, 40)
		if _, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
			FirstDirection: "west", SecondDirection: "east", SecondTileX: 4, SecondTileY: 4,
		}, town, moor); err == nil {
			t.Fatalf("west edge without continuous authored floor accepted (inner floor %v)", withInnerFloor)
		}
	}
}

// A closer interior floor strip is a blind alley when collision blocks its continuation toward the map boundary.
func TestResolveAuthoredTownEdgeSkipsBlindAlley(t *testing.T) {
	town, _ := gameworld.NewOpenMap(40, 40)
	town.SpecialTiles = []gameworld.SpecialTile{{X: 3, Y: 2, Orientation: 10}}
	for _, y := range []int{2, 5} {
		town.Tiles = append(town.Tiles, gameworld.TilePlacement{X: 7, Y: y,
			Layer: gameworld.LayerFloor, Reference: gameworld.TileReference{Path: "floor.dt1"}})
	}
	var flags [25]dt1.SubTileFlags
	for index := range flags {
		if index%gameworld.SubtilesPerTile == 3 {
			flags[index].BlockWalk = true
		}
	}
	town.Tiles = append(town.Tiles, gameworld.TilePlacement{X: 7, Y: 2,
		Layer: gameworld.LayerLowerWall, Reference: gameworld.TileReference{Path: "wall.dt1", SubTileFlags: flags}})
	town.RebuildFlags()
	moor, _ := gameworld.NewOpenMap(40, 40)
	seam, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
		FirstDirection: "east", SecondDirection: "west", SecondTileX: 4, SecondTileY: 4,
	}, town, moor)
	if err != nil {
		t.Fatal(err)
	}
	if seam.First.X != 35 || seam.First.Y < 25 || seam.First.Y > 29 {
		t.Fatalf("selected blind alley rather than outward corridor: %#v", seam.First)
	}
}

// A town arrival stranded behind collision cannot advertise a reversible transition.
func TestResolveAuthoredTownEdgeRequiresReturnRoute(t *testing.T) {
	town, _ := gameworld.NewOpenMap(40, 40)
	// The guide can reach the exit east of the wall; the inset return arrival cannot.
	town.SpecialTiles = []gameworld.SpecialTile{{X: 6, Y: 4, Orientation: 10}}
	town.Tiles = []gameworld.TilePlacement{{X: 7, Y: 4, Layer: gameworld.LayerFloor,
		Reference: gameworld.TileReference{Path: "floor.dt1"}}}
	blockSubtileColumn(town, 30)
	moor, _ := gameworld.NewOpenMap(40, 40)
	if _, err := gametransition.ResolveSeam(d2mapgen.SeamSpec{
		FirstDirection: "east", SecondDirection: "west", SecondTileX: 4, SecondTileY: 4,
	}, town, moor); err == nil || !strings.Contains(err.Error(), "no route to return trigger") {
		t.Fatalf("town seam with unreachable return trigger error = %v", err)
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
