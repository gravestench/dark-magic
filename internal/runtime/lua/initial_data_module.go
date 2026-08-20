package modruntime

import lua "github.com/yuin/gopher-lua"

// InitialDataModule exposes immutable composition-time values to trusted mod
// startup code. It is policy-neutral: the host transports a copied value tree;
// the loaded mod decides which ECS state and rules those values create.
func InitialDataModule(values map[string]any) Module {
	return Module{
		Name: "engine.initial_data/v1",
		Help: documentedModule(
			"Read immutable host-provided session bootstrap values.",
			map[string]CommandHelp{
				"get": commandHelp(
					"engine.initial_data.get(name)",
					"Return a deep Lua copy of one bootstrap value.",
				),
			},
		),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"get": func(state *lua.LState) int {
					value, found := values[state.CheckString(1)]
					if !found {
						state.Push(lua.LNil)
						return 1
					}

					converted, err := goToLua(state, value, 0)
					if err != nil {
						state.RaiseError("copy initial data: %v", err)
						return 0
					}

					state.Push(converted)

					return 1
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}
