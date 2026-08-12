package modruntime

import (
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/targeting"
	lua "github.com/yuin/gopher-lua"
)

// TargetingModule is a temporary d2legacy presentation adapter. It stays out of
// the authoritative capability set and will disappear once pointer selection
// consumes a generic spatial-query primitive directly.
func TargetingModule(resolver *targeting.Resolver) Module {
	return Module{
		Name: "d2legacy.targeting/v1",
		Help: documentedModule("Resolve copied authoritative spawned-entity pointer facts.", map[string]CommandHelp{
			"selectable_at": commandHelp("engine.targeting.selectable_at(x, y)", "Return the highest-priority spawned entity footprint under a world-subtile point."),
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
