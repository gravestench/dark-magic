package modruntime

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	lua "github.com/yuin/gopher-lua"
)

// WorldgenModule admits a mod-authored recipe through the engine's generic
// canonicalization and validation boundary. It does not generate a map and it
// intentionally knows nothing about Diablo records or DRLG policy.
func WorldgenModule() Module {
	return Module{
		Name: "engine.worldgen/v1",
		Help: documentedModule(
			"Validate and canonicalize immutable mod-authored world recipes.",
			map[string]CommandHelp{
				"admit": commandHelp(
					"engine.worldgen.admit(definition)",
					"Validate a recipe and return its canonical snapshot and checksum.",
				),
			},
		),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"admit": admitWorldRecipe,
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}

// admitWorldRecipe owns the admit world recipe step at this boundary, keeping its side effects and failure point
// explicit to callers.
func admitWorldRecipe(state *lua.LState) int {
	value, err := luaToGo(state.CheckTable(1), 0)
	if err != nil {
		return pushLuaError(state, fmt.Errorf("worldgen recipe: %w", err))
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return pushLuaError(state, fmt.Errorf("worldgen recipe encoding: %w", err))
	}

	var definition worldgen.Definition
	if err := json.Unmarshal(encoded, &definition); err != nil {
		return pushLuaError(state, fmt.Errorf("worldgen recipe shape: %w", err))
	}

	zone, err := worldgen.NewZone(definition)
	if err != nil {
		return pushLuaError(state, err)
	}

	state.Push(worldZoneToLua(state, zone))

	return 1
}

// worldZoneToLua publishes world zone to lua in one canonical Lua representation, keeping scripts independent of Go
// storage details.
func worldZoneToLua(state *lua.LState, zone *worldgen.Zone) *lua.LTable {
	encoded, _ := zone.MarshalJSON()

	var snapshot any

	_ = json.Unmarshal(encoded, &snapshot)
	value, _ := goToLua(state, snapshot, 0)
	result := value.(*lua.LTable)
	checksum, _ := zone.Checksum()
	result.RawSetString("checksum", lua.LString(checksum))

	return result
}
