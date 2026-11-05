package modruntime

import (
	"fmt"
	"math"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// TestIntegerModuleNRFStepWords catches rounded products and lost high words.
// NRF-V01 supplies arithmetic stages; RNG policy sequencing is tested separately.
func TestIntegerModuleNRFStepWords(t *testing.T) {
	runtime := newIntegerTestRuntime(t)
	err := runtime.Run(t.Context(), func(state *lua.LState) error {
		return state.DoString(`
local integer = require("engine.integer/v1")
assert(integer.api == 1)
local function words(a, b, c, want_lo, want_hi)
    assert(select("#", integer.mul_add_u32(a, b, c)) == 2)
    local lo, hi = integer.mul_add_u32(a, b, c)
    assert(lo == want_lo and hi == want_hi)
end
words(1, 0x6AC690C5, 666, 0x6AC6935F, 0)
words(0x6AC6935F, 0x6AC690C5, 0, 0x2F2ED81B, 0x2C890AFD)
`)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestIntegerModuleKNBWidthBoundaries catches widened signed arithmetic,
// floor division, saturation, and sign extension of an unsigned high word.
func TestIntegerModuleKNBWidthBoundaries(t *testing.T) {
	runtime := newIntegerTestRuntime(t)
	err := runtime.Run(t.Context(), func(state *lua.LState) error {
		return state.DoString(`
local integer = require("engine.integer/v1")
local function scalar(want, fn, ...)
    assert(select("#", fn(...)) == 1)
    assert(fn(...) == want)
end
-- KNB-V34 and KNB-V23, at their supplied arithmetic boundaries.
scalar(-2147483648, integer.sub_i32, 0, -2147483648)
scalar(-2147483648, integer.abs_i32, -2147483648)
scalar(-2147483345, integer.mul_i32, 3, -2147483547)
scalar(0, integer.div_i32, -2147483345, 2147483547)
scalar(0, integer.div_i32, -6, 7)
scalar(2, integer.add_u16, 65535, 3)
scalar(67, integer.band_u32, 0x6AC6935F, 99)
scalar(65535, integer.high_u16, 0xFFFF0000)
-- Generic engineering boundaries, independent of original-game vectors.
scalar(2147483647, integer.sub_i32, -2147483648, 1)
scalar(0, integer.mul_i32, -2147483648, 2)
scalar(7, integer.abs_i32, -7)
scalar(7, integer.abs_i32, 7)
scalar(0, integer.abs_i32, 0)
scalar(-2, integer.div_i32, 7, -3)
scalar(2, integer.div_i32, -7, -3)
scalar(-2147483648, integer.div_i32, -2147483648, 1)
scalar(65535, integer.add_u16, 0, -1)
scalar(0, integer.add_u16, 0, -2147483648)
scalar(65535, integer.add_u16, 0, 2147483647)
scalar(4294967295, integer.band_u32, 4294967295, 4294967295)
scalar(0, integer.high_u16, 65535)
scalar(65535, integer.high_u16, 4294967295)
`)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestIntegerModuleRejectsInexactInputs catches argument coercion or conversion
// before validation, including invalid values in non-leading argument slots.
func TestIntegerModuleRejectsInexactInputs(t *testing.T) {
	runtime := newIntegerTestRuntime(t)
	type invalidInput struct {
		name  string
		value lua.LValue
	}
	functions := []struct {
		name    string
		domains []string
	}{
		{"mul_add_u32", []string{"u32", "u32", "u32"}},
		{"band_u32", []string{"u32", "u32"}},
		{"sub_i32", []string{"i32", "i32"}},
		{"mul_i32", []string{"i32", "i32"}},
		{"abs_i32", []string{"i32"}},
		{"div_i32", []string{"i32", "i32"}},
		{"add_u16", []string{"u16", "i32"}},
		{"high_u16", []string{"u32"}},
	}
	for _, function := range functions {
		for index, domain := range function.domains {
			invalid := []invalidInput{
				{"fraction", lua.LNumber(0.5)},
				{"nan", lua.LNumber(math.NaN())},
				{"positive_infinity", lua.LNumber(math.Inf(1))},
				{"negative_infinity", lua.LNumber(math.Inf(-1))},
				{"numeric_string", lua.LString("1")},
				{"nil", lua.LNil},
				{"boolean", lua.LTrue},
			}
			if domain == "i32" {
				invalid = append(invalid,
					invalidInput{"below_minimum", lua.LNumber(-2147483649)},
					invalidInput{"above_maximum", lua.LNumber(2147483648)},
				)
			} else {
				maximum := lua.LNumber(4294967296)
				if domain == "u16" {
					maximum = 65536
				}
				invalid = append(invalid,
					invalidInput{"negative", lua.LNumber(-1)},
					invalidInput{"above_maximum", maximum},
				)
			}
			for _, input := range invalid {
				t.Run(fmt.Sprintf("%s/argument_%d/%s", function.name, index+1, input.name), func(t *testing.T) {
					err := runtime.Run(t.Context(), func(state *lua.LState) error {
						args := state.NewTable()
						for i := range function.domains {
							args.RawSetInt(i+1, lua.LNumber(1))
						}
						args.RawSetInt(index+1, input.value)
						state.SetGlobal("integer_args", args)
						return state.DoString(fmt.Sprintf(`
local integer = require("engine.integer/v1")
local results = {pcall(integer.%s, unpack(integer_args, 1, %d))}
assert(results[1] == false, "invalid input accepted")
assert(#results == 2 and type(results[2]) == "string", "error returned a numeric result")
`, function.name, len(function.domains)))
					})
					if err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
	err := runtime.Run(t.Context(), func(state *lua.LState) error {
		return state.DoString(`
local integer = require("engine.integer/v1")
for _, args in ipairs({{1, 0}, {-2147483648, -1}}) do
    local results = {pcall(integer.div_i32, args[1], args[2])}
    assert(results[1] == false)
    assert(#results == 2 and type(results[2]) == "string")
end
assert(select("#", integer.mul_add_u32(4294967295, 4294967295, 4294967295)) == 2)
local lo, hi = integer.mul_add_u32(4294967295, 4294967295, 4294967295)
assert(lo == 0 and hi == 4294967295)
`)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func newIntegerTestRuntime(t *testing.T) *Runtime {
	t.Helper()
	runtime := New()
	if err := runtime.RegisterModule(IntegerModule()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(t.Context()) })
	return runtime
}
