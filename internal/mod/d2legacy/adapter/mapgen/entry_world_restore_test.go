package mapgen

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

type entryRecords map[string][]map[string]string

// Load returns deterministic in-memory record rows so the test exercises production Lua policy without external data.
func (records entryRecords) Load(path string) ([]map[string]string, error) {
	return records[path], nil
}

// Invalidate is intentionally inert because test records never change during a generation.
func (entryRecords) Invalidate(string) {}

// Loaded reports all fixture tables as available, matching the eager in-memory test boundary.
func (entryRecords) Loaded(string) bool { return true }

// TestLuaSelectedEntryWorldRestoresToSameTopologyChecksum verifies that policy-selected recipes and seams survive the
// durable snapshot boundary exactly, including difficulty and canonical checksum identity.
func TestLuaSelectedEntryWorldRestoresToSameTopologyChecksum(t *testing.T) {
	records := entryRecords{
		"data/global/excel/Levels.txt": {
			{
				"Id": "1", "Act": "0", "DrlgType": "2", "LevelType": "1",
				"SizeX": "40", "SizeY": "30", "SizeX(H)": "48", "SizeY(H)": "38",
			},
			{
				"Id": "2", "Act": "0", "DrlgType": "3", "LevelType": "2",
				"SizeX": "80", "SizeY": "80", "SizeX(H)": "96", "SizeY(H)": "96",
			},
		},
		"data/global/excel/LvlTypes.txt": {
			{}, {"File 1": `Act1\Town\floor.dt1`}, {"File 1": `Act1\Outdoors\Outdoor1.dt1`}, {"File 1": "Act1/Caves/floor.dt1"},
		},
		"data/global/excel/LvlMaze.txt": {{"Level": "8", "Rooms": "1", "Rooms(H)": "8", "SizeX": "24", "SizeY": "24", "Merge": "0"}},
		"data/global/excel/LvlPrest.txt": {{
			"Def": "7", "LevelId": "1", "Files": "1", "File1": `Act1\Town\TownE1.ds1`,
			"Dt1Mask": "1", "Populate": "1", "Logicals": "1",
		}},
	}

	records["data/global/excel/Levels.txt"] = append(records["data/global/excel/Levels.txt"], map[string]string{"Id": "8", "Act": "0", "DrlgType": "1", "LevelType": "3"})
	for definition := 53; definition <= 98; definition++ {
		records["data/global/excel/LvlPrest.txt"] = append(records["data/global/excel/LvlPrest.txt"], map[string]string{
			"Def": fmt.Sprint(definition), "SizeX": "24", "SizeY": "24", "Files": "1", "File1": "Act1/Caves/room.ds1", "Dt1Mask": "1", "Populate": "1",
		})
	}
	for _, definition := range []int{17, 26, 27, 28, 29, 30, 35, 51} {
		row := map[string]string{
			"Def": fmt.Sprint(definition), "SizeX": "8", "SizeY": "8",
			"Files": "1", "File1": fmt.Sprintf("Act1/Outdoors/fill%d.ds1", definition),
			"Dt1Mask": "1", "Populate": "1",
		}

		if definition == 26 || definition == 27 || definition == 28 {
			row["Files"] = "4"

			for variant := 2; variant <= 4; variant++ {
				path := fmt.Sprintf("Act1/Outdoors/structure%d-%d.ds1", definition, variant)
				row[fmt.Sprintf("File%d", variant)] = path
			}
		}

		records["data/global/excel/LvlPrest.txt"] = append(records["data/global/excel/LvlPrest.txt"], row)
	}

	generated, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), records, 42, 2)
	if err != nil {
		t.Fatal(err)
	}

	if generated.Seam.FirstLevel != 1 || generated.Seam.SecondLevel != 2 ||
		generated.Seam.FirstDirection != "east" || generated.Seam.SecondDirection != "west" {
		t.Fatalf("Lua-selected entry seam = %#v", generated.Seam)
	}

	if generated.Town.Request().Difficulty != 2 || generated.Wilderness.Request().Difficulty != 2 {
		t.Fatalf(
			"entry-world difficulty = town %d, wilderness %d",
			generated.Town.Request().Difficulty,
			generated.Wilderness.Request().Difficulty,
		)
	}

	want, err := generated.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := generated.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	restored, err := RestoreEntryWorld(snapshot)
	if err != nil {
		t.Fatal(err)
	}

	got, err := restored.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("restored entry topology checksum = %s, want %s", got, want)
	}
	if restored.Dungeon == nil || restored.Dungeon.Request().LevelID != 8 || restored.Dungeon.Request().Difficulty != 2 || restored.DungeonConnection == nil {
		t.Fatal("dungeon recipe/connection were not restored")
	}
	for _, stamp := range restored.Dungeon.Stamps() {
		if stamp.Role == "next-level" {
			t.Fatal("Den has a downward exit")
		}
	}
	for _, mutate := range []func(*entryWorldSnapshot){
		func(s *entryWorldSnapshot) { s.DungeonConnection.FirstStamp = 0 },
		func(s *entryWorldSnapshot) { s.DungeonConnection.SecondLevel = 3 },
		func(s *entryWorldSnapshot) { s.Dungeon = nil },
		func(s *entryWorldSnapshot) { s.DungeonConnection = nil },
	} {
		var invalid entryWorldSnapshot
		if err := json.Unmarshal(snapshot, &invalid); err != nil {
			t.Fatal(err)
		}
		mutate(&invalid)
		encoded, err := json.Marshal(invalid)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := RestoreEntryWorld(encoded); err == nil {
			t.Fatal("accepted invalid dungeon connection")
		}
	}
	generated.Dungeon, generated.DungeonConnection = nil, nil
	legacy, err := generated.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restoredLegacy, err := RestoreEntryWorld(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if restoredLegacy.Dungeon != nil || restoredLegacy.DungeonConnection != nil {
		t.Fatal("invented Den in old two-area snapshot")
	}
	legacyChecksum, err := generated.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	restoredChecksum, err := restoredLegacy.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	if legacyChecksum != restoredChecksum {
		t.Fatal("old snapshot checksum changed")
	}
}

// TestGenerateEntryWorldRejectsInvalidDifficulty verifies that unsupported modes fail before starting the Lua runtime
// or reading recovered records.
func TestGenerateEntryWorldRejectsInvalidDifficulty(t *testing.T) {
	if _, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryRecords{}, 42, 3); err == nil {
		t.Fatal("expected invalid difficulty to be rejected")
	}
}
