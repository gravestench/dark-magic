package modruntime

import (
	"context"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	lua "github.com/yuin/gopher-lua"
)

func TestPlayerCompositeResolvesCOFLayerWeaponClasses(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(Module{Name: "dm.render/v1", Loader: func(state *lua.LState) int {
		module := state.NewTable()
		state.SetField(module, "cof_info", state.NewFunction(func(state *lua.LState) int {
			if got := state.CheckString(1); got != "data/global/chars/AM/COF/AMWLHTH.cof" {
				state.RaiseError("unexpected COF %q", got)
			}
			info := state.NewTable()
			layers := state.NewTable()
			for _, layer := range []struct{ component, weaponClass string }{{"HD", "1ht"}, {"RA", "hth"}, {"RH", "hth"}} {
				entry := state.NewTable()
				entry.RawSetString("type", lua.LString(layer.component))
				entry.RawSetString("weapon_class", lua.LString(layer.weaponClass))
				layers.Append(entry)
			}
			info.RawSetString("layers", layers)
			state.Push(info)
			return 1
		}))
		state.Push(module)
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	script := `
local composite=require("darkmagic.gameplay.player_composite").unarmed({
  token="AM", mode="WL", weapon_class="HTH", palette="data/global/Palette/units/pal.dat", direction=3,
})
assert(composite.key=="AM:WL:HTH:3")
assert(composite.components.HD=="data/global/chars/AM/HD/AMHDLITWL1HT.dcc")
assert(composite.components.RA=="data/global/chars/AM/RA/AMRALITWLHTH.dcc")
assert(composite.components.RH==nil)
`
	scope := &Scope{}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRealArchivesComposeUnarmedPlayerModes(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	runtime := New()
	var composer render.Composer
	if err := runtime.RegisterInstaller(ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(NewRenderCapability(runtime, &composer, assets).Module()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	script := `
local render=require("dm.render/v1")
local adapter=require("darkmagic.gameplay.player_composite")
for _,token in ipairs({"AM","SO","NE","PA","BA","AI","DZ"}) do
  for _,mode in ipairs({"NU","WL","RN"}) do
    local ok,err=pcall(function()
      local composite=adapter.unarmed({token=token,mode=mode,weapon_class="HTH",palette="data/global/Palette/units/pal.dat",direction=3})
      local node=render.create("world")
      local frames=node:set_cof_animation(composite.cof,composite.palette,composite.direction,composite.components,"loop",composite.rate)
      assert(frames > 0)
      node:destroy()
    end)
    assert(ok,token..mode..": "..tostring(err))
  end
end
`
	scope := &Scope{}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
}
