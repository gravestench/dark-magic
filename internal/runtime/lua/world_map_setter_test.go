package modruntime

import (
	"context"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	lua "github.com/yuin/gopher-lua"
)

func TestSetWorldMapRegistersMethodsWithoutWorldModule(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(Module{Name: "test.collision/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"set": func(state *lua.LState) int {
				collision := state.Get(1)
				state.SetGlobal("test_collision", collision)
				return 0
			},
		})
		state.Push(module)
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })

	world := &gameworld.Map{WidthSubtiles: 2, HeightSubtiles: 2}
	if err := SetWorldMap(context.Background(), runtime, "test.collision/v1", "set", world); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`assert(test_collision:blocked_position(0, 0) == false)`)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetWorldMapForLevelPassesLevelBeforeCollision(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(Module{Name: "test.level_collision/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"set": func(state *lua.LState) int {
				state.SetGlobal("test_level", state.Get(1))
				state.SetGlobal("test_collision", state.Get(2))
				return 0
			},
		})
		state.Push(module)
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })
	if err := SetWorldMapForLevel(context.Background(), runtime, "test.level_collision/v1", "set", 7,
		&gameworld.Map{WidthSubtiles: 2, HeightSubtiles: 2}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`assert(test_level == 7 and test_collision:blocked_position(0, 0) == false)`)
	}); err != nil {
		t.Fatal(err)
	}
}
