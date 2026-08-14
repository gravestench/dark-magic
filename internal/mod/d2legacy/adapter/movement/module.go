package movement

import (
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// RulesModule exposes the same movement resolver to authoritative Lua that the
// owning client's rollback/replay predictor uses in Go.
func RulesModule() modruntime.Module {
	return modruntime.Module{
		Name: "d2legacy.movement_rules/v1",
		Help: modruntime.ModuleHelp{
			Summary: "Resolve d2legacy movement intent with the production walk, run, diagonal, and arrival rules.",
			Commands: map[string]modruntime.CommandHelp{
				"velocity": {Usage: "d2legacy.movement_rules.velocity(x, y, payload)", Summary: "Return authoritative x/y velocity and whether movement is active."},
			},
		},
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"velocity": func(state *lua.LState) int {
					payload := state.CheckTable(3)
					move := MovePayload{
						X:       luaInteger(payload.RawGetString("x")),
						Y:       luaInteger(payload.RawGetString("y")),
						Running: lua.LVAsBool(payload.RawGetString("running")),
					}
					if target, ok := payload.RawGetString("target").(*lua.LTable); ok {
						move.Target = &MoveTarget{
							X:          luaNumber(target.RawGetString("x")),
							Y:          luaNumber(target.RawGetString("y")),
							StopRadius: luaOptionalNumber(target.RawGetString("stop_radius")),
						}
					}
					resolved := Resolve(gameworld.Point{X: float64(state.CheckNumber(1)), Y: float64(state.CheckNumber(2))}, move)
					state.Push(lua.LNumber(resolved.Velocity.X))
					state.Push(lua.LNumber(resolved.Velocity.Y))
					state.Push(lua.LBool(resolved.Moving))
					return 3
				},
			})
			state.Push(module)
			return 1
		},
	}
}

func luaInteger(value lua.LValue) int {
	number, _ := value.(lua.LNumber)
	return int(number)
}

func luaNumber(value lua.LValue) float64 {
	number, _ := value.(lua.LNumber)
	return float64(number)
}

func luaOptionalNumber(value lua.LValue) float64 {
	if value == lua.LNil {
		return 0
	}
	return luaNumber(value)
}
