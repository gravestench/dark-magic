package modruntime

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	lua "github.com/yuin/gopher-lua"
)

func TestMonsterCompositeUsesJoinedMonStats2Pieces(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(Module{Name: "engine.render/v1", Loader: func(state *lua.LState) int {
		module := state.NewTable()
		state.SetField(module, "cof_info", state.NewFunction(func(state *lua.LState) int {
			if got := state.CheckString(1); got != "data/global/monsters/FA/COF/FAWLHTH.cof" {
				state.RaiseError("unexpected COF %q", got)
			}
			info, layers := state.NewTable(), state.NewTable()
			info.RawSetString("directions", lua.LNumber(8))
			info.RawSetString("frames", lua.LNumber(8))
			info.RawSetString("events", state.NewTable())
			for _, value := range []string{"HD", "TR", "SH"} {
				layer := state.NewTable()
				layer.RawSetString("type", lua.LString(value))
				layer.RawSetString("weapon_class", lua.LString("hth"))
				layers.Append(layer)
			}
			info.RawSetString("layers", layers)
			state.Push(info)
			return 1
		}))
		state.SetField(module, "asset_exists", state.NewFunction(func(state *lua.LState) int {
			path := state.CheckString(1)
			state.Push(lua.LBool(path != "data/global/monsters/FA/SH/FASHLITWLHTH.dcc"))
			return 1
		}))
		state.SetField(module, "animdata_info", state.NewFunction(func(state *lua.LState) int { state.Push(lua.LNil); return 1 }))
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
local adapter=require("d2.gameplay.monster_composite")
assert(adapter.facing(6,0,0)==6)
assert(adapter.facing(0,1,0)==3)
local composite=adapter.resolve({token="FA",mode="WL",weapon_class="HTH",components="HD=LIT,TR=MED",direction=3})
assert(composite.direction==7)
assert(composite.components.HD=="data/global/monsters/FA/HD/FAHDLITWLHTH.dcc")
assert(composite.components.TR=="data/global/monsters/FA/TR/FATRMEDWLHTH.dcc")
assert(composite.components.SH==nil)
assert(composite.rate==128 and composite.frames==8)
`
	if err := runtime.RunScoped(context.Background(), &Scope{}, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
}
