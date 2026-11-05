package modruntime

import (
	"fmt"
	"math"

	lua "github.com/yuin/gopher-lua"
)

// IntegerModule exposes pure fixed-width arithmetic. Wide products cross the
// Lua boundary as two exact u32 words rather than one rounded floating value.
func IntegerModule() Module {
	return Module{
		Name: "engine.integer/v1",
		Help: documentedModule(
			"Calculate exact fixed-width integer results without mutable state.",
			map[string]CommandHelp{
				"mul_add_u32": commandHelp(
					"engine.integer.mul_add_u32(a, b, c)",
					"Return the exact low and high u32 words of a * b + c for u32 arguments.",
				),
				"band_u32": commandHelp(
					"engine.integer.band_u32(a, b)",
					"Return the bitwise AND of two u32 values.",
				),
				"sub_i32": commandHelp(
					"engine.integer.sub_i32(a, b)",
					"Return a - b wrapped to i32.",
				),
				"mul_i32": commandHelp(
					"engine.integer.mul_i32(a, b)",
					"Return a * b wrapped to i32.",
				),
				"abs_i32": commandHelp(
					"engine.integer.abs_i32(a)",
					"Return the wrapping i32 absolute value; minimum i32 remains minimum i32.",
				),
				"div_i32": commandHelp(
					"engine.integer.div_i32(a, b)",
					"Divide i32 values toward zero; reject division by zero or signed overflow.",
				),
				"add_u16": commandHelp(
					"engine.integer.add_u16(a, b)",
					"Add an i32 to a u16 and return the low u16 bits.",
				),
				"high_u16": commandHelp(
					"engine.integer.high_u16(bits)",
					"Return the unsigned high 16 bits of a u32 value.",
				),
			},
		),
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"mul_add_u32": integerMulAddU32,
				"band_u32":    integerBandU32,
				"sub_i32":     integerSubI32,
				"mul_i32":     integerMulI32,
				"abs_i32":     integerAbsI32,
				"div_i32":     integerDivI32,
				"add_u16":     integerAddU16,
				"high_u16":    integerHighU16,
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)
			return 1
		},
	}
}

func integerMulAddU32(state *lua.LState) int {
	a := checkedIntegerU32(state, 1)
	b := checkedIntegerU32(state, 2)
	c := checkedIntegerU32(state, 3)
	// Even the largest u32 inputs give at most 2^64 - 2^32.
	value := uint64(a)*uint64(b) + uint64(c)
	state.Push(lua.LNumber(uint32(value)))
	state.Push(lua.LNumber(uint32(value >> 32)))
	return 2
}

func integerBandU32(state *lua.LState) int {
	a := checkedIntegerU32(state, 1)
	b := checkedIntegerU32(state, 2)
	state.Push(lua.LNumber(a & b))
	return 1
}

func integerSubI32(state *lua.LState) int {
	a := checkedIntegerI32(state, 1)
	b := checkedIntegerI32(state, 2)
	state.Push(lua.LNumber(int32(uint32(a) - uint32(b))))
	return 1
}

func integerMulI32(state *lua.LState) int {
	a := checkedIntegerI32(state, 1)
	b := checkedIntegerI32(state, 2)
	state.Push(lua.LNumber(int32(uint32(a) * uint32(b))))
	return 1
}

func integerAbsI32(state *lua.LState) int {
	a := checkedIntegerI32(state, 1)
	if a < 0 {
		a = int32(uint32(0) - uint32(a))
	}
	state.Push(lua.LNumber(a))
	return 1
}

func integerDivI32(state *lua.LState) int {
	a := checkedIntegerI32(state, 1)
	b := checkedIntegerI32(state, 2)
	if b == 0 || (a == math.MinInt32 && b == -1) {
		state.RaiseError("i32 division by zero or signed overflow")
		return 0
	}
	state.Push(lua.LNumber(a / b))
	return 1
}

func integerAddU16(state *lua.LState) int {
	a := uint16(checkedIntegerNumber(state, 1, 0, math.MaxUint16, "u16"))
	b := checkedIntegerI32(state, 2)
	state.Push(lua.LNumber(uint16(uint32(a) + uint32(b))))
	return 1
}

func integerHighU16(state *lua.LState) int {
	bits := checkedIntegerU32(state, 1)
	state.Push(lua.LNumber(bits >> 16))
	return 1
}

func checkedIntegerU32(state *lua.LState, index int) uint32 {
	return uint32(checkedIntegerNumber(state, index, 0, math.MaxUint32, "u32"))
}

func checkedIntegerI32(state *lua.LState, index int) int32 {
	return int32(checkedIntegerNumber(state, index, math.MinInt32, math.MaxInt32, "i32"))
}

// Check the number and its domain before any fixed-width conversion. Checking
// the Lua type explicitly also prevents numeric strings from being coerced.
func checkedIntegerNumber(state *lua.LState, index int, minimum, maximum float64, domain string) float64 {
	value, ok := state.Get(index).(lua.LNumber)
	number := float64(value)
	if !ok || math.IsNaN(number) || math.IsInf(number, 0) ||
		math.Trunc(number) != number || number < minimum || number > maximum {
		state.ArgError(index, fmt.Sprintf("expected a finite exact %s integer", domain))
	}
	return number
}
