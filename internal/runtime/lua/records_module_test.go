package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/store"
	lua "github.com/yuin/gopher-lua"
)

type trackingRecordsGateway struct {
	*recordstore.Store
	invalidated []string
}

func (g *trackingRecordsGateway) Invalidate(path string) {
	g.invalidated = append(g.invalidated, path)
	g.Store.Invalidate(path)
}

func TestRecordsModuleLoadsAndInvalidatesLayeredTSV(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"items.txt": &fstest.MapFile{Data: []byte("code\tname\na\tAlpha\n")},
		"test.lua":  &fstest.MapFile{Data: []byte(`local r=require("dm.records/v1"); rows=assert(r.load("items.txt")); was_loaded=r.loaded("items.txt"); r.reload("items.txt"); is_loaded=r.loaded("items.txt")`)},
	}
	runtime := New()
	records := &trackingRecordsGateway{Store: recordstore.New(source)}
	if err := runtime.RegisterModule(RecordsModule(records)); err != nil {
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
		rows := state.GetGlobal("rows").(*lua.LTable)
		if rows.RawGetInt(1).(*lua.LTable).RawGetString("name").String() != "Alpha" || state.GetGlobal("was_loaded") != lua.LTrue || state.GetGlobal("is_loaded") != lua.LFalse {
			t.Fatalf("records state = %s/%s/%s", rows, state.GetGlobal("was_loaded"), state.GetGlobal("is_loaded"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(records.invalidated) != 1 || records.invalidated[0] != "items.txt" {
		t.Fatalf("invalidations = %v", records.invalidated)
	}
}
