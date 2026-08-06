package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	lua "github.com/yuin/gopher-lua"
)

func TestLootModuleRollsLayeredTSVDeterministically(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"treasure.txt": &fstest.MapFile{Data: []byte("Treasure Class\tPicks\tNoDrop\tItem1\tProb1\nRoot\t1\t0\tr01\t1\n")},
		"test.lua":     &fstest.MapFile{Data: []byte(`drops = assert(require("dm.loot/v1").roll_tsv("treasure.txt", "Root", 7))`)},
	}
	runtime := New()
	if err := runtime.RegisterModule(LootModule(source)); err != nil {
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
		drops := state.GetGlobal("drops").(*lua.LTable)
		if drops.Len() != 1 || drops.RawGetInt(1).(*lua.LTable).RawGetString("code").String() != "r01" {
			t.Fatalf("drops = %s", drops)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
