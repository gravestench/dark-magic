package mapgen

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// Keep multiple fill variants in this fixture so a changed draw stream or
// variant order cannot hide behind a single-choice preset.
func entryWorldV3PolicyRecords() entryRecords {
	records := stonyFieldRecords()
	records[levelsPath] = append(records[levelsPath], coldPlainsRecords()[levelsPath]...)
	return records
}

// A misplaced v3 exit, a disconnected corridor, or a changed fill draw must
// fail without changing the legacy one-entry generation path.
func TestEntryWorldV3PolicyPlacesReciprocalOutdoorSeam(t *testing.T) {
	runtime := coldPlainsRuntime(t, entryWorldV3PolicyRecords())
	for _, mode := range []struct{ difficulty, size int }{{0, 80}, {1, 88}, {2, 96}} {
		for _, facing := range []struct{ source, entry string }{
			{"north", "south"}, {"east", "west"}, {"south", "north"}, {"west", "east"},
		} {
			t.Run(fmt.Sprintf("difficulty=%d/%s", mode.difficulty, facing.source), func(t *testing.T) {
				legacy, err := runtime.Generate(t.Context(), "cold_plains", float64(3), float64(42),
					facing.source, float64(mode.difficulty))
				if err != nil {
					t.Fatal(err)
				}
				if mode.difficulty == 0 && facing.source == "east" {
					hash, err := legacy.Checksum()
					if err != nil {
						t.Fatal(err)
					}
					if hash != "cd4d3610634e3c1e37ddce9aa6c9e29cc7f2b834d40d86668b39af3d0b94e993" {
						t.Fatalf("legacy recipe checksum changed: %s", hash)
					}
				}
				bloodMoor := bloodMoorExitFacts(levelThreeExit(facing.source, 3))
				oldPolicy, err := runtime.generateFrom(t.Context(), "d2legacy.mapgen.entry_world", "cold_plains",
					float64(42), float64(mode.difficulty), bloodMoor)
				if err != nil || !reflect.DeepEqual(oldPolicy.Definition(), legacy.Definition()) {
					t.Fatalf("legacy entry-world policy changed the direct recipe: %v", err)
				}
				plains, err := runtime.generateFrom(t.Context(), "d2legacy.mapgen.entry_world", "cold_plains_v3",
					float64(42), float64(mode.difficulty), bloodMoor)
				if err != nil {
					t.Fatal(err)
				}
				direct, err := runtime.generateFrom(t.Context(), "d2legacy.mapgen.cold_plains", "generate_with_stony_exit",
					float64(3), float64(42), facing.source, float64(mode.difficulty))
				if err != nil || !reflect.DeepEqual(direct.Definition(), plains.Definition()) {
					t.Fatalf("v3 policy changed the explicit generator recipe: %v", err)
				}
				if !reflect.DeepEqual(plains.Stamps(), legacy.Stamps()) ||
					!reflect.DeepEqual(plains.Rooms(), legacy.Rooms()) || !reflect.DeepEqual(plains.Links(), legacy.Links()) {
					t.Fatal("v3 changed Cold Plains fill, variants, or room topology")
				}
				oldWarps, warps := legacy.Warps(), plains.Warps()
				if len(oldWarps) != 1 || len(warps) != 2 || warps[0] != oldWarps[0] {
					t.Fatalf("legacy/v3 warps = %#v / %#v", oldWarps, warps)
				}
				entry, exit := warps[0], warps[1]
				if entry.Role != "blood-moor-entry" || entry.Direction != facing.entry || entry.DestinationLevel != 2 {
					t.Fatalf("Cold Plains entry = %#v", entry)
				}
				x, y := entryWorldV3PolicyMidpoint(facing.source, mode.size)
				if exit.ID == entry.ID || exit.Role != "stony-field-exit" || exit.Direction != facing.source ||
					exit.DestinationLevel != 4 || exit.X != x || exit.Y != y {
					t.Fatalf("Cold Plains exit = %#v, want %s at (%d,%d) targeting 4", exit, facing.source, x, y)
				}
				if !coldPlainsPathConnects(plains.Paths(), entry.X, entry.Y, mode.size/2, mode.size/2) ||
					!coldPlainsPathConnects(plains.Paths(), exit.X, exit.Y, mode.size/2, mode.size/2) {
					t.Fatal("Cold Plains dirt route does not join both endpoints to the interior")
				}
				for _, tile := range plains.Paths() {
					if tile.MainIndex != 0 || tile.SubIndex == 0 {
						t.Fatalf("route tile lacks authored dirt identity: %#v", tile)
					}
				}
				plainsFacts, err := zoneFacts(plains, "Cold Plains")
				if err != nil {
					t.Fatal(err)
				}
				stony, err := runtime.generateFrom(t.Context(), "d2legacy.mapgen.entry_world", "stony_field",
					float64(42), float64(mode.difficulty), plainsFacts)
				if err != nil {
					t.Fatal(err)
				}
				stonyWarps := stony.Warps()
				if len(stonyWarps) != 1 {
					t.Fatalf("Stony Field warps = %#v", stonyWarps)
				}
				stonyEntry := stonyWarps[0]
				sx, sy := entryWorldV3PolicyMidpoint(facing.entry, 80)
				if stonyEntry.Role != "cold-plains-entry" || stonyEntry.Direction != facing.entry ||
					stonyEntry.DestinationLevel != 3 || stonyEntry.X != sx || stonyEntry.Y != sy {
					t.Fatalf("Stony Field entry = %#v", stonyEntry)
				}
				stonyFacts, err := zoneFacts(stony, "Stony Field")
				if err != nil {
					t.Fatal(err)
				}
				value, err := modruntime.Call(t.Context(), runtime.lua, "d2legacy.mapgen.entry_world", "stony_field_seam",
					plainsFacts, stonyFacts)
				if err != nil {
					t.Fatal(err)
				}
				seam, err := parseSeamFacts(value.(map[string]any))
				if err != nil {
					t.Fatal(err)
				}
				want := SeamSpec{
					FirstGenerated: true, FirstLevel: 3, SecondLevel: 4,
					FirstDirection: facing.source, SecondDirection: facing.entry,
					FirstTileX: x, FirstTileY: y, SecondTileX: sx, SecondTileY: sy,
					FirstWarpID: int(exit.ID), FirstWarpRole: "stony-field-exit",
					SecondWarpID: int(stonyEntry.ID), SecondWarpRole: "cold-plains-entry",
					TriggerRadius: 2, MaxArrivalDistance: 10, ArrivalInset: 6,
				}
				if seam != want {
					t.Fatalf("seam = %#v, want %#v", seam, want)
				}
			})
		}
	}
}

