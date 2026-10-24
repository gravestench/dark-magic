package entryworld

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// Pin the owned 1.14d rows that make level 8 an ordinary monster area. This
// guards candidate selection without shipping the protected tables as fixtures.
func TestDenPopulationOwnedRowsReal(t *testing.T) {
	records := recordstore.New(realEntryWorldContent(t))
	levels, err := records.Load("data/global/excel/Levels.txt")
	if err != nil {
		t.Fatal(err)
	}
	level := ownedDenRow(t, levels, "Id", "8")
	for field, want := range map[string]string{
		"NumMon": "3", "MonDen": "600", "MonDen(N)": "600", "MonDen(H)": "600",
		"mon1": "zombie1", "mon2": "brute1", "mon3": "fallenshaman1",
		"nmon1": "zombie1", "nmon2": "brute1", "nmon3": "fallenshaman1",
	} {
		if got := level[field]; got != want {
			t.Fatalf("Levels.txt level 8 %s=%q, want %q", field, got, want)
		}
	}
	monsters, err := records.Load("data/global/excel/MonStats.txt")
	if err != nil {
		t.Fatal(err)
	}
	graphics, err := records.Load("data/global/excel/MonStats2.txt")
	if err != nil {
		t.Fatal(err)
	}
	for id, life := range map[string][2]string{
		"zombie1": {"101", "181"}, "brute1": {"129", "215"}, "fallenshaman1": {"58", "101"},
	} {
		row := ownedDenRow(t, monsters, "Id", id)
		image := ownedDenRow(t, graphics, "Id", row["MonStatsEx"])
		wantSize := "2"
		if id == "brute1" {
			wantSize = "3"
		}
		if image["SizeX"] != wantSize || image["SizeY"] != wantSize {
			t.Fatalf("MonStats2.txt %s collider size changed: %q/%q", id, image["SizeX"], image["SizeY"])
		}
		if row["minHP"] != life[0] || row["maxHP"] != life[1] || row["noRatio"] != "" ||
			row["minHP(N)"] != "" || row["maxHP(N)"] != "" ||
			row["minHP(H)"] != "" || row["maxHP(H)"] != "" {
			t.Fatalf("MonStats.txt %s Den life percentages changed", id)
		}
		if row["enabled"] != "1" || row["isSpawn"] != "1" || row["npc"] != "" {
			t.Fatalf("MonStats.txt %s is not an ordinary spawn: enabled=%q isSpawn=%q npc=%q",
				id, row["enabled"], row["isSpawn"], row["npc"])
		}
	}
	presets, err := records.Load("data/global/excel/LvlPrest.txt")
	if err != nil {
		t.Fatal(err)
	}
	for definition := 83; definition <= 86; definition++ {
		row := ownedDenRow(t, presets, "Def", fmt.Sprint(definition))
		if row["Populate"] != "1" {
			t.Fatalf("LvlPrest.txt Def %d Populate=%q, want 1", definition, row["Populate"])
		}
	}
}

// ownedDenRow fails at the table boundary so missing source data cannot look like an empty field.
func ownedDenRow(t *testing.T, rows []map[string]string, key, value string) map[string]string {
	t.Helper()
	for _, row := range rows {
		if row[key] == value {
			return row
		}
	}
	t.Fatalf("owned row %s=%s is missing", key, value)
	return nil
}

// The prepared cave must yield deterministic, arrival-reachable room points
// in each difficulty. Seed 42 exercises the disconnected nearest-center pocket.
func TestDenPopulationGeometryReal(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	seenDefinitions := map[int]bool{}
	for _, seed := range []uint64{0, 1, 2, 3, 4, 17, 42, 73} {
		for difficulty := 0; difficulty < 3; difficulty++ {
			t.Run(fmt.Sprintf("seed-%d/difficulty-%d", seed, difficulty), func(t *testing.T) {
				prepared, err := Build(t.Context(), source, content.D2Legacy(), records,
					entryWorldObjectResolver{}, seed, difficulty)
				if err != nil {
					t.Fatal(err)
				}
				for _, stamp := range prepared.Zones[8].Stamps() {
					seenDefinitions[stamp.PresetDef] = true
					t.Logf("owned Den chamber definition %d variant %d", stamp.PresetDef, stamp.Variant)
				}
				first, err := prepared.PopulationData(0)
				if err != nil {
					t.Fatal(err)
				}
				second, err := prepared.PopulationData(0)
				if err != nil || !reflect.DeepEqual(first, second) {
					t.Fatalf("Den population changed between identical calls: %v", err)
				}
				zones := first["zones"].([]any)
				if len(zones) != 4 || zones[3].(map[string]any)["level_id"] != float64(8) {
					t.Fatalf("population zones=%d, want final level 8", len(zones))
				}
				den := zones[3].(map[string]any)
				rooms := den["rooms"].([]any)
				if len(rooms) != 1 {
					t.Fatalf("Den rooms=%d, want 1", len(rooms))
				}
				room := rooms[0].(map[string]any)
				if !room["populate"].(bool) {
					t.Fatal("owned Den room is not marked for population")
				}
				assertPopulationRoomPoints(t, prepared.Worlds[8], room)
				points := room["points"].([]any)
				if len(points) == 0 {
					t.Fatal("Den has no population points")
				}
				for _, raw := range points {
					point := raw.(map[string]any)
					if _, err := prepared.Worlds[8].FindPath(gameworld.PathRequest{
						Start: gameworld.Point{X: prepared.CaveWarp.First.ArrivalX, Y: prepared.CaveWarp.First.ArrivalY},
						Goal:  gameworld.Point{X: point["x"].(float64), Y: point["y"].(float64)}, Radius: 1,
					}); err != nil {
						t.Fatalf("Den candidate cannot be reached from arrival: %v", err)
					}
				}
			})
		}
	}
	for definition := 83; definition <= 86; definition++ {
		if !seenDefinitions[definition] {
			t.Fatalf("owned Den chamber definition %d was not exercised", definition)
		}
	}
}

// Fresh and restored admission must share the fully validated owned Den arrival.
// Seeds 17/42 regress inherited fill fences; 42 also exercises the cave chamber
// with a disconnected nearest-center pocket. No protected assets are fixtures.
func TestDenOfEvilPreparedReal(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	for _, seed := range []uint64{0, 17, 42, 73} {
		for difficulty := 0; difficulty < 3; difficulty++ {
			t.Run(fmt.Sprintf("seed-%d/difficulty-%d", seed, difficulty), func(t *testing.T) {
				prepared, err := Build(
					t.Context(), source, content.D2Legacy(), records, entryWorldObjectResolver{}, seed, difficulty,
				)
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
				restored, err := RestorePrepared(
					t.Context(), source, content.D2Legacy(), records, entryWorldObjectResolver{}, saved,
				)
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
