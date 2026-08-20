package modruntime

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// Call invokes one trusted module function with serializable values and returns
// a copied serializable result. It is the policy-neutral escape hatch used by
// narrow adapters that need a mod decision without exposing native objects.
func Call(
	ctx context.Context,
	runtime *Runtime,
	moduleName, functionName string,
	arguments ...any,
) (any, error) {
	if runtime == nil || moduleName == "" || functionName == "" {
		//nolint:staticcheck // Preserve established diagnostic capitalization for compatibility.
		return nil, fmt.Errorf("Lua call requires runtime, module, and function")
	}

	var result any

	err := runtime.Run(ctx, func(state *lua.LState) error {
		require, ok := state.GetGlobal("require").(*lua.LFunction)
		if !ok {
			//nolint:staticcheck // Preserve established diagnostic capitalization for compatibility.
			return fmt.Errorf("Lua require function is unavailable")
		}

		if err := state.CallByParam(
			lua.P{Fn: require, NRet: 1, Protect: true},
			lua.LString(moduleName),
		); err != nil {
			return err
		}

		module, ok := state.Get(-1).(*lua.LTable)
		state.Pop(1)

		if !ok {
			//nolint:staticcheck // Preserve established diagnostic capitalization for compatibility.
			return fmt.Errorf("Lua module %q is not a table", moduleName)
		}

		function, ok := module.RawGetString(functionName).(*lua.LFunction)
		if !ok {
			//nolint:staticcheck // Preserve established diagnostic capitalization for compatibility.
			return fmt.Errorf("Lua function %q.%s is absent", moduleName, functionName)
		}

		values := make([]lua.LValue, 0, len(arguments))
		for _, argument := range arguments {
			value, err := goToLua(state, argument, 0)
			if err != nil {
				return err
			}

			values = append(values, value)
		}

		if err := state.CallByParam(
			lua.P{Fn: function, NRet: 1, Protect: true},
			values...); err != nil {
			return err
		}

		value := state.Get(-1)
		state.Pop(1)

		converted, err := luaToGo(value, 0)
		if err != nil {
			return err
		}

		result = converted

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("call Lua function %s.%s: %w", moduleName, functionName, err)
	}

	return result, nil
}
