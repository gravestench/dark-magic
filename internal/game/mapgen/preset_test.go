package mapgen

import (
	"errors"
	"slices"
	"testing"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

func presetFixture() gamedata.Snapshot {
	return gamedata.Snapshot{
		LevelsByID:   map[int]model.LevelData{1: {Id: 1, Act: 0, DrlgType: 2, LevelType: 1, SizeX: 40, SizeY: 30}},
		LevelPresets: []model.LevelPreset{{Def: 7, LevelId: 1, Files: 2, File1: `Act1\Town\one.ds1`, File2: "Act1/Town/two.ds1", Dt1Mask: 5, Populate: 1}},
		LevelTypes:   []model.LevelType{{Name: "None"}, {Name: "Act 1 Town", File1: `Act1\Town\floor.dt1`, File2: "ignored.dt1", File3: "Act1/Town/wall.dt1"}},
	}
}

func TestPresetGeneratorBuildsTypedActOneRecipeDeterministically(t *testing.T) {
	request := Request{Version: ContractVersion, Seed: 11, Act: 1, LevelID: 1, Difficulty: Normal}
	left, err := NewPresetGenerator(presetFixture()).Generate(request)
	if err != nil {
		t.Fatal(err)
	}
	right, err := NewPresetGenerator(presetFixture()).Generate(request)
	if err != nil {
		t.Fatal(err)
	}
	leftSum, _ := left.Checksum()
	rightSum, _ := right.Checksum()
	if rightSum != leftSum {
		t.Fatal("same typed records and seed changed the zone")
	}
	stamp := left.Stamps()[0]
	if stamp.PresetDef != 7 || stamp.Width != 40 || stamp.Height != 30 || !stamp.Populate {
		t.Fatalf("stamp = %#v", stamp)
	}
	if !slices.Equal(stamp.TilePaths, []string{"data/global/tiles/Act1/Town/floor.dt1", "data/global/tiles/Act1/Town/wall.dt1"}) {
		t.Fatalf("tile paths = %#v", stamp.TilePaths)
	}
	if len(left.Trace()) != 3 {
		t.Fatalf("trace = %#v", left.Trace())
	}
}

func TestPresetGeneratorRejectsNonPresetLevel(t *testing.T) {
	fixture := presetFixture()
	level := fixture.LevelsByID[1]
	level.DrlgType = 1
	fixture.LevelsByID[1] = level
	_, err := NewPresetGenerator(fixture).Generate(Request{Version: ContractVersion, Act: 1, LevelID: 1})
	if !errors.Is(err, ErrRequest) {
		t.Fatalf("error = %v", err)
	}
}

func TestAssetPathNormalizesLegacySeparators(t *testing.T) {
	if got := assetPath(`data\global\tiles\Act1\Town\town.ds1`); got != "data/global/tiles/Act1/Town/town.ds1" {
		t.Fatalf("path = %q", got)
	}
}

func TestStaticTownPresetUsesEveryAuthoredVariantAndCarriesExit(t *testing.T) {
	fixture := presetFixture()
	fixture.LevelPresets[0].Files = 0
	fixture.LevelPresets[0].File1 = `Act1\Town\TownN1.ds1`
	fixture.LevelPresets[0].File2 = `Act1\Town\TownE1.ds1`
	fixture.LevelPresets[0].File3 = `Act1\Town\TownS1.ds1`
	fixture.LevelPresets[0].File4 = `Act1\Town\TownW1.ds1`

	seen := map[string]bool{}
	for seed := uint64(0); seed < 128 && len(seen) < 4; seed++ {
		zone, err := NewPresetGenerator(fixture).Generate(Request{Version: ContractVersion, Seed: seed, Act: 1, LevelID: 1})
		if err != nil {
			t.Fatal(err)
		}
		stamp := zone.Stamps()[0]
		seen[stamp.Role] = true
	}
	for _, role := range []string{"act1-town:exit-north", "act1-town:exit-east", "act1-town:exit-south", "act1-town:exit-west"} {
		if !seen[role] {
			t.Errorf("never selected %q; roles = %#v", role, seen)
		}
	}
}
