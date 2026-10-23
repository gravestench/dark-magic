package mapgen

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

const (
	levelsPath  = "data/global/excel/Levels.txt"
	typesPath   = "data/global/excel/LvlTypes.txt"
	presetsPath = "data/global/excel/LvlPrest.txt"
)

// coldPlainsRecords supplies a complete, small fixture so each malformed-row
// case can change one admission rule without depending on owned game data.
func coldPlainsRecords() entryRecords {
	return entryRecords{
		levelsPath: {{
			"Id": "3", "Act": "0", "DrlgType": "3", "LevelType": "2",
			"SizeX": "80", "SizeY": "80",
			"SizeX(N)": "88", "SizeY(N)": "88",
			"SizeX(H)": "96", "SizeY(H)": "96",
		}},
		typesPath: {
			{}, {}, {"File 1": `Act1\Outdoors\Outdoor1.dt1`},
		},
		presetsPath: {
			{"Def": "29", "LevelId": "0", "Outdoors": "1", "SizeX": "8", "SizeY": "8",
				"Files": "1", "File1": `Act1\Outdoors\Circle.ds1`, "Dt1Mask": "1"},
			{"Def": "30", "LevelId": "0", "Outdoors": "1", "SizeX": "8", "SizeY": "8",
				"Files": "1", "File1": `Act1\Outdoors\Clover.ds1`, "Dt1Mask": "1"},
		},
	}
}

// coldPlainsRuntime closes with a live context because test cleanup runs after
// the test context is canceled.
func coldPlainsRuntime(t *testing.T, records entryRecords) *Runtime {
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

// TestColdPlainsProducesConnectedZone checks every mode and approach direction
// because the entry route must remain traversable after either choice changes.
func TestColdPlainsProducesConnectedZone(t *testing.T) {
	runtime := coldPlainsRuntime(t, coldPlainsRecords())
	for _, mode := range []struct {
		difficulty int
		size       int
	}{{0, 80}, {1, 88}, {2, 96}} {
		for _, facing := range []struct {
			source, entry string
		}{{"north", "south"}, {"east", "west"}, {"south", "north"}, {"west", "east"}} {
			zone, err := runtime.Generate(
				t.Context(), "cold_plains", float64(3), float64(42), facing.source, float64(mode.difficulty),
			)
			if err != nil {
				t.Fatalf("difficulty=%d source=%s: %v", mode.difficulty, facing.source, err)
			}
			if request := zone.Request(); request.LevelID != 3 || int(request.Difficulty) != mode.difficulty {
				t.Fatalf("request = %#v", request)
			}
			bounds := zone.Bounds()
			if bounds.Width != mode.size || bounds.Height != mode.size {
				t.Fatalf("difficulty=%d bounds=%#v, want %dx%d", mode.difficulty, bounds, mode.size, mode.size)
			}
			wantRooms := mode.size / 8 * (mode.size / 8)
			if len(zone.Rooms()) != wantRooms || len(zone.Stamps()) != wantRooms {
				t.Fatalf("difficulty=%d rooms/stamps=%d/%d, want %d",
					mode.difficulty, len(zone.Rooms()), len(zone.Stamps()), wantRooms)
			}
			if !coldPlainsRoomsConnected(zone) {
				t.Fatalf("difficulty=%d source=%s: room graph disconnected", mode.difficulty, facing.source)
			}
			warps := zone.Warps()
			if len(warps) != 1 {
				t.Fatalf("difficulty=%d source=%s: %d warps, want one", mode.difficulty, facing.source, len(warps))
			}
			warp := warps[0]
			if warp.Role != "blood-moor-entry" || warp.Direction != facing.entry || warp.DestinationLevel != 2 {
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

// TestColdPlainsChecksumSurvivesRestore keeps canonical replay separate from
// geometry checks, including a direction change that must alter the identity.
func TestColdPlainsChecksumSurvivesRestore(t *testing.T) {
	runtime := coldPlainsRuntime(t, coldPlainsRecords())
	first, err := runtime.Generate(t.Context(), "cold_plains", float64(3), float64(42), "east", float64(0))
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := runtime.Generate(t.Context(), "cold_plains", float64(3), float64(42), "east", float64(0))
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := runtime.Generate(t.Context(), "cold_plains", float64(3), float64(42), "north", float64(0))
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

// TestColdPlainsRejectsInvalidInputsAndRecords requires every source field to
// be valid before admission, even when the selected seed would not use it.
func TestColdPlainsRejectsInvalidInputsAndRecords(t *testing.T) {
	cases := []struct {
		name      string
		level     int
		direction string
		mutate    func(entryRecords)
	}{
		{name: "wrong level", level: 4, direction: "east"},
		{name: "empty direction", level: 3, direction: ""},
		{name: "diagonal direction", level: 3, direction: "northeast"},
		{name: "wrong act", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[levelsPath][0]["Act"] = "1" }},
		{name: "wrong drlg type", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[levelsPath][0]["DrlgType"] = "2" }},
		{name: "wrong level type", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[levelsPath][0]["LevelType"] = "3" }},
		{name: "misaligned size", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[levelsPath][0]["SizeX"] = "81" }},
		{name: "missing normal size", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[levelsPath][0]["SizeX"] = "" }},
		{name: "missing type row", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[typesPath] = rows[typesPath][:2] }},
		{name: "missing preset", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[presetsPath] = rows[presetsPath][:1] }},
		{name: "wrong preset level", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[presetsPath][0]["LevelId"] = "3" }},
		{name: "nonoutdoor preset", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[presetsPath][0]["Outdoors"] = "0" }},
		{name: "wrong preset width", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[presetsPath][0]["SizeX"] = "16" }},
		{name: "wrong preset height", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[presetsPath][0]["SizeY"] = "16" }},
		{name: "missing ds1 reference", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[presetsPath][0]["File1"] = "" }},
		{name: "missing dt1 reference", level: 3, direction: "east",
			mutate: func(rows entryRecords) { rows[typesPath][2]["File 1"] = "" }},
		{name: "missing declared ds1 variant", level: 3, direction: "east",
			mutate: func(rows entryRecords) {
				rows[presetsPath][0]["Files"] = "2"
				rows[presetsPath][0]["File2"] = ""
			}},
		{name: "missing selected dt1 tile", level: 3, direction: "east",
			mutate: func(rows entryRecords) {
				rows[presetsPath][0]["Dt1Mask"] = "3"
				rows[typesPath][2]["File 2"] = ""
			}},
		{name: "directory ds1 reference", level: 3, direction: "east",
			mutate: func(rows entryRecords) {
				rows[presetsPath][0]["File1"] = "/"
			}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			rows := coldPlainsRecords()
			if test.mutate != nil {
				test.mutate(rows)
			}
			runtime := coldPlainsRuntime(t, rows)
			if _, err := runtime.Generate(
				t.Context(), "cold_plains", float64(test.level), float64(42), test.direction, float64(0),
			); err == nil {
				t.Fatal("expected malformed Cold Plains request or records to be rejected")
			}
		})
	}
}

