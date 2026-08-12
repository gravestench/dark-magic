package targeting

import (
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// TargetingModule is a temporary d2legacy presentation adapter. It stays out of
// the authoritative capability set and will disappear once pointer selection
// consumes a generic spatial-query primitive directly.
func Module(resolver *Resolver) modruntime.Module {
	return modruntime.Module{
		Name: "d2legacy.targeting/v1",
		Help: modruntime.ModuleHelp{Summary: "Resolve copied authoritative spawned-entity pointer facts.", Commands: map[string]modruntime.CommandHelp{
			"selectable_at": {Usage: "d2legacy.targeting.selectable_at(x, y)", Summary: "Return the highest-priority spawned entity footprint under a world-subtile point."},
		}},
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
