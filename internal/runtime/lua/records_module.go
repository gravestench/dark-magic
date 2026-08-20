package modruntime

import lua "github.com/yuin/gopher-lua"

type recordsGateway interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

// RecordsModule exposes generic cached TSV records from layered content.
func RecordsModule(records recordsGateway) Module {
	return Module{
		Name: "engine.records/v1",
		Help: documentedModule(
			"Load and inspect raw tabular content records.",
			map[string]CommandHelp{
				"load": commandHelp(
					"engine.records.load(table)",
					"Load records from a named game-data table.",
				),
				"reload": commandHelp(
					"engine.records.reload()",
					"Invalidate and reload the record catalog.",
				),
				"loaded": commandHelp(
					"engine.records.loaded()",
					"List the record tables currently loaded.",
				),
			},
		),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"load": func(state *lua.LState) int {
					path := state.CheckString(1)

					rows, err := records.Load(path)
					if err != nil {
						state.Push(lua.LNil)
						state.Push(lua.LString(err.Error()))

						return 2
					}

					result := state.NewTable()
					for _, row := range rows {
						entry := state.NewTable()
						for key, value := range row {
							entry.RawSetString(key, lua.LString(value))
						}

						result.Append(entry)
					}

					state.Push(result)

					return 1
				},
				"reload": func(state *lua.LState) int {
					path := state.CheckString(1)
					records.Invalidate(path)
					state.Push(lua.LTrue)

					return 1
				},
				"loaded": func(state *lua.LState) int {
					state.Push(lua.LBool(records.Loaded(state.CheckString(1))))
					return 1
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}