// Border midpoint expectations depend only on the synthetic level dimensions.
func entryWorldV3PolicyMidpoint(direction string, size int) (int, int) {
	switch direction {
	case "north":
		return size / 2, 0
	case "east":
		return size - 1, size / 2
	case "south":
		return size / 2, size - 1
	default:
		return 0, size / 2
	}
}

// Both generation and seam description must refuse an ambiguous or
// misdirected source, rather than silently connecting unrelated endpoints.
func TestEntryWorldV3PolicyRejectsMalformedExit(t *testing.T) {
	runtime := coldPlainsRuntime(t, entryWorldV3PolicyRecords())
	exit := func(direction string, target int) map[string]any {
		return map[string]any{"id": float64(2), "role": "stony-field-exit", "direction": direction,
			"x": float64(79), "y": float64(40), "destination_level": float64(target)}
	}
	stony := map[string]any{
		"request": map[string]any{"level_id": float64(4)},
		"warps": []any{map[string]any{"id": float64(1), "role": "cold-plains-entry", "direction": "west",
			"x": float64(0), "y": float64(40), "destination_level": float64(3)}},
	}
	for _, test := range []struct {
		name, message string
		warps         []any
	}{
		{"absent", "has no stony-field-exit", nil},
		{"duplicate", "has duplicate stony-field-exit", []any{exit("east", 4), exit("east", 4)}},
		{"mistargeted", "targets the wrong level", []any{exit("east", 3)}},
		{"noncardinal", "has non-cardinal stony-field-exit", []any{exit("northeast", 4)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			plains := map[string]any{"request": map[string]any{"level_id": float64(3)}, "warps": test.warps}
			for _, call := range []struct {
				function string
				args     []any
			}{
				{"stony_field", []any{float64(42), float64(0), plains}},
				{"stony_field_seam", []any{plains, stony}},
			} {
				_, err := modruntime.Call(t.Context(), runtime.lua, "d2legacy.mapgen.entry_world", call.function, call.args...)
				if err == nil || !strings.Contains(err.Error(), test.message) {
					t.Fatalf("%s error = %v, want %q", call.function, err, test.message)
				}
			}
		})
	}
}

