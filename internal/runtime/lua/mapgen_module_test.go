package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

type mapgenCatalogStub struct{ snapshot gamedata.Snapshot }

func (stub mapgenCatalogStub) Snapshot() (gamedata.Snapshot, error) { return stub.snapshot, nil }

func TestMapgenModuleExposesBloodMoorTownEdge(t *testing.T) {
	runtime := New()
	fixture := mapgenCatalogStub{snapshot: outdoorModuleFixture()}
	if err := runtime.RegisterModule(MapgenModule(fixture)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`local z=require("d2legacy.mapgen.native/v1").outdoor(2,42,"east"); assert(z.kind=="outdoor" and #z.rooms==100 and z.warps[1].direction=="west" and #z.structures>0); local bridges,open=0,0; for _,s in ipairs(z.structures) do if s.kind=="bridge" then bridges=bridges+1; if s.passable then open=open+1 end end end; assert(bridges==64 and open==48)`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func outdoorModuleFixture() gamedata.Snapshot {
	result := gamedata.Snapshot{LevelsByID: map[int]model.LevelData{2: {Id: 2, Act: 0, DrlgType: 3, LevelType: 2, SizeX: 80, SizeY: 80}}, LevelPresetByDef: map[int]model.LevelPreset{}, LevelTypes: []model.LevelType{{}, {}, {File1: "Act1/Outdoors/Outdoor1.dt1"}}}
	for _, def := range []int{17, 26, 27, 28, 29, 30, 35} {
		result.LevelPresetByDef[def] = model.LevelPreset{Def: def, SizeX: 8, SizeY: 8, Files: 1, File1: "Act1/Outdoors/fill.ds1", Dt1Mask: 1}
	}
	for _, def := range []int{26, 27, 28} {
		record := result.LevelPresetByDef[def]
		record.Files = 4
		record.File2 = "Act1/Outdoors/structure2.ds1"
		record.File3 = "Act1/Outdoors/structure3.ds1"
		record.File4 = "Act1/Outdoors/structure4.ds1"
		result.LevelPresetByDef[def] = record
	}
	return result
}
