package modruntime

import (
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	lua "github.com/yuin/gopher-lua"
)

// PlayerControlModule accepts local presentation intents without granting Lua
// direct access to the authoritative session or ECS world.
func PlayerControlModule(controller *gamesession.MovementController) Module {
	return Module{Name: "dm.player/v1", Help: documentedModule("Request local-player actions through the authoritative fixed-tick command source.", map[string]CommandHelp{
		"request_running": commandHelp("dm.player.request_running(running)", "Request walk or run mode for the next admitted movement command."),
		"assign_skill":    commandHelp("dm.player.assign_skill(slot, skill_id)", "Request an authoritative left or right skill assignment."),
		"request_move":    commandHelp("dm.player.request_move(x, y)", "Request movement toward an authoritative world-subtile target."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"request_running": func(state *lua.LState) int {
				controller.SetRunning(state.CheckBool(1))
				return 0
			},
			"assign_skill": func(state *lua.LState) int {
				if err := controller.AssignSkill(state.CheckString(1), int64(state.CheckInt(2))); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"request_move": func(state *lua.LState) int {
				if err := controller.SetMoveTarget(float64(state.CheckNumber(1)), float64(state.CheckNumber(2))); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
