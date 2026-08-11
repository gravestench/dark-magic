package modruntime

import lua "github.com/yuin/gopher-lua"

// DevModule exposes copied, read-only launch values to development scenes.
// Keeping this tiny prevents diagnostic flags from leaking into game authority.
func DevModule(values map[string]any) Module {
	seed, _ := values["random_seed"].(int)
	if seed <= 0 {
		seed = 1
	}
	return Module{Name: "dm.dev/v1", Help: documentedModule("Read development-only launch options.", map[string]CommandHelp{
		"option": commandHelp("dm.dev.option(name)", "Return a copied development launch option."),
		"seed":   commandHelp("dm.dev.seed()", "Return a different positive pseudo-random seed on each call."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"seed": func(state *lua.LState) int {
				seed = seed * 48271 % 2147483647
				if seed == 0 {
					seed = 1
				}
				state.Push(lua.LNumber(seed))
				return 1
			},
			"option": func(state *lua.LState) int {
				value, found := values[state.CheckString(1)]
				if !found {
					state.Push(lua.LNil)
					return 1
				}
				switch typed := value.(type) {
				case string:
					state.Push(lua.LString(typed))
				case int:
					state.Push(lua.LNumber(typed))
				case bool:
					state.Push(lua.LBool(typed))
				default:
					state.Push(lua.LNil)
				}
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
