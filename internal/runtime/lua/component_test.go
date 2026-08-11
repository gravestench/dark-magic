package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/content"
	lua "github.com/yuin/gopher-lua"
)

func TestLuaDefinitionUsesSharedManagerLifecycle(t *testing.T) {
	t.Parallel()

	runtime := New()
	calls := ""
	if err := runtime.RegisterModule(Module{Name: "test.sink/v1", Loader: func(state *lua.LState) int {
		state.Push(state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{"add": func(state *lua.LState) int { calls += state.CheckString(1); return 0 }}))
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"system.lua": &fstest.MapFile{Data: []byte(`
local sink = require("test.sink/v1")
return {
    id = "game.inventory",
    depends_on = { "engine.render" },
    start = function(self) sink.add("start " .. self.id .. ";") end,
    stop = function(self) sink.add("stop " .. self.id .. ";") end,
}`)}}
	definition, err := LoadDefinition(context.Background(), runtime, source, "system.lua")
	if err != nil {
		t.Fatal(err)
	}
	manager := host.NewManager()
	if err := manager.Register(host.ManagedDefinition{ID: "engine.render", New: func(context.Context) (host.Component, error) {
		return componentFunc{}, nil
	}}); err != nil {
		t.Fatal(err)
	}
	if err := manager.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(context.Background(), "game.inventory"); err != nil {
		t.Fatal(err)
	}
	if err := manager.DisableCascade(context.Background(), "engine.render"); err != nil {
		t.Fatal(err)
	}
	if calls != "start game.inventory;stop game.inventory;" {
		t.Fatalf("calls = %q", calls)
	}
}

func TestVFSVersionedCapability(t *testing.T) {
	t.Parallel()

	contentFS, err := content.New(content.Layer{Name: "shim", FS: fstest.MapFS{"value.txt": &fstest.MapFile{Data: []byte("hello")}, "tiles/one.DT1": &fstest.MapFile{Data: []byte("tile")}}})
	if err != nil {
		t.Fatal(err)
	}
	runtime := New()
	if err := runtime.RegisterModule(VFSModule(contentFS)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local vfs = require("dm.vfs/v1")
value = assert(vfs.read("value.txt"))
origin = assert(vfs.source("value.txt")).layer
listed = assert(vfs.list(".", ".dt1"))
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("value").String() != "hello" || state.GetGlobal("origin").String() != "shim" {
			t.Fatalf("value/source = %s/%s", state.GetGlobal("value"), state.GetGlobal("origin"))
		}
		listed := state.GetGlobal("listed").(*lua.LTable)
		if listed.Len() != 1 || listed.RawGetInt(1).String() != "tiles/one.DT1" {
			t.Fatalf("listed assets = %s", listed)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReloadDefinitionTransfersSerializableState(t *testing.T) {
	t.Parallel()

	runtime := New()
	var importedCount, importedSecond string
	if err := runtime.RegisterModule(Module{Name: "test.sink/v1", Loader: func(state *lua.LState) int {
		state.Push(state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{"imported": func(state *lua.LState) int {
			importedCount, importedSecond = state.CheckString(1), state.CheckString(2)
			return 0
		}}))
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	oldSource := fstest.MapFS{"system.lua": &fstest.MapFile{Data: []byte(`
return {
    id = "game.counter",
    start = function(self) self.count = self.count or 7 end,
    export_state = function(self) return { count = self.count, names = { "a", "b" } } end,
}`)}}
	definition, err := LoadDefinition(context.Background(), runtime, oldSource, "system.lua")
	if err != nil {
		t.Fatal(err)
	}
	manager := host.NewManager()
	if err := manager.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	newSource := fstest.MapFS{"system.lua": &fstest.MapFile{Data: []byte(`
local sink = require("test.sink/v1")
return {
    id = "game.counter",
    import_state = function(self, state) self.count = state.count; self.second = state.names[2] end,
    start = function(self) sink.imported(tostring(self.count), self.second) end,
}`)}}
	if err := ReloadDefinition(context.Background(), manager, runtime, newSource, "system.lua"); err != nil {
		t.Fatal(err)
	}
	if importedCount != "7" || importedSecond != "b" {
		t.Fatalf("imported state = %s/%s", importedCount, importedSecond)
	}
}

func TestReloadDefinitionKeepsWorkingInstanceOnCompileFailure(t *testing.T) {
	t.Parallel()

	runtime := New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	valid := fstest.MapFS{"system.lua": &fstest.MapFile{Data: []byte(`return { id = "game.valid" }`)}}
	definition, err := LoadDefinition(context.Background(), runtime, valid, "system.lua")
	if err != nil {
		t.Fatal(err)
	}
	manager := host.NewManager()
	if err := manager.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	broken := fstest.MapFS{"system.lua": &fstest.MapFile{Data: []byte(`return { id = `)}}
	if err := ReloadDefinition(context.Background(), manager, runtime, broken, "system.lua"); err == nil {
		t.Fatal("expected compile failure")
	}
	status, _ := manager.Status(definition.ID)
	if status.State != host.StateEnabled {
		t.Fatalf("status after failed reload = %#v", status)
	}
}

type componentFunc struct{}

func (componentFunc) Start(context.Context) error { return nil }
func (componentFunc) Stop(context.Context) error  { return nil }
