package mapgen

import (
	"fmt"
	"testing"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

func outdoorFixture() gamedata.Snapshot {
	presets := map[int]model.LevelPreset{}
	for _, def := range []int{29, 30, 35} {
		presets[def] = model.LevelPreset{Def: def, SizeX: 8, SizeY: 8, Files: 1, File1: fmt.Sprintf("Act1/Outdoors/fill%d.ds1", def), Dt1Mask: 1, Populate: 1}
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
	if len(left.Stamps()) != 100 || len(left.Rooms()) != 100 || len(left.Links()) != 180 {
		t.Fatalf("grid = %d stamps, %d rooms, %d links", len(left.Stamps()), len(left.Rooms()), len(left.Links()))
	}
	warp := left.Warps()[0]
	if warp.Role != "town-entry" || warp.Direction != "west" || warp.X != 0 || warp.Y != 40 || warp.DestinationLevel != 1 {
		t.Fatalf("town warp = %#v", warp)
	}
}

func TestBloodMoorRejectsTownWithoutCardinalExit(t *testing.T) {
	_, err := NewActOneOutdoorGenerator(outdoorFixture()).GenerateFromTown(Request{Version: ContractVersion, Act: 1, LevelID: 2}, Stamp{Role: "act1-town"})
	if err == nil {
		t.Fatal("accepted town without a cardinal exit")
	}
}
