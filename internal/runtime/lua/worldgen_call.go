package modruntime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	lua "github.com/yuin/gopher-lua"
)

// GenerateWorldRecipe calls a trusted mod policy function and converts its
// admitted value snapshot back into the generic engine recipe. The function is
// deliberately mod-name agnostic; composition roots choose the Lua module and
// strategy while the adapter only transports serializable values.
func GenerateWorldRecipe(ctx context.Context, runtime *Runtime, moduleName, functionName string, arguments ...any) (*worldgen.Zone, error) {
	if runtime == nil || moduleName == "" || functionName == "" {
		return nil, fmt.Errorf("worldgen call requires runtime, module, and function")
	}
	var definition worldgen.Definition
	err := runtime.Run(ctx, func(state *lua.LState) error {
		require, ok := state.GetGlobal("require").(*lua.LFunction)
		if !ok {
			return fmt.Errorf("Lua require function is unavailable")
		}
		if err := state.CallByParam(lua.P{Fn: require, NRet: 1, Protect: true}, lua.LString(moduleName)); err != nil {
			return err
		}
		module := state.Get(-1)
		state.Pop(1)
		table, ok := module.(*lua.LTable)
		if !ok {
			return fmt.Errorf("worldgen module %q is not a table", moduleName)
		}
		function, ok := table.RawGetString(functionName).(*lua.LFunction)
		if !ok {
			return fmt.Errorf("worldgen function %q.%s is absent", moduleName, functionName)
		}
		values := make([]lua.LValue, 0, len(arguments))
		for _, argument := range arguments {
			value, err := goToLua(state, argument, 0)
			if err != nil {
				return err
			}
			values = append(values, value)
		}
		if err := state.CallByParam(lua.P{Fn: function, NRet: 1, Protect: true}, values...); err != nil {
			return err
		}
		result := state.Get(-1)
		state.Pop(1)
		converted, err := luaToGo(result, 0)
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(converted)
		if err != nil {
			return err
		}
		return json.Unmarshal(encoded, &definition)
	})
	if err != nil {
		return nil, fmt.Errorf("generate world recipe with %s.%s: %w", moduleName, functionName, err)
	}
	return worldgen.NewZone(definition)
}
