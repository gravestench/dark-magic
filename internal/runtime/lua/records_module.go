package modruntime

import (
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

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
					"engine.records.load(table[, columns])",
					"Load independent rows, optionally selecting named columns.",
				),
				"reload": commandHelp(
					"engine.records.reload()",
					"Invalidate and reload the record catalog.",
				),
				"revision": commandHelp(
					"engine.records.revision()",
					"Return an opaque equality token changed by any record invalidation, or nil for an unversioned source.",
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

					var columns []string

					projected := state.Get(2) != lua.LNil
					if projected {
						list := state.CheckTable(2)
						columns = make([]string, list.Len())

						count := 0

						list.ForEach(func(key, value lua.LValue) {
							position, numeric := key.(lua.LNumber)
							text, stringValue := value.(lua.LString)

							if !numeric || position < 1 || position > lua.LNumber(len(columns)) ||
								position != lua.LNumber(int(position)) || !stringValue {
								state.ArgError(2, "expected a contiguous list of column names")
							}

							columns[int(position)-1] = string(text)
							count++
						})

						if count != len(columns) {
							state.ArgError(2, "expected a contiguous list of column names")
						}
					}

					rows, err := records.Load(path)
					if err != nil {
						state.Push(lua.LNil)
						state.Push(lua.LString(err.Error()))

						return 2
					}

					result := state.CreateTable(len(rows), 0)
					for _, row := range rows {
						capacity := len(row)
						if projected {
							capacity = min(capacity, len(columns))
						}

						entry := state.CreateTable(0, capacity)

						if projected {
							for _, key := range columns {
								if value, found := row[key]; found {
									entry.RawSetString(key, lua.LString(value))
								}
							}
						} else {
							for key, value := range row {
								entry.RawSetString(key, lua.LString(value))
							}
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
				"revision": func(state *lua.LState) int {
					if source, ok := records.(interface{ Revision() uint64 }); ok {
						state.Push(lua.LString(strconv.FormatUint(source.Revision(), 10)))
					} else {
						state.Push(lua.LNil)
					}

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
