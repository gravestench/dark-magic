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

// Invalidate owns the invalidate step at this boundary, keeping its side effects and failure point explicit to
// callers.
func (g *trackingRecordsGateway) Invalidate(path string) {
	g.invalidated = append(g.invalidated, path)
	g.Store.Invalidate(path)
}

// TestRecordsModuleLoadsAndInvalidatesLayeredTSV verifies typed TSV loading and explicit cache invalidation through
// the Lua records API.
func TestRecordsModuleLoadsAndInvalidatesLayeredTSV(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"items.txt": &fstest.MapFile{Data: []byte("code\tname\na\tAlpha\n")},
		"test.lua": &fstest.MapFile{
			Data: []byte(
				`local r=require("engine.records/v1"); rows=assert(r.load("items.txt")); ` +
					`was_loaded=r.loaded("items.txt"); r.reload("items.txt"); is_loaded=r.loaded("items.txt")`,
			),
		},
	}
	runtime := New()

	records := &trackingRecordsGateway{Store: recordstore.New(source)}
	if err := runtime.RegisterModule(RecordsModule(records)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	if err := runtime.Execute(context.Background(), source, "test.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		rows := state.GetGlobal("rows").(*lua.LTable)
		if rows.RawGetInt(1).(*lua.LTable).RawGetString("name").String() != "Alpha" ||
			state.GetGlobal("was_loaded") != lua.LTrue ||
			state.GetGlobal("is_loaded") != lua.LFalse {
			t.Fatalf(
				"records state = %s/%s/%s",
				rows,
				state.GetGlobal("was_loaded"),
				state.GetGlobal("is_loaded"),
			)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if len(records.invalidated) != 1 || records.invalidated[0] != "items.txt" {
		t.Fatalf("invalidations = %v", records.invalidated)
	}
}

func TestExperimentRecordProjectionPreservesOwnershipAndReload(t *testing.T) {
	source := fstest.MapFS{"items.txt": &fstest.MapFile{Data: []byte("code\tname\tblank\na\tAlpha\t\nb\tBeta\t\n")}}
	records := recordstore.New(source)
	runtime := New()
	if err := runtime.RegisterModule(RecordsModule(records)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	run := func(script string) {
		t.Helper()
		if err := runtime.Run(t.Context(), func(L *lua.LState) error { return L.DoString(script) }); err != nil {
			t.Fatal(err)
		}
	}
	run(`r=require("engine.records/v1"); for _,columns in ipairs({{[0]="code"},{[-1]="code"},{[1.5]="code"},{[1]="code",[3]="name"}}) do assert(not pcall(r.load,"items.txt",columns)) end; assert(next(r.load("items.txt",{})[1])==nil); assert(not pcall(r.load,"items.txt",true)); assert(not pcall(r.load,"items.txt",{123})); assert(not pcall(r.load,"items.txt",{named="code"})); first=r.load("items.txt",{"code","blank","unknown"}); assert(#first==2 and first[1].code=="a" and first[2].code=="b" and first[1].blank=="" and first[1].unknown==nil and first[1].name==nil); first[1].code="mutated"; assert(r.load("items.txt",{"code"})[1].code=="a"); assert(r.load("items.txt")[1].name=="Alpha")`)
	source["items.txt"].Data = []byte("code\tname\tblank\nc\tGamma\t\n")
	run(`assert(r.load("items.txt",{"code"})[1].code=="a"); r.reload("items.txt"); assert(r.load("items.txt",{"code","blank"})[1].code=="c" and r.load("items.txt",{"blank"})[1].blank=="")`)
}
