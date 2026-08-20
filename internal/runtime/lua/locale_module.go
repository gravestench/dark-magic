package modruntime

import (
	"github.com/gravestench/dark-magic/internal/localization"
	lua "github.com/yuin/gopher-lua"
)

// LocaleModule exposes string lookup without exposing locale storage ownership.
func LocaleModule(locale *localization.Locale) Module {
	return Module{
		Name: "engine.locale/v1",
		Help: documentedModule("Resolve localized game strings.", map[string]CommandHelp{
			"text": commandHelp(
				"engine.locale.text(key)",
				"Return localized text for a string key.",
			),
		}),
		Loader: func(state *lua.LState) int {
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
		},
	}
}
