package modruntime

import (
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

// Lua stores every number as a floating-point value. Integers above this limit
// cannot be represented exactly, so the authoritative API never returns a raw
// uint64 produced by the engine's random generator.
const maximumExactLuaInteger = uint64(1<<53 - 1)

// AuthorityRandomModule gives trusted gameplay code narrow access to the
// engine-owned deterministic streams. Names are registered by the mod's
// composition root; ordinary policy modules only consume their assigned name.
func AuthorityRandomModule(streams *simulation.RandomStreams) Module {
	return Module{
		Name: "engine.authority_random/v1",
		Loader: func(state *lua.LState) int {
			module := state.NewTable()
			state.SetFuncs(module, map[string]lua.LGFunction{
				"integer": randomInteger(streams),
				"chance":  randomChance(streams),
			})
			state.Push(module)
			return 1
		},
		Help: ModuleHelp{Summary: "Draw from registered deterministic authoritative random streams."},
	}
}

func randomInteger(streams *simulation.RandomStreams) lua.LGFunction {
	return func(state *lua.LState) int {
		name := state.CheckString(1)
		limit := checkExactPositiveInteger(state, 2, "limit")
		value, err := streams.Uint64n(name, limit)
		if err != nil {
			state.RaiseError("draw deterministic integer: %v", err)
			return 0
		}
		state.Push(lua.LNumber(value))
		return 1
	}
}

func randomChance(streams *simulation.RandomStreams) lua.LGFunction {
	return func(state *lua.LState) int {
		name := state.CheckString(1)
		numerator := checkExactNonNegativeInteger(state, 2, "numerator")
		denominator := checkExactPositiveInteger(state, 3, "denominator")
		if numerator > denominator {
			state.ArgError(2, "numerator must not exceed denominator")
			return 0
		}
		roll, err := streams.Uint64n(name, denominator)
		if err != nil {
			state.RaiseError("draw deterministic chance: %v", err)
			return 0
		}
		state.Push(lua.LBool(roll < numerator))
		return 1
	}
}

func checkExactPositiveInteger(state *lua.LState, index int, label string) uint64 {
	value := checkExactNonNegativeInteger(state, index, label)
	if value == 0 {
		state.ArgError(index, label+" must be greater than zero")
	}
	return value
}

func checkExactNonNegativeInteger(state *lua.LState, index int, label string) uint64 {
	value := state.CheckNumber(index)
	if value < 0 {
		state.ArgError(index, label+" must be non-negative")
	}
	integer := uint64(value)
	if value != lua.LNumber(integer) || integer > maximumExactLuaInteger {
		state.ArgError(index, fmt.Sprintf("%s must be an exact non-negative integer no greater than %d", label, maximumExactLuaInteger))
	}
	return integer
}
