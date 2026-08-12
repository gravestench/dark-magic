package d2legacy_test

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	. "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

func TestMissilePresentationResolvesCopiedRecipe(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `
local adapter=require("d2legacy.gameplay.missile_presentation")
local recipe=adapter.resolve({dcc="data/global/missiles/firebolt.dcc",palette="",velocity_x=1,velocity_y=0,directions=8,frames_per_second=0,loop=true})
assert(recipe.path=="data/global/missiles/firebolt.dcc")
assert(recipe.palette=="data/global/palette/units/pal.dat")
assert(recipe.direction==3 and recipe.frames_per_second==25 and recipe.loop=="loop")
assert(adapter.resolve({dcc="",velocity_x=0,velocity_y=0,directions=1})==nil)
`
	if err := runtime.RunScoped(context.Background(), &Scope{}, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
}
