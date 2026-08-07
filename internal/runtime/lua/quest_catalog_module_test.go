package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/recovered"
	lua "github.com/yuin/gopher-lua"
)

func TestQuestCatalogModuleExposesHierarchyAndSpeech(t *testing.T) {
	source := fstest.MapFS{
		recovered.QuestsPath:   &fstest.MapFile{Data: []byte("id\tname\tact\torder\tvisible\ticon\tquestdone\tqstr\tqsts1\n0\tPrologue\t0\t\t\t\t\tq0\ts0\n1\tDen\t0\t1\t1\ta1q1\t0\tq1\ts1\n")},
		recovered.SpeechPath:   &fstest.MapFile{Data: []byte("sound\tsoundstr\nakara_intro\tAkaraIntroGossip1\n")},
		recovered.DS1TypesPath: &fstest.MapFile{Data: []byte("Name\tDef\tLevelType\nTown\t1\t1\n")},
		recovered.ObjectsPath:  &fstest.MapFile{Data: []byte("Act\tId\tDescription\tObjectId\n1\t0\tFountain\t12\n")},
	}
	runtime := New()
	catalog := recovered.New(source)
	if err := runtime.RegisterModule(QuestCatalogModule(catalog)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(MapCatalogModule(catalog)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`
local catalog=require("dm.quest_catalog/v1")
local maps=require("dm.map_catalog/v1")
local den=catalog.quest(1)
assert(den.name=="Den" and den.prerequisite_id==0 and den.stages[1].string_key=="s1")
assert(#catalog.quests(0)==2)
assert(catalog.speech("AKARA_INTRO").string_key=="AkaraIntroGossip1")
assert(catalog.quest(99)==nil and catalog.speech("missing")==nil)
assert(maps.ds1_type(1).name=="Town")
assert(maps.object(1,0).object_id==12 and maps.object(2,0)==nil)
`)
	}); err != nil {
		t.Fatal(err)
	}
}
