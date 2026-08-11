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

func TestMapgenModuleExposesPresetValueSnapshot(t *testing.T) {
	runtime := New()
	stub := mapgenCatalogStub{snapshot: gamedata.Snapshot{
		LevelsByID:   map[int]model.LevelData{1: {Id: 1, DrlgType: 2, LevelType: 1, SizeX: 8, SizeY: 8}},
		LevelPresets: []model.LevelPreset{{Def: 1, LevelId: 1, Files: 1, File1: "Act1/Town/town.ds1", Dt1Mask: 1}},
		LevelTypes:   []model.LevelType{{}, {File1: "Act1/Town/floor.dt1"}},
	}}
	if err := runtime.RegisterModule(MapgenModule(stub)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`local m=require("dm.mapgen/v1"); local z=m.preset(1,42); assert(z.kind..":"..z.stamps[1].ds1 == "preset:data/global/tiles/Act1/Town/town.ds1")`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
