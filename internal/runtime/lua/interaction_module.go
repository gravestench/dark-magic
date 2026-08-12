package modruntime

import (
	gameinteraction "github.com/gravestench/dark-magic/internal/game/interaction"
	lua "github.com/yuin/gopher-lua"
)

// InteractionModule exposes a read-only copy of the player's current NPC
// context. Scripts may ask to open or close a known target, but the request is
// only applied by the fixed-tick session on its next update.
func InteractionModule(authority *gameinteraction.Authority, controller *gameinteraction.Controller, owner string) Module {
	return Module{Name: "engine.interaction/v1", Help: documentedModule("Inspect and request authoritative NPC interactions.", map[string]CommandHelp{
		"snapshot": commandHelp("engine.interaction.snapshot()", "Return a copy of the active NPC, vendor categories, and services."),
		"open":     commandHelp("engine.interaction.open(target_id)", "Request entry into a server-known NPC interaction."),
		"open_at":  commandHelp("engine.interaction.open_at(x, y)", "Request authoritative selection and interaction at world-subtile coordinates."),
		"close":    commandHelp("engine.interaction.close()", "Request leaving the current NPC interaction."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"snapshot": func(state *lua.LState) int {
				context, err := authority.Snapshot(owner)
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(interactionTable(state, context))
				return 1
			},
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

func interactionTable(state *lua.LState, context gameinteraction.Context) *lua.LTable {
	result := state.NewTable()
	result.RawSetString("active", lua.LBool(context.TargetID != ""))
	result.RawSetString("target_id", lua.LString(context.TargetID))
	result.RawSetString("npc", lua.LString(context.NPC))
	result.RawSetString("vendor", lua.LString(context.Vendor))
	result.RawSetString("categories", stringTable(state, context.Categories))
	result.RawSetString("services", stringTable(state, context.Services))
	return result
}

func stringTable(state *lua.LState, values []string) *lua.LTable {
	result := state.NewTable()
	for _, value := range values {
		result.Append(lua.LString(value))
	}
	return result
}
