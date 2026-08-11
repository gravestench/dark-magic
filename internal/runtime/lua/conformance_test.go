package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/persistence"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/presentation/scene"
	"github.com/gravestench/dark-magic/internal/video"
	lua "github.com/yuin/gopher-lua"
)

func TestVersionedCapabilityConformance(t *testing.T) {
	source := fstest.MapFS{}
	contentFS, err := content.New(content.Layer{Name: "test", FS: source})
	if err != nil {
		t.Fatal(err)
	}
	runtime := New()
	var input inputstate.Store
	var mixer audio.Mixer
	var composer render.Composer
	itemState, err := gameitem.NewState(gameitem.Layout{Grids: map[gameitem.Container]gameitem.Grid{gameitem.ContainerInventory: {Width: 10, Height: 4}}}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	itemAuthority := gameitem.NewAuthority()
	if err := itemAuthority.Register("local-player", itemState); err != nil {
		t.Fatal(err)
	}
	scenes := NewScenes(runtime, navigation.New())
	modules := []Module{
		AppModule("test", func() {}),
		VFSModule(contentFS), DataModule(contentFS), WorldModule(contentFS), InputModule(&input), AudioModule(runtime, &mixer, source, gamedata.New(recordstore.New(source))),
		SettingsModule(preferences.NewTransient(), &mixer),
		VideoModule(runtime, video.Unavailable{}, source),
		RecordsModule(recordstore.New(source)), GameDataModule(staticGameData{snapshot: gamedata.Snapshot{}}), LocaleModule(localization.New(source, "English")),
		LootModule(gamedata.New(recordstore.New(source))), SaveModule(persistence.New()), PlayerControlModule(&gamesession.MovementController{}), ItemModule(itemAuthority, &gameitem.Controller{}, "local-player"), SimulationModule(NewSimulation(scene.New(1, 10, 10))),
		RenderModule(runtime, &composer), scenes.Module(),
	}
	expected := map[string][]string{
		"dm.app/v1": {"request_exit", "version"},
		"dm.vfs/v1": {"list", "read", "read_text", "source"}, "dm.input/v1": {"down", "pressed", "released", "cursor", "text"},
		"dm.data/v1":  {"load", "load_manifest"},
		"dm.world/v1": {"load"},
		"dm.audio/v1": {"diagnostics", "exists", "play", "play_persistent", "play_record", "set_bus_volume", "stop_group"}, "dm.records/v1": {"load", "reload", "loaded"},
		"dm.settings/v1":  {"get", "save", "set", "status"},
		"dm.video/v1":     {"available", "play"},
		"dm.game_data/v1": {"character_class", "skill", "unique_titles"},
		"dm.locale/v1":    {"text"}, "dm.loot/v1": {"event_seed", "roll"},
		"dm.save/v1":       {"characters", "create", "create_named", "delete", "select", "selected"},
		"dm.player/v1":     {"assign_skill", "request_running"},
		"dm.items/v1":      {"move", "snapshot"},
		"dm.simulation/v1": {"move_hero", "state"}, "dm.render/v1": {"create", "diagnostics"},
		"dm.scene/v1": {"register", "replace", "push", "pop", "toggle_overlay"},
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
	documentation := runtime.ModuleHelp()
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
			module.ForEach(func(key, value lua.LValue) {
				if value.Type() != lua.LTFunction {
					return
				}
				doc, ok := documentation[name].Commands[key.String()]
				if !ok || doc.Summary == "" || doc.Usage == "" {
					t.Errorf("%s.%s lacks authored help metadata", name, key)
				}
			})
			for typeName, typeDoc := range documentation[name].Types {
				metatable, ok := state.GetTypeMetatable(typeName).(*lua.LTable)
				if !ok {
					t.Errorf("%s does not register documented userdata %s", name, typeName)
					continue
				}
				methods, _ := metatable.RawGetString("__index").(*lua.LTable)
				if methods == nil {
					t.Errorf("%s userdata %s has no method table", name, typeName)
					continue
				}
				methods.ForEach(func(key, value lua.LValue) {
					if value.Type() != lua.LTFunction {
						return
					}
					doc, ok := typeDoc.Methods[key.String()]
					if !ok || doc.Summary == "" || doc.Usage == "" {
						t.Errorf("%s %s.%s lacks authored help metadata", name, typeName, key)
					}
				})
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
