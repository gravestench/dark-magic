package modruntime

import (
	"github.com/gravestench/dark-magic/internal/game/data/store"
	lua "github.com/yuin/gopher-lua"
)

// RecordsModule exposes generic cached TSV records from layered content.
func RecordsModule(records *recordstore.Store) Module {
	return Module{Name: "dm.records/v1", Loader: func(state *lua.LState) int {
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
	}}
}
