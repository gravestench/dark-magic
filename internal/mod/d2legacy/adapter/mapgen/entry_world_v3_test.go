package mapgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// entryWorldV3Records joins synthetic records for the four admitted recipes.
func entryWorldV3Records() entryRecords {
	records := entryWorldV2Records()
	stony := stonyFieldRecords()
	records[levelsPath] = append(records[levelsPath], stony[levelsPath]...)
	// Cold Plains already supplies the shared 29/30 outdoor presets and level type.
	return records
}

// New worlds must publish every reciprocal seam and a deterministic durable recipe at every difficulty.
func TestEntryWorldV3Generation(t *testing.T) {
	for difficulty := 0; difficulty <= 2; difficulty++ {
		world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV3Records(), 42, difficulty)
		if err != nil {
			t.Fatal(err)
		}
		if world.Version != 3 || len(world.Zones) != 4 || len(world.Seams) != 3 {
			t.Fatalf("topology version=%d zones=%d seams=%d", world.Version, len(world.Zones), len(world.Seams))
		}
		for level := 1; level <= 4; level++ {
			zone := world.Zones[level]
			if zone == nil || zone.Request().LevelID != level || zone.Request().Seed != 42 ||
				int(zone.Request().Difficulty) != difficulty {
				t.Fatalf("invalid level %d: %#v", level, zone)
			}
		}
		for index, seam := range world.Seams {
			if seam.FirstLevel != index+1 || seam.SecondLevel != index+2 || seam.FirstGenerated != (index > 0) ||
				seam.FirstDirection != "east" || seam.SecondDirection != "west" {
				t.Fatalf("invalid seam %d: %#v", index, seam)
			}
		}
		seam := world.Seams[2]
		if seam.FirstWarpRole != "stony-field-exit" || seam.SecondWarpRole != "cold-plains-entry" {
			t.Fatalf("wrong Stony Field roles: %#v", seam)
		}
		snapshot, err := world.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		restored, err := RestoreEntryWorld(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		got, err := restored.Snapshot()
		if err != nil || !bytes.Equal(snapshot, got) {
			t.Fatalf("restore changed bytes: %v", err)
		}
		repeated, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV3Records(), 42, difficulty)
		if err != nil {
			t.Fatal(err)
		}
		want, err := world.Checksum()
		if err != nil {
			t.Fatal(err)
		}
		checksum, err := repeated.Checksum()
		if err != nil || checksum != want {
			t.Fatalf("repeated checksum=%s want=%s: %v", checksum, want, err)
		}
	}
}

