package modruntime

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// ScriptError preserves the source, lifecycle phase, and Lua traceback for
// diagnostics returned across the runtime boundary.
type ScriptError struct {
	Source    string
	Phase     string
	Traceback string
	Err       error
}

func (e *ScriptError) Error() string {
	if e.Traceback != "" {
		return fmt.Sprintf("%s %q: %v\n%s", e.Phase, e.Source, e.Err, e.Traceback)
	}
	return fmt.Sprintf("%s %q: %v", e.Phase, e.Source, e.Err)
}

func (e *ScriptError) Unwrap() error { return e.Err }

func scriptError(source, phase string, err error) error {
	if err == nil {
		return nil
	}
	var apiError *lua.ApiError
	traceback := ""
	if errors.As(err, &apiError) {
		traceback = apiError.StackTrace
	}
	return &ScriptError{Source: source, Phase: phase, Traceback: traceback, Err: err}
}

func isolatedEnvironment(state *lua.LState) *lua.LTable {
	environment := state.NewTable()
	environment.RawSetString("_G", environment)
	metatable := state.NewTable()
	metatable.RawSetString("__index", state.Get(lua.GlobalsIndex))
	state.SetMetatable(environment, metatable)
	return environment
}
