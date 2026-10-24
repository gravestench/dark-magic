package mapgen

import (
	"context"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

// stonyFieldRecords supplies synthetic rows and asset identities, independent
// of owned game data or a claim about the runtime table version.
func stonyFieldRecords() entryRecords {
	return entryRecords{
		levelsPath: {{
			"Id": "4", "Act": "0", "DrlgType": "3", "LevelType": "2",
			"SizeX": "80", "SizeY": "80",
			"SizeX(N)": "80", "SizeY(N)": "80",
			"SizeX(H)": "80", "SizeY(H)": "80",
		}},
		typesPath: {
			{}, {}, {"File 1": `Act1\Outdoors\Outdoor1.dt1`},
		},
		presetsPath: {
			{"Def": "29", "LevelId": "0", "Outdoors": "1", "SizeX": "8", "SizeY": "8",
				"Files": "3", "File1": `Act1\Outdoors\Circle.ds1`, "File2": `Act1\Outdoors\Circle2.ds1`, "File3": `Act1\Outdoors\Circle3.ds1`, "Dt1Mask": "1"},
			{"Def": "30", "LevelId": "0", "Outdoors": "1", "SizeX": "8", "SizeY": "8",
				"Files": "3", "File1": `Act1\Outdoors\Clover.ds1`, "File2": `Act1\Outdoors\Clover2.ds1`, "File3": `Act1\Outdoors\Clover3.ds1`, "Dt1Mask": "1"},
		},
	}
}

// stonyFieldRuntime closes with a live context because test cleanup runs after
// the test context is canceled.
func stonyFieldRuntime(t *testing.T, records entryRecords) *Runtime {
	t.Helper()
	runtime, err := NewRuntime(t.Context(), content.D2Legacy(), records)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(context.Background()); err != nil {
			t.Errorf("close mapgen runtime: %v", err)
		}
	})
	return runtime
}

// TestStonyFieldProducesConnectedZone checks every mode and approach direction
// because the entry route must remain traversable after either choice changes.
func TestStonyFieldProducesConnectedZone(t *testing.T) {
	runtime := stonyFieldRuntime(t, stonyFieldRecords())
	for _, mode := range []struct {
		difficulty int
		size       int
	}{{0, 80}, {1, 80}, {2, 80}} {
		for _, facing := range []struct {
			source, entry string
		}{{"north", "south"}, {"east", "west"}, {"south", "north"}, {"west", "east"}} {
			zone, err := runtime.Generate(
				t.Context(), "stony_field", float64(4), float64(42), facing.source, float64(mode.difficulty),
			)
			if err != nil {
				t.Fatalf("difficulty=%d source=%s: %v", mode.difficulty, facing.source, err)
			}
			if request := zone.Request(); request.LevelID != 4 || request.Seed != 42 ||
				request.Act != 1 || request.Version != 1 || int(request.Difficulty) != mode.difficulty {
				t.Fatalf("request = %#v", request)
			}
			bounds := zone.Bounds()
			if bounds.X != 0 || bounds.Y != 0 || bounds.Width != mode.size || bounds.Height != mode.size {
				t.Fatalf("difficulty=%d bounds=%#v, want %dx%d", mode.difficulty, bounds, mode.size, mode.size)
			}
			wantRooms := mode.size / 8 * (mode.size / 8)
			if len(zone.Rooms()) != wantRooms || len(zone.Stamps()) != wantRooms {
				t.Fatalf("difficulty=%d rooms/stamps=%d/%d, want %d",
					mode.difficulty, len(zone.Rooms()), len(zone.Stamps()), wantRooms)
			}
			stamps := zone.Stamps()
			for i, room := range zone.Rooms() {
				stamp := stamps[i]
				if room.ID != uint32(i+1) || stamp.ID != uint32(i+1) || room.StampID != stamp.ID ||
					room.X != (i%10)*8 || room.Y != (i/10)*8 || room.Width != 8 || room.Height != 8 ||
					stamp.X != room.X || stamp.Y != room.Y || stamp.Width != 8 || stamp.Height != 8 ||
					stamp.Role != "stony-field-fill" || (stamp.PresetDef != 29 && stamp.PresetDef != 30) {
					t.Fatalf("room/stamp %d = %#v / %#v", i, room, stamp)
				}
			}
			trace := strings.Join(zone.Trace(), "\n")
			for _, evidence := range []string{"inspected 1.14d", "Levels", "LvlTypes", "LvlPrest", "29", "30", "generic 8x8", "three DS1 variants", "Dark Magic", "border endpoint", "does not establish their version"} {
				if !strings.Contains(trace, evidence) {
					t.Fatalf("trace lacks %q: %s", evidence, trace)
				}
			}
			if !coldPlainsRoomsConnected(zone) {
				t.Fatalf("difficulty=%d source=%s: room graph disconnected", mode.difficulty, facing.source)
			}
			warps := zone.Warps()
			if len(warps) != 1 {
				t.Fatalf("difficulty=%d source=%s: %d warps, want one", mode.difficulty, facing.source, len(warps))
			}
			warp := warps[0]
			if warp.Role != "cold-plains-entry" || warp.Direction != facing.entry || warp.DestinationLevel != 3 {
				t.Fatalf("difficulty=%d source=%s: warp=%#v", mode.difficulty, facing.source, warp)
			}
			if !coldPlainsWarpOnEdge(warp, mode.size) {
				t.Fatalf("difficulty=%d source=%s: warp=%#v is off edge", mode.difficulty, facing.source, warp)
			}
			if !coldPlainsPathConnects(zone.Paths(), warp.X, warp.Y, mode.size/2, mode.size/2) {
				t.Fatalf("difficulty=%d source=%s: semantic path does not connect entry to center",
					mode.difficulty, facing.source)
			}
		}
	}
}

// TestStonyFieldChecksumSurvivesRestore keeps canonical replay separate from
// geometry checks, including a direction change that must alter the identity.
func TestStonyFieldChecksumSurvivesRestore(t *testing.T) {
	runtime := stonyFieldRuntime(t, stonyFieldRecords())
	first, err := runtime.Generate(t.Context(), "stony_field", float64(4), float64(42), "east", float64(0))
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := runtime.Generate(t.Context(), "stony_field", float64(4), float64(42), "east", float64(0))
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := runtime.Generate(t.Context(), "stony_field", float64(4), float64(42), "north", float64(0))
	if err != nil {
		t.Fatal(err)
	}
	restored, err := worldgen.NewZone(first.Definition())
	if err != nil {
		t.Fatal(err)
	}
	firstHash, err := first.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	for label, other := range map[string]*worldgen.Zone{"repeated": repeated, "restored": restored} {
		hash, err := other.Checksum()
		if err != nil {
			t.Fatal(err)
		}
		if hash != firstHash {
			t.Fatalf("%s checksum=%s, want %s", label, hash, firstHash)
		}
	}
	rotatedHash, err := rotated.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	if rotatedHash == firstHash {
		t.Fatal("changing only source direction did not change the zone checksum")
	}
}
