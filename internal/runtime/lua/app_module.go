package modruntime

import lua "github.com/yuin/gopher-lua"

// AppModule exposes narrow client-process information and exit intent.
func AppModule(version string, requestExit func()) Module {
	return Module{Name: "dm.app/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"request_exit": func(state *lua.LState) int {
				if requestExit != nil {
					requestExit()
				}
				return 0
			},
			"version": func(state *lua.LState) int { state.Push(lua.LString(version)); return 1 },
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
