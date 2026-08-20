package modruntime

import (
	lua "github.com/yuin/gopher-lua"
)

// PlayerIntentController is the generic presentation-facing mailbox contract.
// The engine Lua host does not know which mod schemas or command names the
// concrete controller uses behind this interface.
type PlayerIntentController interface {
	SetRunning(bool)
	SetMoveTargetWithRadius(float64, float64, float64) error
	HasMoveTarget() bool
}

// PlayerControlModule accepts local presentation intents without granting Lua
// direct access to the authoritative session or ECS world.
func PlayerControlModule(controller PlayerIntentController) Module {
	return Module{
		Name: "engine.player/v1",
		Help: documentedModule(
			"Request local-player actions through the authoritative fixed-tick command source.",
			map[string]CommandHelp{
				"request_running": commandHelp(
					"engine.player.request_running(running)",
					"Request walk or run mode for the next admitted movement command.",
				),
				"request_move": commandHelp(
					"engine.player.request_move(x, y)",
					"Request movement toward an authoritative world-subtile target.",
				),
				"movement_pending": commandHelp(
					"engine.player.movement_pending()",
					"Report whether a pointer path target remains active.",
				),
			},
		),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"request_running": func(state *lua.LState) int {
					controller.SetRunning(state.CheckBool(1))
					return 0
				},
				"request_move": func(state *lua.LState) int {
					stopRadius := 0.0
					if state.GetTop() >= 3 {
						stopRadius = float64(state.CheckNumber(3))
					}

					if err := controller.SetMoveTargetWithRadius(
						float64(state.CheckNumber(1)),
						float64(state.CheckNumber(2)),
						stopRadius,
					); err != nil {
						state.RaiseError("%v", err)
					}

					return 0
				},
				"movement_pending": func(state *lua.LState) int { state.Push(lua.LBool(controller.HasMoveTarget())); return 1 },
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}
