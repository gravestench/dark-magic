package modruntime

import (
	"github.com/gravestench/dark-magic/internal/shell"
	lua "github.com/yuin/gopher-lua"
)

// ShellModule exposes live, validated shell presentation settings. Mutation is
// immediate, while persistence remains an explicit user action.
func ShellModule(settings *shell.Settings) Module {
	valuesTable := func(state *lua.LState, values shell.SettingsValues) *lua.LTable {
		result := state.NewTable()
		result.RawSetString("font_size", lua.LNumber(values.FontSize))
		return result
	}
	return Module{Name: "dm.shell/v1", Help: ModuleHelp{
		Summary: "Inspect and edit interactive shell presentation settings.",
		Commands: map[string]CommandHelp{
			"values":   {Summary: "Return the active runtime settings.", Usage: "dm.shell.values()"},
			"defaults": {Summary: "Return the built-in default settings.", Usage: "dm.shell.defaults()"},
			"get":      {Summary: "Read one active setting by name.", Usage: "dm.shell.get(name)"},
			"set": {Summary: "Change a setting immediately without saving it.", Usage: "dm.shell.set(name, value)", Parameters: []ParameterHelp{
				{Name: "name", Type: "string", Description: `Currently "font_size".`},
				{Name: "value", Type: "number", Description: "New setting value."},
			}},
			"reset":  {Summary: "Reset all runtime settings to defaults without saving.", Usage: "dm.shell.reset()"},
			"save":   {Summary: "Persist the active settings for future launches.", Usage: "dm.shell.save()"},
			"status": {Summary: "Return persistence path and unsaved-change state.", Usage: "dm.shell.status()"},
		},
	}, Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"values": func(state *lua.LState) int {
				state.Push(valuesTable(state, settings.Values()))
				return 1
			},
			"defaults": func(state *lua.LState) int {
				state.Push(valuesTable(state, settings.Defaults()))
				return 1
			},
			"get": func(state *lua.LState) int {
				switch name := state.CheckString(1); name {
				case "font_size":
					state.Push(lua.LNumber(settings.Values().FontSize))
					return 1
				default:
					state.RaiseError("unknown shell setting %q", name)
					return 0
				}
			},
			"set": func(state *lua.LState) int {
				name := state.CheckString(1)
				if name != "font_size" {
					state.RaiseError("unknown shell setting %q", name)
					return 0
				}
				if err := settings.SetFontSize(float64(state.CheckNumber(2))); err != nil {
					state.RaiseError("set shell setting %q: %v", name, err)
				}
				return 0
			},
			"reset": func(*lua.LState) int { settings.Reset(); return 0 },
			"save": func(state *lua.LState) int {
				if err := settings.Save(); err != nil {
					state.RaiseError("save shell settings: %v", err)
				}
				return 0
			},
			"status": func(state *lua.LState) int {
				result := state.NewTable()
				result.RawSetString("path", lua.LString(settings.Path()))
				result.RawSetString("dirty", lua.LBool(settings.Dirty()))
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		module.RawSetString("name", lua.LString("dm.shell/v1"))
		state.Push(module)
		return 1
	}}
}
