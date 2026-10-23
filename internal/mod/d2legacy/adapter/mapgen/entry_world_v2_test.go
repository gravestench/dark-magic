package mapgen

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// entryWorldV2Records supplies only the records needed to exercise all three admitted recipes without owned assets.
func entryWorldV2Records() entryRecords {
	records := coldPlainsRecords()
	records[levelsPath] = append(records[levelsPath],
		map[string]string{
			"Id": "1", "Act": "0", "DrlgType": "2", "LevelType": "1",
			"SizeX": "40", "SizeY": "30", "SizeX(N)": "44", "SizeY(N)": "34",
			"SizeX(H)": "48", "SizeY(H)": "38",
		},
		map[string]string{
			"Id": "2", "Act": "0", "DrlgType": "3", "LevelType": "2",
			"SizeX": "80", "SizeY": "80", "SizeX(N)": "88", "SizeY(N)": "88",
			"SizeX(H)": "96", "SizeY(H)": "96",
		},
	)
	records[typesPath][1] = map[string]string{"File 1": `Act1\Town\floor.dt1`}
	records[presetsPath] = append(records[presetsPath], map[string]string{
		"Def": "7", "LevelId": "1", "Files": "1", "File1": `Act1\Town\TownE1.ds1`,
		"Dt1Mask": "1", "Populate": "1", "Logicals": "1",
	})
	for _, definition := range []int{17, 26, 27, 28, 35} {
		row := map[string]string{
			"Def": fmt.Sprint(definition), "SizeX": "8", "SizeY": "8",
			"Files": "1", "File1": fmt.Sprintf("Act1/Outdoors/fill%d.ds1", definition),
			"Dt1Mask": "1", "Populate": "1",
		}
		if definition == 26 || definition == 27 || definition == 28 {
			row["Files"] = "4"
			for variant := 2; variant <= 4; variant++ {
				row[fmt.Sprintf("File%d", variant)] = fmt.Sprintf("Act1/Outdoors/structure%d-%d.ds1", definition, variant)
			}
		}
		records[presetsPath] = append(records[presetsPath], row)
	}
	return records
}

// bloodMoorExitFacts lets malformed-edge tests bypass generation while preserving the Lua input contract.
func bloodMoorExitFacts(warps ...map[string]any) map[string]any {
	listed := make([]any, len(warps))
	for index, warp := range warps {
		listed[index] = warp
	}
	return map[string]any{
		"request": map[string]any{"level_id": float64(2)},
		"warps":   listed,
	}
}

// levelThreeExit varies only the source direction and destination under test.
func levelThreeExit(direction string, destination int) map[string]any {
	return map[string]any{
		"id": float64(2), "role": "next-level-exit", "direction": direction,
		"x": float64(79), "y": float64(40), "destination_level": float64(destination),
	}
}

// Lua policy must admit exactly one cardinal Blood Moor exit to level 3;
// otherwise a valid looking zone could be joined to the wrong destination.
func TestEntryWorldV2RejectsMalformedSourceExit(t *testing.T) {
	runtime := coldPlainsRuntime(t, entryWorldV2Records())
	for _, test := range []struct {
		name  string
		warps []map[string]any
	}{
		{"missing", nil},
		{"duplicate", []map[string]any{levelThreeExit("east", 3), levelThreeExit("east", 3)}},
		{"wrong target", []map[string]any{levelThreeExit("east", 4)}},
		{"noncardinal", []map[string]any{levelThreeExit("northeast", 3)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := modruntime.Call(t.Context(), runtime.lua, "d2legacy.mapgen.entry_world", "cold_plains",
				float64(42), float64(0), bloodMoorExitFacts(test.warps...))
			if err == nil {
				t.Fatal("malformed source exit was accepted")
			}
		})
	}
}

// Every cardinal Blood Moor exit must select the opposing Cold Plains entry edge.
func TestEntryWorldV2UsesOppositeColdPlainsEdge(t *testing.T) {
	runtime := coldPlainsRuntime(t, entryWorldV2Records())
	for _, facing := range []struct{ source, entry string }{
		{"north", "south"}, {"east", "west"}, {"south", "north"}, {"west", "east"},
	} {
		t.Run(facing.source, func(t *testing.T) {
			value, err := modruntime.Call(t.Context(), runtime.lua, "d2legacy.mapgen.entry_world", "cold_plains",
				float64(42), float64(0), bloodMoorExitFacts(levelThreeExit(facing.source, 3)))
			if err != nil {
				t.Fatal(err)
			}
			facts, ok := value.(map[string]any)
			if !ok {
				t.Fatalf("Cold Plains facts have type %T", value)
			}
			warps, ok := facts["warps"].([]any)
			if !ok || len(warps) != 1 {
				t.Fatalf("Cold Plains warps = %#v", facts["warps"])
			}
			entry, ok := warps[0].(map[string]any)
			if !ok || entry["direction"] != facing.entry {
				t.Fatalf("Cold Plains entry = %#v, want %s", warps[0], facing.entry)
			}
		})
	}
}

