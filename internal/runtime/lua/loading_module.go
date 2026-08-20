package modruntime

import (
	"context"

	"github.com/gravestench/dark-magic/internal/loading"
	lua "github.com/yuin/gopher-lua"
)

// LoadingModule exposes read-only progress for engine-owned transition work.
func LoadingModule(coordinator *loading.Coordinator) Module {
	return Module{
		Name: "engine.loading/v1",
		Help: documentedModule("Coordinate and inspect game loading work.", map[string]CommandHelp{
			"begin": commandHelp(
				"engine.loading.begin()",
				"Begin the configured loading sequence.",
			),
			"status": commandHelp(
				"engine.loading.status()",
				"Return the current loading progress snapshot.",
			),
		}),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"begin": func(state *lua.LState) int {
					table := state.CheckTable(1)
					ids := make([]string, 0, table.Len())
					table.ForEach(func(_, value lua.LValue) {
						ids = append(ids, lua.LVAsString(value))
					})

					if err := coordinator.Begin(context.Background(), ids); err != nil {
						state.RaiseError("%v", err)
					}

					return 0
				},
				"status": func(state *lua.LState) int {
					snapshot := coordinator.Snapshot()
					result := state.NewTable()
					result.RawSetString("state", lua.LString(snapshot.State))
					result.RawSetString("progress", lua.LNumber(snapshot.Progress()))
					result.RawSetString("completed", lua.LNumber(snapshot.Completed))
					result.RawSetString("total", lua.LNumber(snapshot.Total))
					result.RawSetString("current", lua.LString(snapshot.Current))

					if snapshot.Err != nil {
						result.RawSetString("error", lua.LString(snapshot.Err.Error()))
					}

					state.Push(result)

					return 1
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}
