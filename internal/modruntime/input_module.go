package modruntime

import (
	"github.com/gravestench/dark-magic/internal/inputcore"
	lua "github.com/yuin/gopher-lua"
)

// InputModule exposes frame-stable logical actions rather than backend key
// codes or direct Raylib calls.
func InputModule(input *inputcore.Store) Module {
	return Module{Name: "dm.input/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"down": func(state *lua.LState) int {
				action := input.Snapshot().Actions[state.CheckString(1)]
				state.Push(lua.LBool(action.Down || action.Pressed))
				return 1
			},
			"pressed": func(state *lua.LState) int {
				action := input.Snapshot().Actions[state.CheckString(1)]
				state.Push(lua.LBool(action.Pressed))
				return 1
			},
			"released": func(state *lua.LState) int {
				action := input.Snapshot().Actions[state.CheckString(1)]
				state.Push(lua.LBool(action.Released))
				return 1
			},
			"cursor": func(state *lua.LState) int {
				frame := input.Snapshot()
				state.Push(lua.LNumber(frame.CursorX))
				state.Push(lua.LNumber(frame.CursorY))
				return 2
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
