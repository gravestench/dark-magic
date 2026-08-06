package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	lua "github.com/yuin/gopher-lua"
)

func TestLootModuleRollsLayeredTSVDeterministically(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"data/global/excel/TreasureClassEx.txt": &fstest.MapFile{Data: []byte("Treasure Class\tPicks\tNoDrop\tItem1\tProb1\nRoot\t1\t0\tr01\t1\n")},
		"test.lua":                              &fstest.MapFile{Data: []byte(`local loot=require("dm.loot/v1"); event_seed=assert(loot.event_seed(9,"monster",17,2)); drops=assert(loot.roll("Root", event_seed))`)},
	}
	runtime := New()
	if err := runtime.RegisterModule(LootModule(gamedata.New(recordstore.New(source)))); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), source, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("event_seed") == lua.LNil {
			t.Fatal("event seed was not exposed")
		}
		drops := state.GetGlobal("drops").(*lua.LTable)
		if drops.Len() != 1 || drops.RawGetInt(1).(*lua.LTable).RawGetString("code").String() != "r01" {
			t.Fatalf("drops = %s", drops)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
