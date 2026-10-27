package mapgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

// entryWorldV4Records extends the synthetic outdoor rows with reciprocal Den facts for v4 tests.
func entryWorldV4Records() entryRecords {
	r := entryWorldV3Records()
	for _, row := range r[levelsPath] {
		if row["Id"] == "2" {
			row["Vis3"] = "8"
			row["Warp3"] = "0"
		}
	}
	r[levelsPath] = append(r[levelsPath], map[string]string{
		"Id": "8", "Act": "0", "DrlgType": "1", "LevelType": "3", "Vis0": "2", "Warp0": "4",
	})
	r[typesPath] = append(r[typesPath], map[string]string{"File 1": "Act1/Caves/cave.dt1"})
	r["data/global/excel/LvlMaze.txt"] = []map[string]string{{
		"Level": "8", "SizeX": "24", "SizeY": "24",
		"Rooms": "1", "Rooms(N)": "1", "Rooms(H)": "1",
	}}
	for _, id := range []int{51, 83, 84, 85, 86} {
		size := "24"
		if id == 51 {
			size = "8"
		}
		row := map[string]string{"Def": fmt.Sprint(id), "SizeX": size, "SizeY": size, "Files": "4", "Dt1Mask": "1"}
		for i := 1; i <= 4; i++ {
			row[fmt.Sprintf("File%d", i)] = fmt.Sprintf("Act1/fixture%d-%d.ds1", id, i)
		}
		r[presetsPath] = append(r[presetsPath], row)
	}
	return r
}

// V4 must publish the Lua-composed nonconsecutive topology and retain canonical bytes.
func TestEntryWorldV4RoundTrip(t *testing.T) {
	for d := 0; d < 3; d++ {
		w, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV4Records(), 42, d)
		if err != nil {
			t.Fatal(err)
		}
		if w.Version != 4 || len(w.Zones) != 5 || len(w.Seams) != 3 {
			t.Fatalf("version/zones/seams=%d/%d/%d", w.Version, len(w.Zones), len(w.Seams))
		}
		for _, id := range []int{1, 2, 3, 4, 8} {
			if w.Zones[id] == nil || w.Zones[id].Request().LevelID != id {
				t.Fatalf("missing level %d", id)
			}
		}
		b, err := w.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		r, err := RestoreEntryWorld(b)
		if err != nil {
			t.Fatal(err)
		}
		again, err := r.Snapshot()
		if err != nil || !bytes.Equal(b, again) {
			t.Fatalf("canonical restore: %v", err)
		}
		a, _ := w.Checksum()
		c, _ := r.Checksum()
		if a != c {
			t.Fatal("checksum drift")
		}
		repeated, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV4Records(), 42, d)
		if err != nil {
			t.Fatal(err)
		}
		repeatedBytes, err := repeated.Snapshot()
		if err != nil || !bytes.Equal(b, repeatedBytes) {
			t.Fatalf("repeat changed canonical v4 bytes: %v", err)
		}
		var envelope map[string]any
		if err := json.Unmarshal(b, &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope["cave_warp"] == nil {
			t.Fatal("missing reciprocal cave spec")
		}
	}
}

// Corruption must never publish an incomplete or mismatched cave topology.
func TestEntryWorldV4RejectsCaveDrift(t *testing.T) {
	w, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV4Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}
	b, err := w.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if w.Version != 4 {
		t.Fatalf("version=%d", w.Version)
	}
	for _, test := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"missing Den", func(s map[string]any) { s["zones"] = s["zones"].([]any)[:4] }},
		{"wrong Den identity", func(s map[string]any) {
			s["zones"].([]any)[4].(map[string]any)["request"].(map[string]any)["level_id"] = float64(5)
		}},
		{"changed Den extent", func(s map[string]any) {
			s["zones"].([]any)[4].(map[string]any)["bounds"].(map[string]any)["width"] = float64(32)
		}},
		{"missing pair", func(s map[string]any) { delete(s, "cave_warp") }},
		{"wrong slot", func(s map[string]any) { s["cave_warp"].(map[string]any)["first_vis_slot"] = float64(0) }},
		{"wrong destination", func(s map[string]any) {
			s["zones"].([]any)[4].(map[string]any)["warps"].([]any)[0].(map[string]any)["destination_level"] = float64(3)
		}},
		{"missing endpoint", func(s map[string]any) { s["zones"].([]any)[4].(map[string]any)["warps"] = []any{} }},
		{"duplicate role", func(s map[string]any) {
			z := s["zones"].([]any)[4].(map[string]any)
			w := z["warps"].([]any)
			copy := map[string]any{}
			for k, v := range w[0].(map[string]any) {
				copy[k] = v
			}
			copy["id"] = float64(99)
			z["warps"] = append(w, copy)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var s map[string]any
			if err := json.Unmarshal(b, &s); err != nil {
				t.Fatal(err)
			}
			test.edit(s)
			bad, _ := json.Marshal(s)
			got, err := RestoreEntryWorld(bad)
			if err == nil || len(got.Zones) > 0 {
				t.Fatalf("published corrupt topology: %v", err)
			}
		})
	}
}
