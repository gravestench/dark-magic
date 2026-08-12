package modruntime

import lua "github.com/yuin/gopher-lua"

// AppModule exposes narrow client-process information and exit intent.
func AppModule(version string, requestExit func()) Module {
	return Module{Name: "engine.app/v1", Help: documentedModule("Inspect and control the current Dark Magic process.", map[string]CommandHelp{
		"request_exit": commandHelp("engine.app.request_exit()", "Request an orderly application shutdown."),
		"version":      commandHelp("engine.app.version()", "Return the engine build version."),
	}), Loader: func(state *lua.LState) int {
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