// The reciprocal entry must also be unique, correctly targeted, and face the
// admitted source; binding an unrelated destination would corrupt the seam.
func TestEntryWorldV3PolicyRejectsMalformedReciprocalEntry(t *testing.T) {
	runtime := coldPlainsRuntime(t, entryWorldV3PolicyRecords())
	plains := map[string]any{
		"request": map[string]any{"level_id": float64(3)},
		"warps": []any{map[string]any{"id": float64(2), "role": "stony-field-exit", "direction": "east",
			"x": float64(79), "y": float64(40), "destination_level": float64(4)}},
	}
	for _, test := range []struct {
		name, direction, message string
		target, count            int
	}{
		{"absent", "west", "has no cold-plains-entry", 3, 0},
		{"duplicate", "west", "has duplicate cold-plains-entry", 3, 2},
		{"mistargeted", "west", "targets the wrong level", 2, 1},
		{"misfacing", "east", "edges do not face one another", 3, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			warps := make([]any, test.count)
			for i := range warps {
				warps[i] = map[string]any{"id": float64(i + 1), "role": "cold-plains-entry", "direction": test.direction,
					"x": float64(0), "y": float64(40), "destination_level": float64(test.target)}
			}
			stony := map[string]any{"request": map[string]any{"level_id": float64(4)}, "warps": warps}
			if _, err := modruntime.Call(t.Context(), runtime.lua, "d2legacy.mapgen.entry_world", "stony_field_seam",
				plains, stony); err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("seam error = %v, want %q", err, test.message)
			}
		})
	}
}

// Describing an already admitted seam must preserve its endpoint identities
// and off-center coordinates rather than inventing new midpoint warps.
func TestEntryWorldV3PolicyBindsAdmittedEndpoints(t *testing.T) {
	runtime := coldPlainsRuntime(t, entryWorldV3PolicyRecords())
	plains := map[string]any{
		"request": map[string]any{"level_id": float64(3)},
		"warps": []any{map[string]any{"id": float64(23), "role": "stony-field-exit", "direction": "south",
			"x": float64(27), "y": float64(79), "destination_level": float64(4)}},
	}
	stony := map[string]any{
		"request": map[string]any{"level_id": float64(4)},
		"warps": []any{map[string]any{"id": float64(42), "role": "cold-plains-entry", "direction": "north",
			"x": float64(17), "y": float64(0), "destination_level": float64(3)}},
	}

	value, err := modruntime.Call(t.Context(), runtime.lua,
		"d2legacy.mapgen.entry_world", "stony_field_seam", plains, stony)
	if err != nil {
		t.Fatal(err)
	}
	seam, err := parseSeamFacts(value.(map[string]any))
	if err != nil {
		t.Fatal(err)
	}
	want := SeamSpec{
		FirstGenerated: true, FirstLevel: 3, SecondLevel: 4,
		FirstDirection: "south", SecondDirection: "north",
		FirstTileX: 27, FirstTileY: 79, SecondTileX: 17, SecondTileY: 0,
		FirstWarpID: 23, FirstWarpRole: "stony-field-exit",
		SecondWarpID: 42, SecondWarpRole: "cold-plains-entry",
		TriggerRadius: 2, MaxArrivalDistance: 10, ArrivalInset: 6,
	}
	if seam != want {
		t.Fatalf("seam = %#v, want %#v", seam, want)
	}
}
