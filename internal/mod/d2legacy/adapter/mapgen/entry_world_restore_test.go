package mapgen

import (
	"fmt"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

type entryRecords map[string][]map[string]string

func (records entryRecords) Load(path string) ([]map[string]string, error) { return records[path], nil }
func (entryRecords) Invalidate(string)                                     {}
func (entryRecords) Loaded(string) bool                                    { return true }

func TestLuaSelectedEntryWorldRestoresToSameTopologyChecksum(t *testing.T) {
	records := entryRecords{
		"data/global/excel/Levels.txt": {
			{"Id": "1", "Act": "0", "DrlgType": "2", "LevelType": "1", "SizeX": "40", "SizeY": "30", "SizeX(H)": "48", "SizeY(H)": "38"},
			{"Id": "2", "Act": "0", "DrlgType": "3", "LevelType": "2", "SizeX": "80", "SizeY": "80", "SizeX(H)": "96", "SizeY(H)": "96"},
		},
		"data/global/excel/LvlTypes.txt": {
			{}, {"File 1": `Act1\Town\floor.dt1`}, {"File 1": `Act1\Outdoors\Outdoor1.dt1`},
		},
		"data/global/excel/LvlPrest.txt": {{
			"Def": "7", "LevelId": "1", "Files": "1", "File1": `Act1\Town\TownE1.ds1`,
			"Dt1Mask": "1", "Populate": "1", "Logicals": "1",
		}},
	}
	for _, definition := range []int{17, 26, 27, 28, 29, 30, 35} {
		row := map[string]string{"Def": fmt.Sprint(definition), "SizeX": "8", "SizeY": "8",
			"Files": "1", "File1": fmt.Sprintf("Act1/Outdoors/fill%d.ds1", definition),
			"Dt1Mask": "1", "Populate": "1"}
		if definition == 26 || definition == 27 || definition == 28 {
			row["Files"] = "4"
			for variant := 2; variant <= 4; variant++ {
				row[fmt.Sprintf("File%d", variant)] = fmt.Sprintf("Act1/Outdoors/structure%d-%d.ds1", definition, variant)
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
		t.Fatalf("entry-world difficulty = town %d, wilderness %d", generated.Town.Request().Difficulty, generated.Wilderness.Request().Difficulty)
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
}

func TestGenerateEntryWorldRejectsInvalidDifficulty(t *testing.T) {
	if _, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryRecords{}, 42, 3); err == nil {
		t.Fatal("expected invalid difficulty to be rejected")
	}
}
