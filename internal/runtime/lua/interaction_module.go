package modruntime

import (
	gameinteraction "github.com/gravestench/dark-magic/internal/game/interaction"
	lua "github.com/yuin/gopher-lua"
)

// InteractionModule exposes a read-only copy of the player's current NPC
// context. Scripts may ask to open or close a known target, but the request is
// only applied by the fixed-tick session on its next update.
func InteractionModule(controller *gameinteraction.Controller) Module {
	return Module{Name: "engine.interaction/v1", Help: documentedModule("Queue NPC interaction command intents.", map[string]CommandHelp{
		"open":    commandHelp("engine.interaction.open(target_id)", "Request entry into a server-known NPC interaction."),
		"open_at": commandHelp("engine.interaction.open_at(x, y)", "Request authoritative selection and interaction at world-subtile coordinates."),
		"close":   commandHelp("engine.interaction.close()", "Request leaving the current NPC interaction."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"open": func(state *lua.LState) int {
				if err := controller.Open(state.CheckString(1)); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"open_at": func(state *lua.LState) int {
				if err := controller.OpenAt(float64(state.CheckNumber(1)), float64(state.CheckNumber(2))); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"close": func(state *lua.LState) int {
				controller.Close()
				return 0
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
