package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/localecore"
	"github.com/gravestench/dark-magic/internal/navigation"
	"github.com/gravestench/dark-magic/internal/recordstore"
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/internal/savecore"
	"github.com/gravestench/dark-magic/internal/videocore"
	"github.com/gravestench/dark-magic/pkg/scene"
	lua "github.com/yuin/gopher-lua"
)

func TestVersionedCapabilityConformance(t *testing.T) {
	source := fstest.MapFS{}
	contentFS, err := content.New(content.Layer{Name: "test", FS: source})
	if err != nil {
		t.Fatal(err)
	}
	runtime := New()
	var input inputcore.Store
	var mixer audiocore.Mixer
	var composer rendercore.Composer
	scenes := NewScenes(runtime, navigation.New())
	modules := []Module{
		AppModule("test", func() {}),
		VFSModule(contentFS), DataModule(contentFS), InputModule(&input), AudioModule(runtime, &mixer, source),
		VideoModule(runtime, videocore.Unavailable{}, source),
		RecordsModule(recordstore.New(source)), LocaleModule(localecore.New(source, "English")),
		LootModule(source), SaveModule(savecore.New()), SimulationModule(NewSimulation(scene.New(1, 10, 10))),
		RenderModule(runtime, &composer), scenes.Module(),
	}
	expected := map[string][]string{
		"dm.app/v1": {"request_exit", "version"},
		"dm.vfs/v1": {"read", "read_text", "source"}, "dm.input/v1": {"down", "pressed", "released", "cursor", "text"},
		"dm.data/v1":  {"load", "load_manifest"},
		"dm.audio/v1": {"diagnostics", "exists", "play", "play_persistent", "play_record", "set_bus_volume", "stop_group"}, "dm.records/v1": {"load", "reload", "loaded"},
		"dm.video/v1":  {"available", "play"},
		"dm.locale/v1": {"text"}, "dm.loot/v1": {"event_seed"},
		"dm.save/v1":       {"characters", "create", "create_named", "delete", "select", "selected"},
		"dm.simulation/v1": {"move_hero", "state"}, "dm.render/v1": {"create", "diagnostics"},
		"dm.scene/v1": {"register", "replace", "push", "pop"},
	}
	for _, module := range modules {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		for name, functions := range expected {
			if err := state.CallByParam(lua.P{Fn: state.GetGlobal("require"), NRet: 1, Protect: true}, lua.LString(name)); err != nil {
				t.Fatalf("require %s: %v", name, err)
			}
			module, ok := state.Get(-1).(*lua.LTable)
			state.Pop(1)
			if !ok || module.RawGetString("api") != lua.LNumber(1) {
				t.Fatalf("%s does not declare api=1", name)
			}
			for _, function := range functions {
				if module.RawGetString(function).Type() != lua.LTFunction {
					t.Errorf("%s.%s is not a function", name, function)
				}
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
