package modruntime

import (
	"github.com/gravestench/dark-magic/internal/inputstate"
	lua "github.com/yuin/gopher-lua"
)

// InputModule exposes frame-stable logical actions rather than backend key
// codes or direct Raylib calls.
func InputModule(input *inputstate.Store) Module {
	return Module{Name: "dm.input/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"text": func(state *lua.LState) int {
				state.Push(lua.LString(input.Text()))
				return 1
			},
			"down": func(state *lua.LState) int {
				action := input.Action(state.CheckString(1))
				state.Push(lua.LBool(action.Down || action.Pressed))
				return 1
			},
			"pressed": func(state *lua.LState) int {
				action := input.Action(state.CheckString(1))
				state.Push(lua.LBool(action.Pressed))
				return 1
			},
			"released": func(state *lua.LState) int {
				action := input.Action(state.CheckString(1))
				state.Push(lua.LBool(action.Released))
				return 1
			},
			"cursor": func(state *lua.LState) int {
				x, y := input.Cursor()
				state.Push(lua.LNumber(x))
				state.Push(lua.LNumber(y))
				return 2
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
