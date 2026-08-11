package mapgen

import (
	"testing"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

func mazeFixture() gamedata.Snapshot {
	presets := make([]model.LevelPreset, 0, 15)
	byDef := make(map[int]model.LevelPreset, 15)
	for mask := 1; mask <= 15; mask++ {
		record := model.LevelPreset{Def: 52 + mask, SizeX: 24, SizeY: 24, Files: 1, File1: "Act1/Caves/room.ds1", Dt1Mask: 1}
		presets = append(presets, record)
		byDef[record.Def] = record
	}
	return gamedata.Snapshot{
		LevelsByID:       map[int]model.LevelData{9: {Id: 9, Act: 0, DrlgType: 1, LevelType: 3}},
		LevelMazeByLevel: map[int]model.LevelMazeData{9: {Level: 9, Rooms: 12, RoomsN: 14, RoomsH: 16, SizeX: 24, SizeY: 24, Merge: 500}},
		LevelPresets:     presets, LevelPresetByDef: byDef,
		LevelTypes: []model.LevelType{{}, {}, {}, {File1: "Act1/Caves/floor.dt1"}},
	}
}

func TestMazeGeneratorCreatesConnectedNonOverlappingCompatibleRooms(t *testing.T) {
	zone, err := NewMazeGenerator(mazeFixture()).Generate(Request{Version: ContractVersion, Seed: 42, Act: 1, LevelID: 9})
	if err != nil {
		t.Fatal(err)
	}
	rooms, links, stamps := zone.Rooms(), zone.Links(), zone.Stamps()
	if len(rooms) != 12 || len(links) < len(rooms)-1 || len(stamps) != len(rooms) {
		t.Fatalf("rooms=%d links=%d stamps=%d", len(rooms), len(links), len(stamps))
	}
	seen := map[[2]int]bool{}
	for index, room := range rooms {
		key := [2]int{room.X, room.Y}
		if seen[key] {
			t.Fatalf("overlapping room at %v", key)
		}
		seen[key] = true
		if room.Width != 24 || room.Height != 24 || stamps[index].PresetDef < 53 || stamps[index].PresetDef > 67 {
			t.Fatalf("room/stamp %d = %#v / %#v", index, room, stamps[index])
		}
	}
	visited := map[uint32]bool{1: true}
	for changed := true; changed; {
		changed = false
		for _, link := range links {
			if visited[link.From] && !visited[link.To] {
				visited[link.To], changed = true, true
			}
			if visited[link.To] && !visited[link.From] {
				visited[link.From], changed = true, true
			}
		}
	}
	if len(visited) != len(rooms) {
		t.Fatalf("only %d/%d rooms connected", len(visited), len(rooms))
	}
}

func TestMazeGeneratorIsChecksumStableAndDifficultyAware(t *testing.T) {
	generator := NewMazeGenerator(mazeFixture())
	request := Request{Version: ContractVersion, Seed: 99, Act: 1, LevelID: 9, Difficulty: Nightmare}
	left, err := generator.Generate(request)
	if err != nil {
		t.Fatal(err)
	}
	right, err := generator.Generate(request)
	if err != nil {
		t.Fatal(err)
	}
	leftSum, _ := left.Checksum()
	rightSum, _ := right.Checksum()
	if leftSum != rightSum || len(left.Rooms()) != 14 {
		t.Fatalf("checksums %s/%s rooms=%d", leftSum, rightSum, len(left.Rooms()))
	}
}

func TestConnectionMaskMapsDirectlyToAuthoredCaveDefinition(t *testing.T) {
	masks := map[mazeCell]uint8{}
	center := mazeCell{}
	for _, neighbor := range []mazeCell{{-1, 0}, {1, 0}, {0, 1}, {0, -1}} {
		applyConnection(masks, center, neighbor)
	}
	if masks[center] != 15 || 52+int(masks[center]) != 67 {
		t.Fatalf("center mask=%#x preset=%d", masks[center], 52+int(masks[center]))
	}
}
