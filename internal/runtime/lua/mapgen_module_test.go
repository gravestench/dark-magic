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

func TestMapgenModuleExposesMazeTopology(t *testing.T) {
	runtime := New()
	fixture := mapgenCatalogStub{snapshot: mazeModuleFixture()}
	if err := runtime.RegisterModule(MapgenModule(fixture)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`local z=require("d2legacy.mapgen.native/v1").maze(9,42); assert(z.kind=="maze" and #z.rooms==4 and #z.links>=3 and #z.stamps==4)`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

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

func mazeModuleFixture() gamedata.Snapshot {
	result := gamedata.Snapshot{
		LevelsByID:       map[int]model.LevelData{9: {Id: 9, DrlgType: 1, LevelType: 3}},
		LevelMazeByLevel: map[int]model.LevelMazeData{9: {Level: 9, Rooms: 4, RoomsN: 4, RoomsH: 4, SizeX: 24, SizeY: 24}},
		LevelPresetByDef: map[int]model.LevelPreset{},
		LevelTypes:       []model.LevelType{{}, {}, {}, {File1: "Act1/Caves/floor.dt1"}},
	}
	for mask := 1; mask <= 15; mask++ {
		record := model.LevelPreset{Def: 52 + mask, SizeX: 24, SizeY: 24, Files: 1, File1: "Act1/Caves/room.ds1", Dt1Mask: 1}
		result.LevelPresetByDef[record.Def] = record
	}
	for definition := 83; definition <= 90; definition++ {
		result.LevelPresetByDef[definition] = model.LevelPreset{Def: definition, SizeX: 24, SizeY: 24, Files: 1, File1: "Act1/Caves/special.ds1", Dt1Mask: 1}
	}
	return result
}
