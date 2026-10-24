package entryworld

import (
	"math"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// The published town strip must stay inside the selected, traversable edge corridor in each supported stamp.
func TestOwnedTownExitCorridorsReachVisibleBoundary(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	for _, test := range []struct {
		seed             uint64
		direction        string
		laneMin, laneMax int
	}{
		{1, "east", 82, 86}, {2, "south", 146, 158},
		{5, "west", 106, 118}, {9, "north", 146, 158},
	} {
		t.Run(test.direction, func(t *testing.T) {
			prepared, err := Build(t.Context(), source, content.D2Legacy(), records,
				entryWorldObjectResolver{}, test.seed, 0)
			if err != nil {
				t.Fatal(err)
			}
			town := prepared.Worlds[1]
			seam := prepared.Seams[0]
			if seam.First.Direction != test.direction || seam.First.LevelID != 1 || seam.Second.LevelID != 2 {
				t.Fatalf("seed %d selected unexpected town seam: %#v", test.seed, seam)
			}
			if strip := seam.First.TriggerStrip; strip == nil ||
				strip.LaneMin != test.laneMin || strip.LaneMax != test.laneMax {
				t.Fatalf("seed %d published town strip = %#v, want lanes %d..%d",
					test.seed, strip, test.laneMin, test.laneMax)
			}
			if seam.Second.TriggerStrip != nil {
				t.Fatalf("seed %d reverse endpoint has town strip", test.seed)
			}
			strip := seam.First.TriggerStrip
			if strip.LaneMin > strip.LaneMax {
				t.Fatalf("seed %d town trigger has no lanes", test.seed)
			}

			outer := ownedTownOuterPoint(t, town, seam.First)
			from := gameworld.Point{X: seam.First.X, Y: seam.First.Y}
			stepX, stepY := 0, 0
			switch test.direction {
			case "north":
				stepY = -1
			case "east":
				stepX = 1
			case "south":
				stepY = 1
			case "west":
				stepX = -1
			}
			for from != outer {
				to := gameworld.Point{X: from.X + float64(stepX), Y: from.Y + float64(stepY)}
				if !ownedTownFloorAt(town, to) || !town.WalkableStep(from, to, 1) {
					t.Fatalf("seed %d exit lane stops before visible boundary at %#v", test.seed, to)
				}
				from = to
			}

			spawn := prepared.Spawns[1]
			for lane := strip.LaneMin; lane <= strip.LaneMax; lane++ {
				lanePoint := gameworld.Point{X: seam.First.X, Y: seam.First.Y}
				laneOuter := gameworld.Point{X: outer.X, Y: outer.Y}
				if test.direction == "east" || test.direction == "west" {
					lanePoint.Y, laneOuter.Y = float64(lane), float64(lane)
				} else {
					lanePoint.X, laneOuter.X = float64(lane), float64(lane)
				}
				if _, err := town.FindPath(gameworld.PathRequest{
					Start: gameworld.Point{X: spawn[0], Y: spawn[1]}, Goal: lanePoint, Radius: 1,
				}); err != nil {
					t.Fatalf("seed %d exit lane %d is unreachable from normal spawn: %v", test.seed, lane, err)
				}
				for point := lanePoint; point != laneOuter; {
					next := gameworld.Point{X: point.X + float64(stepX), Y: point.Y + float64(stepY)}
					if !ownedTownFloorAt(town, next) || !town.WalkableStep(point, next, 1) {
						t.Fatalf("seed %d exit lane %d stops before boundary at %v", test.seed, lane, next)
					}
					point = next
				}
			}
			if test.direction == "east" || test.direction == "west" {
				if math.Abs(seam.First.ArrivalX-seam.First.X) <= seam.TriggerRadius {
					t.Fatalf("seed %d town arrival overlaps strip depth", test.seed)
				}
			} else if math.Abs(seam.First.ArrivalY-seam.First.Y) <= seam.TriggerRadius {
				t.Fatalf("seed %d town arrival overlaps strip depth", test.seed)
			}
			if math.Hypot(seam.Second.ArrivalX-seam.Second.X,
				seam.Second.ArrivalY-seam.Second.Y) <= seam.TriggerRadius {
				t.Fatalf("seed %d moor arrival overlaps reverse circle", test.seed)
			}
			if _, err := prepared.Worlds[2].FindPath(gameworld.PathRequest{
				Start:  gameworld.Point{X: seam.Second.ArrivalX, Y: seam.Second.ArrivalY},
				Goal:   gameworld.Point{X: seam.Second.X, Y: seam.Second.Y},
				Radius: 1, StopRadius: seam.TriggerRadius,
			}); err != nil {
				t.Fatalf("seed %d radius-one moor arrival cannot return to seam: %v", test.seed, err)
			}
			if _, err := town.FindPath(gameworld.PathRequest{
				Start: gameworld.Point{X: spawn[0], Y: spawn[1]},
				Goal:  gameworld.Point{X: seam.First.X, Y: seam.First.Y}, Radius: 1,
			}); err != nil {
				t.Fatalf("seed %d ordinary town spawn cannot reach exit: %v", test.seed, err)
			}
			if _, err := town.FindPath(gameworld.PathRequest{
				Start:  gameworld.Point{X: seam.First.ArrivalX, Y: seam.First.ArrivalY},
				Goal:   gameworld.Point{X: seam.First.X, Y: seam.First.Y},
				Radius: 1, StopRadius: seam.TriggerRadius,
			}); err != nil {
				t.Fatalf("seed %d town return arrival cannot reach exit: %v", test.seed, err)
			}
			if test.seed == 1 {
				assertOwnedEastBridgeRoute(t, town, spawn, seam, outer)
			}
		})
	}
}

// ownedTownOuterPoint verifies the trigger is inset one tile, leaving a player-safe walk to the named outer edge.
func ownedTownOuterPoint(t *testing.T, town *gameworld.Map, endpoint gametransition.SeamEndpoint) gameworld.Point {
	t.Helper()
	switch endpoint.Direction {
	case "north":
		if endpoint.Y != gameworld.SubtilesPerTile {
			t.Fatalf("north endpoint y = %v, want one tile inside edge", endpoint.Y)
		}
		return gameworld.Point{X: endpoint.X, Y: 1}
	case "east":
		if endpoint.X != float64(town.WidthSubtiles-gameworld.SubtilesPerTile) {
			t.Fatalf("east endpoint x = %v, want one tile inside edge", endpoint.X)
		}
		return gameworld.Point{X: float64(town.WidthSubtiles - 2), Y: endpoint.Y}
	case "south":
		if endpoint.Y != float64(town.HeightSubtiles-gameworld.SubtilesPerTile) {
			t.Fatalf("south endpoint y = %v, want one tile inside edge", endpoint.Y)
		}
		return gameworld.Point{X: endpoint.X, Y: float64(town.HeightSubtiles - 2)}
	case "west":
		if endpoint.X != gameworld.SubtilesPerTile {
			t.Fatalf("west endpoint x = %v, want one tile inside edge", endpoint.X)
		}
		return gameworld.Point{X: 1, Y: endpoint.Y}
	default:
		t.Fatalf("unknown town direction %q", endpoint.Direction)
		return gameworld.Point{}
	}
}

// ownedTownFloorAt distinguishes a displayed floor from collision space that happens to be open.
func ownedTownFloorAt(town *gameworld.Map, point gameworld.Point) bool {
	tileX := gameworld.CollisionCell(point.X) / gameworld.SubtilesPerTile
	tileY := gameworld.CollisionCell(point.Y) / gameworld.SubtilesPerTile
	for _, tile := range town.Tiles {
		if tile.X == tileX && tile.Y == tileY && tile.Layer == gameworld.LayerFloor && tile.Reference.Path != "" {
			return true
		}
	}
	return false
}

// assertOwnedEastBridgeRoute ties the selected seed to Bridge.dt1 and the ordinary spawn's trigger path.
func assertOwnedEastBridgeRoute(
	t *testing.T, town *gameworld.Map, spawn [2]float64,
	seam gametransition.Seam, outer gameworld.Point,
) {
	t.Helper()
	bridge := false
	for _, tile := range town.Tiles {
		if tile.X == gameworld.CollisionCell(seam.First.X)/gameworld.SubtilesPerTile &&
			tile.Y == gameworld.CollisionCell(seam.First.Y)/gameworld.SubtilesPerTile &&
			tile.Layer == gameworld.LayerFloor && strings.HasSuffix(strings.ToLower(tile.Reference.Path), "/bridge.dt1") {
			bridge = true
		}
	}
	if !bridge {
		t.Fatalf("east seam (%v,%v) is not on the authored bridge floor", seam.First.X, seam.First.Y)
	}
	path, err := town.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: spawn[0], Y: spawn[1]}, Goal: outer, Radius: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range path {
		if math.Hypot(point.X-seam.First.X, point.Y-seam.First.Y) <= seam.TriggerRadius {
			return
		}
	}
	t.Fatal("ordinary path across visible east bridge misses the town seam trigger")
}