// Durable topology must reject drift before publishing any admitted zones to recovery.
func TestEntryWorldV3RejectsTopologyDrift(t *testing.T) {
	world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV3Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}
	if world.Version != 3 {
		t.Fatalf("version=%d, want 3", world.Version)
	}
	encoded, err := world.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	type envelope struct {
		Version int                   `json:"version"`
		Zones   []worldgen.Definition `json:"zones"`
		Seams   []SeamSpec            `json:"seams"`
	}
	type mutation struct {
		name string
		edit func(*envelope)
	}
	tests := []mutation{
		{"missing zone", func(s *envelope) { s.Zones = s.Zones[:3] }},
		{"extra zone", func(s *envelope) { s.Zones = append(s.Zones, s.Zones[3]) }},
		{"reordered zones", func(s *envelope) { s.Zones[2], s.Zones[3] = s.Zones[3], s.Zones[2] }},
		{"duplicate identity", func(s *envelope) { s.Zones[3].Request.LevelID = 3 }},
		{"seed drift", func(s *envelope) { s.Zones[3].Request.Seed++ }},
		{"difficulty drift", func(s *envelope) { s.Zones[3].Request.Difficulty = 2 }},
		{"invalid recipe", func(s *envelope) { s.Zones[3].Bounds.Width = 0 }},
		{"missing seam", func(s *envelope) { s.Seams = s.Seams[:2] }},
		{"extra seam", func(s *envelope) { s.Seams = append(s.Seams, s.Seams[2]) }},
		{"duplicate seam", func(s *envelope) { s.Seams[2] = s.Seams[1] }},
		{"reordered seams", func(s *envelope) { s.Seams[1], s.Seams[2] = s.Seams[2], s.Seams[1] }},
		{"generated town", func(s *envelope) { s.Seams[0].FirstGenerated = true }},
	}
	for index := 0; index < 3; index++ {
		tests = append(tests, mutation{fmt.Sprintf("invalid second role seam %d", index),
			func(s *envelope) { s.Seams[index].SecondWarpRole = "wrong" }})
	}
	for _, test := range []struct {
		name string
		edit func(*SeamSpec)
	}{
		{"authored generated seam", func(s *SeamSpec) { s.FirstGenerated = false }},
		{"wrong first level", func(s *SeamSpec) { s.FirstLevel++ }},
		{"wrong second level", func(s *SeamSpec) { s.SecondLevel = 2 }},
		{"first role", func(s *SeamSpec) { s.FirstWarpRole = "wrong" }},
		{"second role", func(s *SeamSpec) { s.SecondWarpRole = "wrong" }},
		{"first ID", func(s *SeamSpec) { s.FirstWarpID = 999 }},
		{"second ID", func(s *SeamSpec) { s.SecondWarpID = 999 }},
		{"first direction", func(s *SeamSpec) { s.FirstDirection = "north" }},
		{"second direction", func(s *SeamSpec) { s.SecondDirection = "north" }},
		{"opposite but wrong directions", func(s *SeamSpec) { s.FirstDirection = "north"; s.SecondDirection = "south" }},
		{"out of bounds first", func(s *SeamSpec) { s.FirstTileX = 9999 }},
		{"out of bounds second", func(s *SeamSpec) { s.SecondTileY = -1 }},
		{"mismatched first tile", func(s *SeamSpec) { s.FirstTileY++ }},
		{"mismatched second tile", func(s *SeamSpec) { s.SecondTileY++ }},
		{"zero trigger", func(s *SeamSpec) { s.TriggerRadius = 0 }},
		{"negative trigger", func(s *SeamSpec) { s.TriggerRadius = -1 }},
		{"zero arrival distance", func(s *SeamSpec) { s.MaxArrivalDistance = 0 }},
		{"zero inset", func(s *SeamSpec) { s.ArrivalInset = 0 }},
		{"overlapping reverse trigger", func(s *SeamSpec) { s.ArrivalInset = s.TriggerRadius }},
	} {
		for index := 1; index < 3; index++ {
			tests = append(tests, mutation{fmt.Sprintf("%s seam %d", test.name, index),
				func(s *envelope) { test.edit(&s.Seams[index]) }})
		}
	}
	for _, test := range []struct {
		name string
		edit func(*worldgen.Definition)
	}{
		{"missing endpoint", func(z *worldgen.Definition) { z.Warps = nil }},
		{"duplicate endpoint", func(z *worldgen.Definition) { z.Warps = append(z.Warps, z.Warps[0]) }},
		{"wrong endpoint destination", func(z *worldgen.Definition) { z.Warps[0].DestinationLevel = 1 }},
		{"wrong endpoint role", func(z *worldgen.Definition) { z.Warps[0].Role = "wrong" }},
		{"wrong endpoint direction", func(z *worldgen.Definition) { z.Warps[0].Direction = "north" }},
		{"wrong endpoint ID", func(z *worldgen.Definition) { z.Warps[0].ID = 99 }},
	} {
		tests = append(tests, mutation{test.name, func(s *envelope) { test.edit(&s.Zones[3]) }})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var snapshot envelope
			if err := json.Unmarshal(encoded, &snapshot); err != nil {
				t.Fatal(err)
			}
			test.edit(&snapshot)
			bad, err := json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := RestoreEntryWorld(bad)
			if err == nil || len(restored.Zones) != 0 || len(restored.Seams) != 0 {
				t.Fatalf("corrupt topology published: world=%#v err=%v", restored, err)
			}
		})
	}
}