// A generated entry world must use the admitted Blood Moor exit to add the
// only level-3 zone and a reciprocal second pair at every difficulty.
func TestEntryWorldV2SelectsUniqueBloodMoorExit(t *testing.T) {
	for difficulty := 0; difficulty <= 2; difficulty++ {
		world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV2Records(), 42, difficulty)
		if err != nil {
			t.Fatalf("difficulty %d: %v", difficulty, err)
		}
		if world.Version != 2 || len(world.Zones) != 3 || len(world.Seams) != 2 {
			t.Fatalf("difficulty %d: topology version=%d zones=%d seams=%d",
				difficulty, world.Version, len(world.Zones), len(world.Seams))
		}
		for _, level := range []int{1, 2, 3} {
			zone := world.Zones[level]
			if zone == nil || zone.Request().Seed != 42 || int(zone.Request().Difficulty) != difficulty {
				t.Fatalf("difficulty %d level %d: zone=%#v", difficulty, level, zone)
			}
		}
		if world.Seams[0].FirstLevel != 1 || world.Seams[0].SecondLevel != 2 ||
			world.Seams[1].FirstLevel != 2 || world.Seams[1].SecondLevel != 3 {
			t.Fatalf("difficulty %d: seam order=%#v", difficulty, world.Seams)
		}
		if world.Seams[1].FirstDirection != "east" || world.Seams[1].SecondDirection != "west" {
			t.Fatalf("difficulty %d: reciprocal level-3 seam=%#v", difficulty, world.Seams[1])
		}
	}
}

// V2 snapshot bytes must round-trip canonically and reject drift in any admitted zone or seam.
func TestEntryWorldV2SnapshotRejectsTopologyDrift(t *testing.T) {
	world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV2Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := world.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreEntryWorld(encoded)
	if err != nil {
		t.Fatal(err)
	}
	again, err := restored.Snapshot()
	if err != nil || !bytes.Equal(again, encoded) {
		t.Fatalf("v2 restore changed snapshot: %v", err)
	}
	for _, mutate := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"duplicate seam", func(snapshot map[string]any) {
			seams := snapshot["seams"].([]any)
			snapshot["seams"] = append(seams, seams[1])
		}},
		{"missing zone", func(snapshot map[string]any) {
			zones := snapshot["zones"].([]any)
			snapshot["zones"] = zones[:2]
		}},
		{"duplicate zone identity", func(snapshot map[string]any) {
			zones := snapshot["zones"].([]any)
			zones[2] = zones[1]
		}},
		{"seed drift", func(snapshot map[string]any) {
			zones := snapshot["zones"].([]any)
			zones[2].(map[string]any)["request"].(map[string]any)["seed"] = float64(43)
		}},
		{"difficulty drift", func(snapshot map[string]any) {
			zones := snapshot["zones"].([]any)
			zones[2].(map[string]any)["request"].(map[string]any)["difficulty"] = float64(2)
		}},
		{"out of bounds seam tile", func(snapshot map[string]any) {
			seams := snapshot["seams"].([]any)
			seams[1].(map[string]any)["SecondTileX"] = float64(9999)
		}},
		{"mismatched admitted warp", func(snapshot map[string]any) {
			seams := snapshot["seams"].([]any)
			seams[1].(map[string]any)["SecondTileX"] = float64(1)
		}},
		{"mismatched warp role", func(snapshot map[string]any) {
			zones := snapshot["zones"].([]any)
			warps := zones[1].(map[string]any)["warps"].([]any)
			warps[1].(map[string]any)["role"] = "other-exit"
		}},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			var snapshot map[string]any
			if err := json.Unmarshal(encoded, &snapshot); err != nil {
				t.Fatal(err)
			}
			mutate.edit(snapshot)
			corrupt, err := json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := RestoreEntryWorld(corrupt); err == nil {
				t.Fatal("malformed v2 topology restored")
			}
		})
	}
}

// A v1 checkpoint preserves its original two-zone envelope, bytes, and checksum after restore.
func TestEntryWorldV1SnapshotKeepsLegacyEnvelope(t *testing.T) {
	world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV2Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}
	legacySeam := world.Seams[0]
	legacySeam.TriggerRadius = 0
	legacySeam.MaxArrivalDistance = 0
	legacySeam.ArrivalInset = 0
	legacySeam.FirstWarpID = 0
	legacySeam.SecondWarpID = 0
	legacySeam.FirstWarpRole = ""
	legacySeam.SecondWarpRole = ""
	legacy, err := json.Marshal(entryWorldSnapshot{
		Version: 1, Town: world.Zones[1].Definition(), Wilderness: world.Zones[2].Definition(),
		Seam: legacySeam,
	})
	if err != nil {
		t.Fatal(err)
	}
	legacyWorld, err := RestoreEntryWorld(legacy)
	if err != nil {
		t.Fatal(err)
	}
	legacyAgain, err := legacyWorld.Snapshot()
	if err != nil || !bytes.Equal(legacyAgain, legacy) {
		t.Fatalf("v1 restore changed snapshot: %v", err)
	}
	if legacyWorld.Version != 1 || len(legacyWorld.Zones) != 2 || len(legacyWorld.Seams) != 1 {
		t.Fatalf("v1 restore gained topology: %#v", legacyWorld)
	}
	checksum, err := legacyWorld.Checksum()
	if err != nil || checksum != fmt.Sprintf("%x", sha256.Sum256(legacy)) {
		t.Fatalf("v1 checksum=%s, err=%v", checksum, err)
	}
}
