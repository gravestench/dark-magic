package modruntime

import (
	"github.com/gravestench/dark-magic/internal/localecore"
	lua "github.com/yuin/gopher-lua"
)

// LocaleModule exposes string lookup without exposing locale storage ownership.
func LocaleModule(locale *localecore.Locale) Module {
	return Module{Name: "dm.locale/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"text": func(state *lua.LState) int {
				value, err := locale.Text(state.CheckString(1))
				state.Push(lua.LString(value))
				if err != nil {
					state.Push(lua.LString(err.Error()))
					return 2
				}
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
