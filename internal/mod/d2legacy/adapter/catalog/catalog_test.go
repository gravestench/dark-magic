package catalog

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/recovered"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

type testLocale map[string]string

func (locale testLocale) Text(key string) (string, error) {
	value, found := locale[key]
	if !found {
		return key, fmt.Errorf("missing %s", key)
	}
	return value, nil
}

func TestModulesExposeImmutableRecoveredRelationships(t *testing.T) {
	source := fstest.MapFS{
		recovered.QuestsPath:   &fstest.MapFile{Data: []byte("id\tname\tact\torder\tvisible\ticon\tquestdone\tqstr\tqsts1\n0\tPrologue\t0\t\t\t\t\tq0\ts0\n1\tDen\t0\t1\t1\ta1q1\t0\tq1\ts1\n")},
		recovered.SpeechPath:   &fstest.MapFile{Data: []byte("sound\tsoundstr\nakara_intro\tAkaraIntroGossip1\n")},
		recovered.DS1TypesPath: &fstest.MapFile{Data: []byte("Name\tDef\tLevelType\nTown\t1\t1\n")},
		recovered.ObjectsPath:  &fstest.MapFile{Data: []byte("Act\tId\tDescription\tObjectId\n1\t0\tFountain\t12\n")},
	}
	runtime := modruntime.New()
	catalog := recovered.New(source)
	if err := runtime.RegisterModule(QuestModule(catalog, testLocale{"AkaraIntroGossip1": "90\nWelcome, traveler.\nStay awhile."})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(MapModule(catalog)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`
local catalog=require("d2legacy.quest_catalog/v1")
local maps=require("d2legacy.map_catalog/v1")
local den=catalog.quest(1)
assert(den.name=="Den" and den.prerequisite_id==0 and den.stages[1].string_key=="s1")
assert(#catalog.quests(0)==2)
assert(catalog.speech("AKARA_INTRO").string_key=="AkaraIntroGossip1")
local dialog=catalog.dialog("akara_intro")
assert(dialog.sound=="akara_intro" and dialog.text=="Welcome, traveler.\nStay awhile.")
assert(dialog.scroll_lines_per_second==1.5)
assert(catalog.quest(99)==nil and catalog.speech("missing")==nil)
assert(maps.ds1_type(1).name=="Town")
assert(maps.object(1,0).object_id==12 and maps.object(2,0)==nil)
`)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestParseDialogueTextRejectsMalformedPayloads(t *testing.T) {
	if rate, body, err := ParseDialogueText("120\r\nLine one\r\nLine two"); err != nil || rate != 2 || body != "Line one\nLine two" {
		t.Fatalf("dialog = %v, %q, %v", rate, body, err)
	}
	for _, value := range []string{"missing header", "fast\nText", "-1\nText", "60\n"} {
		if _, _, err := ParseDialogueText(value); err == nil {
			t.Errorf("ParseDialogueText(%q) succeeded", value)
		}
	}
}