// Different numeric IDs must not make an ambiguous generated seam role restorable.
func TestEntryWorldV3RejectsDuplicateEndpointRoles(t *testing.T) {
	world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV3Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := world.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	for _, endpoint := range []struct {
		level int
		role  string
	}{
		{2, "town-entry"}, {2, "next-level-exit"}, {3, "blood-moor-entry"},
		{3, "stony-field-exit"}, {4, "cold-plains-entry"},
	} {
		for _, drift := range []string{"same facts", "different coordinates", "different destination"} {
			t.Run(fmt.Sprintf("level %d %s/%s", endpoint.level, endpoint.role, drift), func(t *testing.T) {
				var snapshot entryWorldSnapshotV3
				if err := json.Unmarshal(encoded, &snapshot); err != nil {
					t.Fatal(err)
				}

				zone := &snapshot.Zones[endpoint.level-1]
				found := false

				for _, warp := range zone.Warps {
					if warp.Role != endpoint.role {
						continue
					}

					found = true

					warp.ID = 99

					switch drift {
					case "different coordinates":
						warp.Y++
					case "different destination":
						warp.DestinationLevel = 99
					}

					zone.Warps = append(zone.Warps, warp)

					break
				}

				if !found {
					t.Fatal("fixture lacks bound endpoint")
				}

				bad, err := json.Marshal(snapshot)
				if err != nil {
					t.Fatal(err)
				}

				restored, err := RestoreEntryWorld(bad)
				if err == nil || len(restored.Zones) != 0 || len(restored.Seams) != 0 {
					t.Fatalf("ambiguous endpoint published: version=%d err=%v", restored.Version, err)
				}
			})
		}
	}
}

// Role uniqueness is a v3 contract; older saved definitions retain their prior acceptance behavior.
func TestEntryWorldLegacyRetainsDuplicateEndpointRoles(t *testing.T) {
	for _, version := range []int{1, 2} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			world, err := RestoreEntryWorld(legacyEntryWorldBytes(t, version))
			if err != nil {
				t.Fatal(err)
			}

			definition := world.Zones[2].Definition()
			duplicate := definition.Warps[0]
			duplicate.ID = 99
			duplicate.Y++
			duplicate.DestinationLevel = 99
			definition.Warps = append(definition.Warps, duplicate)

			world.Zones[2], err = worldgen.NewZone(definition)
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
			if err != nil || !bytes.Equal(encoded, again) {
				t.Fatalf("legacy duplicate role changed bytes: %v", err)
			}
		})
	}
}

// Restore must keep saved policy decisions even after the installed Lua generator changes.
func TestEntryWorldV3RestoreDoesNotRegenerate(t *testing.T) {
	world, err := GenerateEntryWorld(t.Context(), content.D2Legacy(), entryWorldV3Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}
	if world.Version != 3 {
		t.Fatalf("version=%d, want 3", world.Version)
	}
	saved, err := world.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	want, err := world.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	changed := fstest.MapFS{}
	err = fs.WalkDir(content.D2Legacy(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(content.D2Legacy(), path)
		if err != nil {
			return err
		}
		changed[path] = &fstest.MapFile{Data: data}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	policy := changed["lua/d2legacy/mapgen/entry_world.lua"]
	if policy == nil {
		t.Fatal("entry policy missing")
	}
	policy.Data = bytes.Replace(policy.Data, []byte("arrival_inset = 6"), []byte("arrival_inset = 7"), 1)
	next, err := GenerateEntryWorld(t.Context(), changed, entryWorldV3Records(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}
	nextSum, err := next.Checksum()
	if err != nil || nextSum == want {
		t.Fatalf("changed policy did not change generation: %v", err)
	}
	restored, err := RestoreEntryWorld(saved)
	if err != nil {
		t.Fatal(err)
	}
	got, err := restored.Snapshot()
	if err != nil || !bytes.Equal(got, saved) {
		t.Fatalf("restored saved policy changed: %v", err)
	}
	sum, err := restored.Checksum()
	if err != nil || sum != want {
		t.Fatalf("restore checksum=%s want=%s: %v", sum, want, err)
	}
}
