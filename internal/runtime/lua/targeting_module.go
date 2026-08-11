package modruntime

import (
	"github.com/gravestench/dark-magic/internal/game/targeting"
	lua "github.com/yuin/gopher-lua"
)

func TargetingModule(resolver *targeting.Resolver) Module {
	return Module{
		Name: "dm.targeting/v1",
		Help: documentedModule("Resolve copied authoritative spawned-entity pointer facts.", map[string]CommandHelp{
			"selectable_at": commandHelp("dm.targeting.selectable_at(x, y)", "Return the highest-priority spawned entity footprint under a world-subtile point."),
		}, nil),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"selectable_at": func(state *lua.LState) int {
					hit, found := resolver.HitAt(float64(state.CheckNumber(1)), float64(state.CheckNumber(2)))
					if !found {
						state.Push(lua.LNil)
						return 1
					}
					result := state.NewTable()
					result.RawSetString("id", lua.LString(hit.ID))
					result.RawSetString("kind", lua.LString(hit.Kind))
					result.RawSetString("label", lua.LString(hit.Label))
					result.RawSetString("owner", lua.LString(hit.Owner))
					result.RawSetString("x", lua.LNumber(hit.X))
					result.RawSetString("y", lua.LNumber(hit.Y))
					result.RawSetString("radius", lua.LNumber(hit.Radius))
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
