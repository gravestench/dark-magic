package entryworld

import (
	"reflect"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// A cave-bearing topology must publish one final zone from its validated map,
// including both places where a resident would obstruct the clicked warp.
func TestPopulationDataDenRequiresCompleteCave(t *testing.T) {
	prepared := populationDenFixture(t)
	data, err := prepared.PopulationData(0)
	if err != nil {
		t.Fatal(err)
	}
	zones := data["zones"].([]any)
	if len(zones) != 4 || zones[3].(map[string]any)["level_id"] != float64(8) {
		t.Fatalf("population zones = %#v, want levels 2,3,4,8", zones)
	}
	den := zones[3].(map[string]any)
	if got := den["exclusions"]; !reflect.DeepEqual(got, []any{
		map[string]any{"x": float64(27), "y": float64(27), "radius": float64(1)},
		map[string]any{"x": float64(47), "y": float64(47), "radius": float64(3.5)},
	}) {
		t.Fatalf("Den endpoint exclusions = %#v", got)
	}
	for _, edit := range []struct {
		name  string
		apply func(*Prepared)
	}{
		{"missing zone", func(p *Prepared) { delete(p.Zones, 8) }},
		{"missing map", func(p *Prepared) { delete(p.Worlds, 8) }},
		{"wrong zone", func(p *Prepared) { p.Zones[8] = p.Zones[4] }},
		{"wrong target", func(p *Prepared) { p.CaveWarp.First.TargetLevel = 4 }},
		{"wrong source", func(p *Prepared) { p.CaveWarp.Second.LevelID = 3 }},
		{"wrong dimensions", func(p *Prepared) { p.CaveWarp.First.Width = 99 }},
	} {
		t.Run(edit.name, func(t *testing.T) {
			p := populationDenFixture(t)
			edit.apply(p)
			if data, err := p.PopulationData(0); err == nil || data != nil {
				t.Fatalf("incomplete cave produced data=%#v err=%v", data, err)
			}
		})
	}
}

// Seed 42's owned chamber has the same important shape: its open nearest-center
// pocket is disconnected from the arrival, so taking the first four open cells
// would publish monsters a player cannot encounter.
func TestPopulationDataDenSkipsDisconnectedCenter(t *testing.T) {
	prepared := populationDenFixture(t)
	den := prepared.Worlds[8]
	for y := 6; y <= 8; y++ {
		for x := 6; x <= 8; x++ {
			if x == 7 && y == 7 {
				continue
			}
			tile := gameworld.TilePlacement{X: x, Y: y, Layer: gameworld.LayerFloor}
			tile.Reference.Path = "synthetic"
			for index := range tile.Reference.SubTileFlags {
				tile.Reference.SubTileFlags[index].BlockWalk = true
			}
			den.Tiles = append(den.Tiles, tile)
		}
	}
	den.RebuildFlags()
	data, err := prepared.PopulationData(0)
	if err != nil {
		t.Fatal(err)
	}
	room := data["zones"].([]any)[3].(map[string]any)["rooms"].([]any)[0].(map[string]any)
	points := room["points"].([]any)
	if len(points) == 0 {
		t.Fatal("Den has no reachable population points")
	}
	for _, raw := range points {
		point := raw.(map[string]any)
		x, y := point["x"].(float64), point["y"].(float64)
		if x >= 35 && x < 40 && y >= 35 && y < 40 {
			t.Fatalf("disconnected center point (%v,%v) was published", x, y)
		}
		if _, err := den.FindPath(gameworld.PathRequest{
			Start: gameworld.Point{X: 27, Y: 27}, Goal: gameworld.Point{X: x, Y: y}, Radius: 1,
		}); err != nil {
			t.Fatalf("point (%v,%v) cannot be reached from cave arrival: %v", x, y, err)
		}
	}
}

func populationDenFixture(t *testing.T) *Prepared {
	t.Helper()
	prepared := populationPreparedFixture(t)
	zone, err := worldgen.NewZone(worldgen.Definition{
		Request: worldgen.Request{Version: worldgen.ContractVersion, Seed: 41, Act: 1, LevelID: 8},
		Kind:    "test", Bounds: worldgen.Bounds{Width: 20, Height: 20},
		Rooms: []worldgen.Room{{ID: 1, X: 5, Y: 5, Width: 5, Height: 5}},
	})
	if err != nil {
		t.Fatal(err)
	}
	den, err := gameworld.NewOpenMap(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	prepared.Zones[8], prepared.Worlds[8] = zone, den
	prepared.CaveWarp = &CaveWarpPair{
		First:  CaveWarpEndpoint{LevelID: 2, TargetLevel: 8, ArrivalX: 27, ArrivalY: 27, Width: 100, Height: 100},
		Second: CaveWarpEndpoint{LevelID: 8, TargetLevel: 2, X: 47, Y: 47, Width: 100, Height: 100},
	}
	return prepared
}

// Level 4 must not reuse player anchors outside a room, or project several
// anchors onto the same open cell.
func TestPopulationDataLevelFourRoomLocalPoints(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		name := "open"
		if blocked {
			name = "blocked"
		}
		t.Run(name, func(t *testing.T) {
			prepared := populationPreparedFixture(t)
			if blocked {
				// Block the whole room while leaving the surrounding map open.
				// Projecting an anchor outside the room must not rescue it.
				for y := 5; y < 10; y++ {
					for x := 5; x < 10; x++ {
						tile := gameworld.TilePlacement{X: x, Y: y, Layer: gameworld.LayerFloor}
						tile.Reference.Path = "synthetic"
						for i := range tile.Reference.SubTileFlags {
							tile.Reference.SubTileFlags[i].BlockWalk = true
						}
						prepared.Worlds[4].Tiles = append(prepared.Worlds[4].Tiles, tile)
					}
				}
				prepared.Worlds[4].RebuildFlags()
			}
			for _, nearby := range []int{0, 1} {
				first, err := prepared.PopulationData(nearby)
				if err != nil {
					t.Fatal(err)
				}
				second, err := prepared.PopulationData(nearby)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(first, second) {
					t.Fatal("population geometry changed between calls")
				}
				room := first["zones"].([]any)[2].(map[string]any)["rooms"].([]any)[0].(map[string]any)
				points := room["points"].([]any)
				if blocked && len(points) != 0 {
					t.Fatalf("blocked room points=%#v", points)
				}
				if !blocked && len(points) != 4 {
					t.Fatalf("open room points=%#v, want four", points)
				}
				assertPopulationRoomPoints(t, prepared.Worlds[4], room)
			}
		})
	}
}

// Candidate points remain distinct, room-local, and open so geometry cannot admit an unsafe center.
func assertPopulationRoomPoints(t *testing.T, gameMap *gameworld.Map, room map[string]any) {
	t.Helper()
	points := room["points"].([]any)
	if len(points) > 4 {
		t.Fatalf("room %v has %d points, want at most four", room["id"], len(points))
	}
	seen := map[[2]float64]bool{}
	for _, raw := range points {
		point := raw.(map[string]any)
		xy := [2]float64{point["x"].(float64), point["y"].(float64)}
		x, y := room["x"].(float64), room["y"].(float64)
		if seen[xy] || xy[0] < x || xy[0] >= x+room["width"].(float64) ||
			xy[1] < y || xy[1] >= y+room["height"].(float64) {
			t.Fatalf("room %v has duplicate or nonlocal point %v", room["id"], xy)
		}
		flags, ok := gameMap.FlagsAt(gameworld.CollisionCell(xy[0]), gameworld.CollisionCell(xy[1]))
		if !ok || flags.Blocked() {
			t.Fatalf("room %v has blocked point %v", room["id"], xy)
		}
		seen[xy] = true
	}
}
