package catalog

import (
	"fmt"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// MapModule exposes immutable executable-derived DS1 and object relationships.
// Lua owns every policy decision made from these records.
func MapModule(source snapshotter) modruntime.Module {
	return modruntime.Module{
		Name: "d2legacy.map_catalog/v1",
		Help: modruntime.ModuleHelp{Summary: "Read recovered DS1 and map-object relationships.", Commands: map[string]modruntime.CommandHelp{
			"ds1_type": {Usage: "d2legacy.map_catalog.ds1_type(definition)", Summary: "Return the recovered DS1 definition and level type."},
			"object":   {Usage: "d2legacy.map_catalog.object(act, id)", Summary: "Resolve an act-local DS1 object ID to Objects.txt."},
		}},
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"ds1_type": func(state *lua.LState) int {
					snapshot, err := source.Snapshot()
					if err != nil {
						return raise(state, err)
					}
					entry, found := snapshot.DS1TypeByDef[state.CheckInt(1)]
					if !found {
						state.Push(lua.LNil)
						return 1
					}
					result := state.NewTable()
					result.RawSetString("name", lua.LString(entry.Name))
					result.RawSetString("definition", lua.LNumber(entry.Definition))
					result.RawSetString("level_type", lua.LNumber(entry.LevelType))
					state.Push(result)
					return 1
				},
				"object": func(state *lua.LState) int {
					snapshot, err := source.Snapshot()
					if err != nil {
						return raise(state, err)
					}
					key := fmt.Sprintf("%d:%d", state.CheckInt(1), state.CheckInt(2))
					entry, found := snapshot.MapObjectByActID[key]
					if !found {
						state.Push(lua.LNil)
						return 1
					}
					result := state.NewTable()
					result.RawSetString("act", lua.LNumber(entry.Act))
					result.RawSetString("id", lua.LNumber(entry.ID))
					result.RawSetString("description", lua.LString(entry.Description))
					result.RawSetString("object_id", lua.LNumber(entry.ObjectID))
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
