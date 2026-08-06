package modruntime

import (
	"context"
	"reflect"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/navigation"
	"github.com/gravestench/dark-magic/internal/rendercore"
	lua "github.com/yuin/gopher-lua"
)

func TestLuaSceneNavigationAndScopedRendering(t *testing.T) {
	t.Parallel()

	runtime := New()
	manager := navigation.New()
	scenes := NewScenes(runtime, manager)
	var composer rendercore.Composer
	for _, module := range []Module{RenderModule(runtime, &composer), scenes.Module()} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"boot.lua": &fstest.MapFile{Data: []byte(`
local render = require("dm.render/v1")
local scenes = require("dm.scene/v1")
return {
  id = "boot",
  start = function(self)
    scenes.register("world", {
	  create = function(self) calls = (calls or "") .. "create;" end,
	  enter = function(self) calls = calls .. "enter;"; self.root = render.create("world") end,
      update = function(self, dt) calls = calls .. "update;" end,
      render = function(self) calls = calls .. "render;" end,
      exit = function(self) calls = calls .. "exit;" end,
      destroy = function(self) calls = calls .. "destroy;" end,
    })
    scenes.replace("world")
  end,
}`)}}
	definition, err := LoadDefinition(context.Background(), runtime, source, "boot.lua")
	if err != nil {
		t.Fatal(err)
	}
	components := host.NewManager()
	if err := components.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := components.Enable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Update(context.Background(), time.Second/60); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Render(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"world"}) {
		t.Fatalf("stack = %v", got)
	}
	if len(composer.Snapshot()) != 1 {
		t.Fatalf("render nodes = %#v", composer.Snapshot())
	}
	if err := manager.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(composer.Snapshot()) != 0 {
		t.Fatalf("render nodes leaked: %#v", composer.Snapshot())
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if got := state.GetGlobal("calls").String(); got != "create;enter;update;render;exit;destroy;" {
			t.Fatalf("calls = %q", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
