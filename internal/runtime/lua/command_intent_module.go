package modruntime

import (
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	lua "github.com/yuin/gopher-lua"
)

// CommandIntentModule is the generic presentation-to-simulation doorway.
// It queues data only. Authentication, admission, validation, and mutation all
// happen later on the fixed simulation tick.
func CommandIntentModule(controller *gamesession.IntentController) Module {
	return Module{
		Name: "engine.command_intent/v1",
		Help: documentedModule(
			"Queue a mod-defined command request for fixed-tick admission.",
			map[string]CommandHelp{
				"submit": commandHelp(
					"engine.command_intent.submit(kind, payload)",
					"Queue one serializable command payload for the local player.",
				),
			},
		),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"submit": func(state *lua.LState) int {
					payload, err := luaToGo(state.CheckAny(2), 0)
					if err != nil {
						state.RaiseError("command intent payload: %v", err)
					}
					if err := controller.Submit(state.CheckString(1), payload); err != nil {
						state.RaiseError("%v", err)
					}
					return 0
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)
			return 1
		},
	}
}
