package modruntime

import (
	"sync"

	"github.com/gravestench/dark-magic/internal/presentation/scene"
	lua "github.com/yuin/gopher-lua"
)

// Simulation owns deterministic gameplay state outside disposable Lua
// components. Lua orchestrates it through values and commands, never pointers.
type Simulation struct {
	mu    sync.RWMutex
	world *scene.State
}

func NewSimulation(world *scene.State) *Simulation { return &Simulation{world: world} }

func (s *Simulation) Move(dx, dy float64) {
	s.mu.Lock()
	s.world.MoveHero(dx, dy)
	s.mu.Unlock()
}

func (s *Simulation) Snapshot() scene.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.world
}

func SimulationModule(simulation *Simulation) Module {
	return Module{Name: "engine.simulation/v1", Help: documentedModule("Inspect and command the deterministic game simulation.", map[string]CommandHelp{
		"move_hero": commandHelp("engine.simulation.move_hero(dx, dy)", "Move the simulated hero by a relative amount."),
		"state":     commandHelp("engine.simulation.state()", "Return the current simulation snapshot."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"move_hero": func(state *lua.LState) int {
				simulation.Move(float64(state.CheckNumber(1)), float64(state.CheckNumber(2)))
				return 0
			},
			"state": func(state *lua.LState) int {
				world := simulation.Snapshot()
				result := state.NewTable()
				result.RawSetString("seed", lua.LNumber(world.Seed))
				result.RawSetString("hero_x", lua.LNumber(world.Hero.X))
				result.RawSetString("hero_y", lua.LNumber(world.Hero.Y))
				result.RawSetString("camera_x", lua.LNumber(world.Camera.X))
				result.RawSetString("camera_y", lua.LNumber(world.Camera.Y))
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
