package character_gen

import (
	"fmt"
	lua "github.com/yuin/gopher-lua"
)

// LuaIntegration represents the integration with Lua scripting.
type LuaIntegration struct {
	L *lua.LState // Lua state
}

// NewLuaIntegration creates a new LuaIntegration instance.
func NewLuaIntegration() *LuaIntegration {
	L := lua.NewState()
	return &LuaIntegration{
		L: L,
	}
}

// LoadScript loads and executes a Lua script.
func (li *LuaIntegration) LoadScript(script string) error {
	if err := li.L.DoString(script); err != nil {
		return err
	}
	return nil
}

// CallLuaFunction calls a Lua function and returns the result.
func (li *LuaIntegration) CallLuaFunction(functionName string) (lua.LValue, error) {
	fn := li.L.GetGlobal(functionName)
	if fn.Type() != lua.LTFunction {
		return nil, fmt.Errorf("%s is not a Lua function", functionName)
	}

	if err := li.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}); err != nil {
		return nil, err
	}

	return li.L.Get(-1), nil
}
