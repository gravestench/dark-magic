package modruntime

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// DeterministicModule exposes pure, stateless keyed draws. It is useful when a
// mod must derive a stable choice from a persisted seed without consuming a
// session RNG stream. Calling the function in a different order cannot change
// its result, which makes it suitable for world recipes and content variants.
func DeterministicModule() Module {
	return Module{
		Name: "engine.deterministic/v1",
		Help: documentedModule("Derive stable values from a seed and purpose without mutable random state.", map[string]CommandHelp{
			"integer": commandHelp("engine.deterministic.integer(seed, purpose, limit [, ordinal])", "Return a stable zero-based integer below limit."),
		}),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"integer": deterministicInteger,
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)
			return 1
		},
	}
}

func deterministicInteger(state *lua.LState) int {
	seed := exactLuaUint(state, 1, "seed")
	purpose := strings.TrimSpace(state.CheckString(2))
	limit := exactLuaUint(state, 3, "limit")
	ordinal := uint64(0)
	if state.GetTop() >= 4 {
		ordinal = exactLuaUint(state, 4, "ordinal")
	}
	if purpose == "" {
		state.ArgError(2, "purpose is required")
	}
	if limit == 0 {
		state.ArgError(3, "limit must be greater than zero")
	}

	var input [16]byte
	binary.LittleEndian.PutUint64(input[:8], seed)
	binary.LittleEndian.PutUint64(input[8:], ordinal)
	hash := sha256.New()
	_, _ = hash.Write(input[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(purpose))
	digest := hash.Sum(nil)
	value := binary.LittleEndian.Uint64(digest[:8]) % limit
	state.Push(lua.LNumber(value))
	return 1
}

func exactLuaUint(state *lua.LState, index int, label string) uint64 {
	value := float64(state.CheckNumber(index))
	if value < 0 || value > float64(maximumExactLuaInteger) || math.Trunc(value) != value {
		state.ArgError(index, fmt.Sprintf("%s must be an exact non-negative integer no greater than %d", label, maximumExactLuaInteger))
	}
	return uint64(value)
}
