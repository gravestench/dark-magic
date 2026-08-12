package modruntime

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// MapCatalogModule exposes recovered DS1 definition names and act-local object
// mappings without coupling Lua map generation to Riiablo's source layout.
func MapCatalogModule(catalog questCatalogSnapshotter) Module {
	return Module{Name: "engine.map_catalog/v1", Help: documentedModule("Query recovered DS1 and map-object relationships.", map[string]CommandHelp{
		"ds1_type": commandHelp("engine.map_catalog.ds1_type(definition)", "Return the recovered DS1 definition and level type."),
		"object":   commandHelp("engine.map_catalog.object(act, id)", "Resolve an act-local DS1 object ID to Objects.txt."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"ds1_type": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
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
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				entry, found := snapshot.MapObjectByActID[fmt.Sprintf("%d:%d", state.CheckInt(1), state.CheckInt(2))]
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
	}}
}
