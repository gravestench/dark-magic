package entryworld

import (
	"fmt"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// Fresh and restored admission must share the fully validated owned Den arrival.
// Seeds 17/42 regress inherited fill fences; 42 also exercises the cave chamber
// with a disconnected nearest-center pocket. No protected assets are fixtures.
func TestDenOfEvilPreparedReal(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	for _, seed := range []uint64{0, 17, 42, 73} {
		for difficulty := 0; difficulty < 3; difficulty++ {
			t.Run(fmt.Sprintf("seed-%d/difficulty-%d", seed, difficulty), func(t *testing.T) {
				prepared, err := Build(t.Context(), source, content.D2Legacy(), records, entryWorldObjectResolver{}, seed, difficulty)
				if err != nil {
					t.Fatal(err)
				}
				if len(prepared.Worlds) != 5 || len(prepared.Seams) != 3 || prepared.CaveWarp == nil {
					t.Fatal("wrong map/seam/cave topology")
				}
				destination, err := prepared.Destination(8)
				if err != nil {
					t.Fatal(err)
				}
				pair := *prepared.CaveWarp
				if destination.X != pair.First.ArrivalX || destination.Y != pair.First.ArrivalY || destination.LevelID != 8 {
					t.Fatal("Den admission did not use the validated arrival")
				}
				for _, endpoints := range [][2]CaveWarpEndpoint{{pair.First, pair.Second}, {pair.Second, pair.First}} {
					side, reverse := endpoints[0], endpoints[1]
					m := prepared.Worlds[side.LevelID]
					if side.TargetLevel != reverse.LevelID || side.PairedID != reverse.ID {
						t.Fatal("unpaired cave")
					}
					target := prepared.Worlds[side.TargetLevel]
					if side.Width != float64(target.WidthSubtiles) || side.Height != float64(target.HeightSubtiles) {
						t.Fatal("wrong destination size")
					}
					if _, err := m.FindPath(gameworld.PathRequest{
						Start: gameworld.Point{X: reverse.ArrivalX, Y: reverse.ArrivalY},
						Goal:  gameworld.Point{X: side.X, Y: side.Y}, Radius: 1,
					}); err != nil {
						t.Fatalf("level %d has no return route: %v", side.LevelID, err)
					}
				}
				outdoorSpawn := prepared.Spawns[2]
				if _, err := prepared.Worlds[2].FindPath(gameworld.PathRequest{
					Start: gameworld.Point{X: pair.First.X, Y: pair.First.Y},
					Goal:  gameworld.Point{X: outdoorSpawn[0], Y: outdoorSpawn[1]}, Radius: 1,
				}); err != nil {
					t.Fatalf("cave is disconnected from the existing outdoor join: %v", err)
				}
				saved, err := prepared.Topology.Snapshot()
				if err != nil {
					t.Fatal(err)
				}
				restored, err := RestorePrepared(t.Context(), source, content.D2Legacy(), records, entryWorldObjectResolver{}, saved)
				if err != nil {
					t.Fatal(err)
				}
				again, err := restored.Destination(8)
				if err != nil || again != destination || restored.CaveWarp == nil || *restored.CaveWarp != pair {
					t.Fatalf("restored Den admission/endpoint identity drift: %v", err)
				}
				if restored.TopologyChecksum != prepared.TopologyChecksum {
					t.Fatal("checksum drift")
				}
			})
		}
	}
}
