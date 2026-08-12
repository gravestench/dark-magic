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
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(Module{Name: "engine.render/v1", Loader: func(state *lua.LState) int {
		module := state.NewTable()
		state.SetField(module, "cof_info", state.NewFunction(func(state *lua.LState) int {
			got := state.CheckString(1)
			if got != "data/global/chars/AM/COF/AMWLHTH.cof" && got != "data/global/chars/AM/COF/AMWL1HS.cof" {
				state.RaiseError("unexpected COF %q", got)
			}
			info := state.NewTable()
			info.RawSetString("directions", lua.LNumber(16))
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
		state.SetField(module, "asset_exists", state.NewFunction(func(state *lua.LState) int {
			path := state.CheckString(1)
			state.Push(lua.LBool(path == "data/global/chars/AM/HD/AMHDLITWL1HT.dcc" || path == "data/global/chars/AM/RA/AMRALITWLHTH.dcc"))
			return 1
		}))
		state.SetField(module, "animdata_info", state.NewFunction(func(state *lua.LState) int {
			got := state.CheckString(1)
			if got != "AMWLHTH" && got != "AMWL1HS" {
				state.RaiseError("unexpected animation key %q", got)
			}
			info := state.NewTable()
			info.RawSetString("speed", lua.LNumber(333))
			info.RawSetString("frames", lua.LNumber(8))
			info.RawSetString("events", state.NewTable())
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
local composite=require("d2legacy.gameplay.player_composite").unarmed({
  token="AM", mode="WL", weapon_class="HTH", palette="data/global/Palette/units/pal.dat", direction=3,
})
assert(string.sub(composite.key,1,12)=="AM:WL:HTH:14")
assert(composite.direction==14 and composite.dcc_direction==3)
assert(composite.components.HD=="data/global/chars/AM/HD/AMHDLITWL1HT.dcc")
assert(composite.components.RA=="data/global/chars/AM/RA/AMRALITWLHTH.dcc")
assert(composite.components.RH==nil)
assert(composite.rate==333 and composite.frames==8)
local fine_direction=require("d2legacy.gameplay.player_composite").unarmed({
  token="AM", mode="WL", weapon_class="HTH", palette="data/global/Palette/units/pal.dat", direction=15,
})
assert(fine_direction.direction==15)
local equipped=require("d2legacy.gameplay.player_composite").resolve({
  token="AM", mode="WL", weapon_class="HTH", palette="data/global/Palette/units/pal.dat", direction=3,
},{active_weapon_set=0,items={{container="equipment",slot="rarm",weapon_set=0,weapon_class="1hs",composite={RH="ssd"}}}})
assert(equipped.cof=="data/global/chars/AM/COF/AMWL1HS.cof")
assert(equipped.components.RH=="data/global/chars/AM/RH/AMRHSSDWLHTH.dcc")
local playback=require("d2legacy.gameplay.player_composite").new_playback({mode="A"})
local crossed=require("d2legacy.gameplay.player_composite").advance(playback,{rate=256,frames=4,events={[2]=1,[3]=3},mode="A"},0.09)
assert(playback.frame==3 and #crossed==2 and crossed[1].event==1 and crossed[2].event==3)
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
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	capability := NewRenderCapability(runtime, &composer, assets)
	if err := runtime.RegisterModule(capability.Module()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	script := `
local render=require("engine.render/v1")
local adapter=require("d2legacy.gameplay.player_composite")
for _,token in ipairs({"AM","SO","NE","PA","BA","AI","DZ"}) do
  for _,mode in ipairs({"NU","WL","RN"}) do
    local ok,err=pcall(function()
      local composite=adapter.unarmed({token=token,mode=mode,weapon_class="HTH",palette="data/global/Palette/units/pal.dat",direction=3})
	  assert(composite.direction==14 and composite.dcc_direction==3)
	  if token=="NE" then
	    assert(composite.components.S1 and composite.components.S2)
	  end
      local node=render.create("world")
      local frames=node:set_cof_animation(composite.cof,composite.palette,composite.direction,composite.components,"loop",composite.rate)
      assert(frames > 0)
      node:destroy()
    end)
    assert(ok,token..mode..": "..tostring(err))
  end
end
for _,mode in ipairs({"NU","WL","RN"}) do
  local composite=adapter.resolve({token="AM",mode=mode,weapon_class="HTH",palette="data/global/Palette/units/pal.dat",direction=3},{
    active_weapon_set=0,
    items={{container="equipment",slot="rarm",weapon_set=0,weapon_class="1HS",composite={RH="SSD"}}},
  })
  local node=render.create("world")
  assert(node:set_cof_animation(composite.cof,composite.palette,composite.direction,composite.components,"loop",composite.rate) > 0)
  node:destroy()
end
`
	scope := &Scope{}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
	before := capability.Diagnostics().DecodeCalls
	secondScope := &Scope{}
	if err := runtime.RunScoped(context.Background(), secondScope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
	if err := secondScope.Close(); err != nil {
		t.Fatal(err)
	}
	if after := capability.Diagnostics().DecodeCalls; after != before {
		t.Fatalf("warm composite recipes decoded again: before=%d after=%d", before, after)
	}
}
