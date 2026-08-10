package modruntime

import (
	"context"
	"reflect"
	"runtime/pprof"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	lua "github.com/yuin/gopher-lua"
)

func TestSceneFrameContextLabelsFocusedScene(t *testing.T) {
	manager := navigation.New()
	scenes := NewScenes(New(), manager)
	ctx := scenes.FrameContext(context.Background())
	if got, ok := pprof.Label(ctx, "scene"); !ok || got != "none" {
		t.Fatalf("empty scene label = %q, %v; want none, true", got, ok)
	}
	if err := manager.Register("title", func(context.Context) (navigation.Scene, error) {
		return &frameLabelScene{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := manager.Replace(context.Background(), "title"); err != nil {
		t.Fatal(err)
	}
	ctx = scenes.FrameContext(context.Background())
	if got, ok := pprof.Label(ctx, "scene"); !ok || got != "title" {
		t.Fatalf("focused scene label = %q, %v; want title, true", got, ok)
	}
}

type frameLabelScene struct{}

func (*frameLabelScene) Create(context.Context) error                { return nil }
func (*frameLabelScene) Enter(context.Context) error                 { return nil }
func (*frameLabelScene) Update(context.Context, time.Duration) error { return nil }
func (*frameLabelScene) Render(context.Context) error                { return nil }
func (*frameLabelScene) Exit(context.Context) error                  { return nil }
func (*frameLabelScene) Destroy(context.Context) error               { return nil }

func TestLuaSceneNavigationAndScopedRendering(t *testing.T) {
	t.Parallel()

	runtime := New()
	calls := ""
	if err := runtime.RegisterModule(Module{Name: "test.sink/v1", Loader: func(state *lua.LState) int {
		state.Push(state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{"add": func(state *lua.LState) int { calls += state.CheckString(1); return 0 }}))
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	manager := navigation.New()
	scenes := NewScenes(runtime, manager)
	profiler := &recordingSceneProfiler{}
	scenes.SetProfiler(profiler)
	var composer render.Composer
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
local sink = require("test.sink/v1")
return {
  id = "boot",
  start = function(self)
    scenes.register("world", {
	  create = function(self) sink.add("create;") end,
	  enter = function(self) sink.add("enter;"); self.root = render.create("world") end,
      update = function(self, dt) sink.add("update;") end,
      render = function(self) sink.add("render;") end,
      exit = function(self) sink.add("exit;") end,
      destroy = function(self) sink.add("destroy;") end,
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
	if calls != "create;enter;update;render;exit;destroy;" {
		t.Fatalf("calls = %q", calls)
	}
	if !reflect.DeepEqual(profiler.scenes, []string{"world"}) {
		t.Fatalf("profiled scenes = %v", profiler.scenes)
	}
}

func TestNonfocusedLuaSceneCannotReadFocusedOverlayInput(t *testing.T) {
	ctx := context.Background()
	runtime := New()
	observed := make(map[string]bool)
	if err := runtime.RegisterModule(Module{Name: "test.focus/v1", Loader: func(state *lua.LState) int {
		state.Push(state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{"record": func(state *lua.LState) int {
			observed[state.CheckString(1)] = state.CheckBool(2)
			return 0
		}}))
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	manager := navigation.New()
	scenes := NewScenes(runtime, manager)
	var input inputstate.Store
	scenes.SetInputStore(&input)
	for _, module := range []Module{InputModule(&input), scenes.Module()} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	source := fstest.MapFS{"boot.lua": &fstest.MapFile{Data: []byte(`
local input=require("dm.input/v1")
local scenes=require("dm.scene/v1")
local focus=require("test.focus/v1")
return {id="boot",api=1,start=function(self)
  scenes.register("world",{update=function(self,elapsed,focused)
    focus.record("world",input.pressed("confirm"))
  end})
  scenes.register("overlay",{blocks_update_below=false,update=function(self,elapsed,focused)
    focus.record("overlay",input.pressed("confirm"))
  end})
  scenes.replace("world")
  scenes.push("overlay")
end}
`)}}
	definition, err := LoadDefinition(ctx, runtime, source, "boot.lua")
	if err != nil {
		t.Fatal(err)
	}
	components := host.NewManager()
	if err := components.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := components.Enable(ctx, definition.ID); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"confirm": {Pressed: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "overlay"}})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	if observed["world"] || !observed["overlay"] {
		t.Fatalf("input visibility = %#v", observed)
	}
}

type recordingSceneProfiler struct{ scenes []string }

func (p *recordingSceneProfiler) CaptureSceneHeap(scene string) error {
	p.scenes = append(p.scenes, scene)
	return nil
}

func TestLuaSceneReplacementDestroysPreviousComposition(t *testing.T) {
	runtime := New()
	manager := navigation.New()
	scenes := NewScenes(runtime, manager)
	var composer render.Composer
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
local render=require("dm.render/v1"); local scenes=require("dm.scene/v1")
local function screen(layer) return {create=function(self)
  self.root=render.create(layer); self.child=render.create(layer,self.root)
end} end
return {id="boot",api=1,start=function(self)
  scenes.register("one",screen("hud")); scenes.register("two",screen("hud")); scenes.replace("one")
end}`)}}
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
	if got := composer.Diagnostics().ActiveNodes; got != 2 {
		t.Fatalf("first scene nodes = %d", got)
	}
	err = runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`require("dm.scene/v1").replace("two")`)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := scenes.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := composer.Diagnostics().ActiveNodes; got != 2 {
		t.Fatalf("replacement retained both scenes: nodes = %d", got)
	}
}

func TestLuaSceneReentryDoesNotShareMutableDefinitionState(t *testing.T) {
	runtime := New()
	manager := navigation.New()
	scenes := NewScenes(runtime, manager)
	var composer render.Composer
	for _, module := range []Module{RenderModule(runtime, &composer), scenes.Module()} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"boot.lua": {Data: []byte(`
local render=require("dm.render/v1"); local scenes=require("dm.scene/v1")
local screen={
  create=function(self) self.root=render.create("hud") end,
  update=function(self) self.root:set_position(10,20) end,
  destroy=function(self) self.root:destroy() end,
}
return {id="boot",api=1,start=function(self)
  scenes.register("screen",screen); scenes.replace("screen")
end}`)}}
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
	if err := manager.Replace(context.Background(), "screen"); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Update(context.Background(), time.Second/60); err != nil {
		t.Fatalf("re-entered scene retained a stale render handle: %v", err)
	}
	if got := composer.Diagnostics().ActiveNodes; got != 1 {
		t.Fatalf("active nodes = %d, want 1", got)
	}
}
