package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	lua "github.com/yuin/gopher-lua"
)

func TestLuaDefinitionUsesSharedManagerLifecycle(t *testing.T) {
	t.Parallel()

	runtime := New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"system.lua": &fstest.MapFile{Data: []byte(`
return {
    id = "game.inventory",
    depends_on = { "engine.render" },
    start = function(self) calls = (calls or "") .. "start " .. self.id .. ";" end,
    stop = function(self) calls = calls .. "stop " .. self.id .. ";" end,
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
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if got := state.GetGlobal("calls").String(); got != "start game.inventory;stop game.inventory;" {
			t.Fatalf("calls = %q", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestVFSVersionedCapability(t *testing.T) {
	t.Parallel()

	contentFS, err := content.New(content.Layer{Name: "shim", FS: fstest.MapFS{"value.txt": &fstest.MapFile{Data: []byte("hello")}}})
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
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("value").String() != "hello" || state.GetGlobal("origin").String() != "shim" {
			t.Fatalf("value/source = %s/%s", state.GetGlobal("value"), state.GetGlobal("origin"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReloadDefinitionTransfersSerializableState(t *testing.T) {
	t.Parallel()

	runtime := New()
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
return {
    id = "game.counter",
    import_state = function(self, state) self.count = state.count; self.second = state.names[2] end,
    start = function(self) imported_count = self.count; imported_second = self.second end,
}`)}}
	if err := ReloadDefinition(context.Background(), manager, runtime, newSource, "system.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("imported_count").String() != "7" || state.GetGlobal("imported_second").String() != "b" {
			t.Fatalf("imported state = %s/%s", state.GetGlobal("imported_count"), state.GetGlobal("imported_second"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
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