// coldPlainsWarpOnEdge checks the facing edge rather than any boundary cell.
func coldPlainsWarpOnEdge(warp worldgen.Warp, size int) bool {
	switch warp.Direction {
	case "north":
		return warp.Y == 0 && warp.X >= 0 && warp.X < size
	case "east":
		return warp.X == size-1 && warp.Y >= 0 && warp.Y < size
	case "south":
		return warp.Y == size-1 && warp.X >= 0 && warp.X < size
	case "west":
		return warp.X == 0 && warp.Y >= 0 && warp.Y < size
	default:
		return false
	}
}

// coldPlainsRoomsConnected checks that the admitted room links reach every room.
func coldPlainsRoomsConnected(zone *worldgen.Zone) bool {
	rooms := zone.Rooms()
	if len(rooms) == 0 {
		return false
	}
	neighbors := map[uint32][]uint32{}
	for _, link := range zone.Links() {
		neighbors[link.From] = append(neighbors[link.From], link.To)
		neighbors[link.To] = append(neighbors[link.To], link.From)
	}
	seen := map[uint32]bool{rooms[0].ID: true}
	queue := []uint32{rooms[0].ID}
	for len(queue) > 0 {
		first := queue[0]
		queue = queue[1:]
		for _, next := range neighbors[first] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return len(seen) == len(rooms)
}

// coldPlainsPathConnects checks semantic route continuity before materialization.
func coldPlainsPathConnects(paths []worldgen.PathTile, startX, startY, endX, endY int) bool {
	points := map[[2]int]bool{}
	for _, point := range paths {
		points[[2]int{point.X, point.Y}] = true
	}
	start, end := [2]int{startX, startY}, [2]int{endX, endY}
	if !points[start] || !points[end] {
		return false
	}
	seen := map[[2]int]bool{start: true}
	queue := [][2]int{start}
	for len(queue) > 0 {
		point := queue[0]
		queue = queue[1:]
		if point == end {
			return true
		}
		for _, delta := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			next := [2]int{point[0] + delta[0], point[1] + delta[1]}
			if points[next] && !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}
