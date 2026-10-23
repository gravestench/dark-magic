package entryworld

import (
	"math"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

type dungeonTestObjects struct{}

func (dungeonTestObjects) ResolveStaticObject(int, int) (int, string, bool) { return 0, "", false }
func (dungeonTestObjects) ResolveDynamicObject(int, int) (string, bool)     { return "", false }

func TestDungeonAuthoredChambersAreReachable(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to an owned Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	})
	records := recordstore.New(source)
	for _, seed := range []uint64{1, 17, 42} {
		prepared, err := Build(t.Context(), source, content.D2Legacy(), records, dungeonTestObjects{}, seed, 0)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(prepared.AdditionalSeams) != 1 {
			t.Fatal("missing dungeon connection")
		}
		seam := prepared.AdditionalSeams[0]
		if seam.Town.LevelID != 2 || seam.Wilderness.LevelID != 8 {
			t.Fatal("not a Moor to Den branch")
		}
		for _, endpoint := range []struct {
			level        int
			x, y, ax, ay float64
		}{
			{2, seam.Town.X, seam.Town.Y, seam.Town.ArrivalX, seam.Town.ArrivalY},
			{8, seam.Wilderness.X, seam.Wilderness.Y, seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY},
		} {
			m := prepared.Worlds[endpoint.level]
			start, goal := gameworld.Point{X: endpoint.ax, Y: endpoint.ay}, gameworld.Point{X: endpoint.x, Y: endpoint.y}
			if !m.WalkableStep(start, start, 1) || math.Hypot(start.X-goal.X, start.Y-goal.Y) <= 3 {
				t.Fatal("unsafe or bouncing arrival")
			}
			if _, err := m.FindPath(gameworld.PathRequest{Start: start, Goal: goal, Radius: 1}); err != nil {
				t.Fatal(err)
			}
		}
		moorStart := prepared.Spawns[2]
		if _, err := prepared.Worlds[2].FindPath(gameworld.PathRequest{
			Start: gameworld.Point{X: moorStart[0], Y: moorStart[1]}, Goal: gameworld.Point{X: seam.Town.X, Y: seam.Town.Y}, Radius: 1,
		}); err != nil {
			t.Fatalf("seed %d: cave entrance unreachable from town arrival: %v", seed, err)
		}
		dungeon := seam.Wilderness
		gameMap := prepared.Worlds[8]
		start := gameworld.Point{X: dungeon.ArrivalX, Y: dungeon.ArrivalY}
		if len(prepared.Zones[8].Rooms()) < 4 {
			t.Fatal("missing promised chambers")
		}
		for _, room := range prepared.Zones[8].Rooms() {
			x, y, found := gameMap.OpenPointNearSubtileForRadius(float64((room.X+room.Width/2)*gameworld.SubtilesPerTile), float64((room.Y+room.Height/2)*gameworld.SubtilesPerTile), 1)
			if !found {
				t.Fatalf("seed %d room %d has no open center", seed, room.ID)
			}
			if _, err := gameMap.FindPath(gameworld.PathRequest{Start: start, Goal: gameworld.Point{X: x, Y: y}, Radius: 1, StopRadius: 20}); err != nil {
				t.Fatalf("seed %d room %d disconnected: %v", seed, room.ID, err)
			}
		}
		population := prepared.PopulationData(0)["additional_zones"].([]any)[0].(map[string]any)
		for _, value := range population["rooms"].([]any) {
			room := value.(map[string]any)
			points := room["points"].([]any)
			if len(points) == 0 {
				t.Fatalf("seed %d room %v has no reachable candidates", seed, room["id"])
			}
			for _, value := range points {
				point := value.(map[string]any)
				goal := gameworld.Point{X: point["x"].(float64), Y: point["y"].(float64)}
				if !gameMap.WalkableStep(goal, goal, point["clearance"].(float64)) {
					t.Fatal("unsafe clearance")
				}
				if _, err := gameMap.FindPath(gameworld.PathRequest{Start: start, Goal: goal, Radius: 1}); err != nil {
					t.Fatalf("inaccessible encounter: %v", err)
				}
			}
		}
		initial := prepared.InitialData("", false)["d2legacy.world_transitions"].(map[string]any)
		if len(initial["seams"].([]any)) != 4 {
			t.Fatal("shared authority bootstrap lost reciprocal connections")
		}
		t.Logf("seed %d: %d reachable chambers and reciprocal safe arrivals", seed, len(prepared.Zones[8].Rooms()))
	}
}
