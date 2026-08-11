package mapgen

import (
	"fmt"
	"testing"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

func outdoorFixture() gamedata.Snapshot {
	presets := map[int]model.LevelPreset{}
	for _, def := range []int{26, 27, 28, 29, 30, 35} {
		presets[def] = model.LevelPreset{Def: def, SizeX: 8, SizeY: 8, Files: 1, File1: fmt.Sprintf("Act1/Outdoors/fill%d.ds1", def), Dt1Mask: 1, Populate: 1}
	}
	for _, def := range []int{26, 27, 28} {
		preset := presets[def]
		preset.Files = 4
		preset.File2 = fmt.Sprintf("Act1/Outdoors/structure%d-2.ds1", def)
		preset.File3 = fmt.Sprintf("Act1/Outdoors/structure%d-3.ds1", def)
		preset.File4 = fmt.Sprintf("Act1/Outdoors/structure%d-4.ds1", def)
		presets[def] = preset
	}
	return gamedata.Snapshot{
		LevelsByID:       map[int]model.LevelData{2: {Id: 2, Act: 0, DrlgType: 3, LevelType: 2, SizeX: 80, SizeY: 80}},
		LevelPresetByDef: presets,
		LevelTypes:       []model.LevelType{{Name: "None"}, {Name: "Town"}, {Name: "Act 1 Wilderness", File1: "Act1/Outdoors/Outdoor1.dt1"}},
	}
}

func TestBloodMoorBuildsDeterministicCoarseGridJoinedToTown(t *testing.T) {
	request := Request{Version: ContractVersion, Seed: 42, Act: 1, LevelID: 2}
	generator := NewActOneOutdoorGenerator(outdoorFixture())
	left, err := generator.GenerateFromTown(request, Stamp{Role: "act1-town:exit-east"})
	if err != nil {
		t.Fatal(err)
	}
	right, err := generator.GenerateFromTown(request, Stamp{Role: "act1-town:exit-east"})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := left.Checksum()
	b, _ := right.Checksum()
	if a != b {
		t.Fatal("same Blood Moor request changed")
	}
	if len(left.Stamps()) != 120 || len(left.Rooms()) != 100 || len(left.Links()) != 180 {
		t.Fatalf("grid = %d stamps, %d rooms, %d links", len(left.Stamps()), len(left.Rooms()), len(left.Links()))
	}
	warp := left.Warps()[0]
	if warp.Role != "town-entry" || warp.Direction != "west" || warp.X != 0 || warp.Y != 40 || warp.DestinationLevel != 1 {
		t.Fatalf("town warp = %#v", warp)
	}
	next := left.Warps()[1]
	if next.Role != "next-level-exit" || next.Direction != "east" || next.X != 79 || next.Y != 40 || next.DestinationLevel != 3 {
		t.Fatalf("next-level warp = %#v", next)
	}
	routeCells := 0
	for _, stamp := range left.Stamps() {
		if stamp.Role == "blood-moor-route" {
			routeCells++
		}
	}
	if routeCells != 10 {
		t.Fatalf("route cells = %d, want 10", routeCells)
	}
	paths := left.Paths()
	pathSet := make(map[PathTile]bool, len(paths))
	for _, tile := range paths {
		pathSet[tile] = true
	}
	if !pathSet[PathTile{X: warp.X, Y: warp.Y}] || !pathSet[PathTile{X: next.X, Y: next.Y}] {
		t.Fatal("tile path does not join both authored edge anchors")
	}
	visited := map[PathTile]bool{}
	queue := []PathTile{{X: warp.X, Y: warp.Y}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nextTile := PathTile{X: current.X + dx, Y: current.Y + dy}
				if pathSet[nextTile] && !visited[nextTile] {
					queue = append(queue, nextTile)
				}
			}
		}
	}
	if len(visited) != len(paths) {
		t.Fatalf("connected path = %d/%d tiles", len(visited), len(paths))
	}
}

func TestBloodMoorRejectsTownWithoutCardinalExit(t *testing.T) {
	_, err := NewActOneOutdoorGenerator(outdoorFixture()).GenerateFromTown(Request{Version: ContractVersion, Act: 1, LevelID: 2}, Stamp{Role: "act1-town"})
	if err == nil {
		t.Fatal("accepted town without a cardinal exit")
	}
}

func TestOutdoorRouteIsContiguousAcrossEveryCardinalPair(t *testing.T) {
	for _, direction := range []string{"north", "east", "south", "west"} {
		route := outdoorRoute(42, 10, 10, direction)
		if len(route) != 10 {
			t.Fatalf("%s route cells = %d", direction, len(route))
		}
		horizontal := direction == "west" || direction == "east"
		forward := direction == "west" || direction == "north"
		previousCross := -1
		for step := 0; step < 10; step++ {
			axis := step
			if !forward {
				axis = 9 - step
			}
			cross := -1
			for cell := range route {
				cellAxis, cellCross := cell[1], cell[0]
				if horizontal {
					cellAxis, cellCross = cell[0], cell[1]
				}
				if cellAxis == axis {
					cross = cellCross
					break
				}
			}
			if cross < 0 || previousCross >= 0 && abs(cross-previousCross) > 1 {
				t.Fatalf("%s route is disconnected at step %d: %d -> %d", direction, step, previousCross, cross)
			}
			previousCross = cross
		}
	}
}

func TestBloodMoorStructuresKeepRoutePassableAcrossEveryCardinalPair(t *testing.T) {
	for _, direction := range []string{"north", "east", "south", "west"} {
		entry := townEdgeWarp(80, 80, oppositeDirection(direction))
		exit := nextLevelEdgeWarp(80, 80, entry.Direction)
		path := outdoorPathTiles(outdoorRoute(42, 10, 10, entry.Direction), 10, 10, entry, exit)
		pathSet := make(map[PathTile]bool, len(path))
		for _, tile := range path {
			pathSet[tile] = true
		}
		generator := NewActOneOutdoorGenerator(outdoorFixture())
		structures, stamps, path, err := generator.outdoorStructures(Request{Version: ContractVersion, Seed: 42, Act: 1, LevelID: 2}, outdoorFixture().LevelsByID[2], 10, 10, entry.Direction, path)
		if err != nil {
			t.Fatal(err)
		}
		pathSet = make(map[PathTile]bool, len(path))
		for _, tile := range path {
			pathSet[tile] = true
		}
		if len(stamps) != 20 {
			t.Fatalf("%s structure stamps = %d, want 20", direction, len(stamps))
		}
		bridges := 0
		pathBridges := 0
		passableBridges := 0
		for _, tile := range structures {
			onPath := pathSet[PathTile{X: tile.X, Y: tile.Y}]
			if onPath && !tile.Passable {
				t.Fatalf("%s route blocked by %s at %d,%d", direction, tile.Kind, tile.X, tile.Y)
			}
			if tile.Kind == "bridge" {
				bridges++
				if tile.Passable {
					passableBridges++
				}
				if onPath {
					pathBridges++
				}
				if onPath && !tile.Passable {
					t.Fatalf("%s bridge is not a passable route crossing: %#v", direction, tile)
				}
			}
		}
		if bridges != 64 {
			t.Fatalf("%s bridge footprint = %d tiles, want 64", direction, bridges)
		}
		if passableBridges != 48 {
			t.Fatalf("%s passable bridge footprint = %d tiles, want 48", direction, passableBridges)
		}
		if (entry.Direction == "west" || entry.Direction == "east") && pathBridges == 0 {
			t.Fatalf("%s route does not cross the bridge", direction)
		}
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
