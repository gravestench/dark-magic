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
		result.RawSetString("console_height", lua.LNumber(values.ConsoleHeight))
		result.RawSetString("opacity", lua.LNumber(values.Opacity))
		result.RawSetString("transcript_limit", lua.LNumber(values.TranscriptLimit))
		result.RawSetString("animation_speed", lua.LNumber(values.AnimationSpeed))
		return result
	}
	setValue := func(state *lua.LState, values *shell.SettingsValues, name string, index int) {
		switch name {
		case "font_size":
			values.FontSize = float64(state.CheckNumber(index))
		case "console_height":
			values.ConsoleHeight = float64(state.CheckNumber(index))
		case "opacity":
			values.Opacity = float64(state.CheckNumber(index))
		case "transcript_limit":
			values.TranscriptLimit = state.CheckInt(index)
		case "animation_speed":
			values.AnimationSpeed = float64(state.CheckNumber(index))
		default:
			state.RaiseError("unknown shell setting %q", name)
		}
	}
	return Module{Name: "dm.shell/v1", Help: ModuleHelp{
		Summary: "Inspect and edit interactive shell presentation settings.",
		Commands: map[string]CommandHelp{
			"values":   {Summary: "Return the active runtime settings.", Usage: "dm.shell.values()"},
			"defaults": {Summary: "Return the built-in default settings.", Usage: "dm.shell.defaults()"},
			"get":      {Summary: "Read one active setting by name.", Usage: "dm.shell.get(name)"},
			"set": {Summary: "Change a setting immediately without saving it.", Usage: "dm.shell.set(name, value)", Parameters: []ParameterHelp{
				{Name: "name", Type: "string", Description: "A key returned by dm.shell.values()."},
				{Name: "value", Type: "number", Description: "New setting value."},
			}},
			"set_many": {Summary: "Validate and apply several settings atomically.", Usage: "dm.shell.set_many(values)"},
			"reset":    {Summary: "Reset all runtime settings to defaults without saving.", Usage: "dm.shell.reset()"},
			"reload":   {Summary: "Discard unsaved changes and reload persistent settings.", Usage: "dm.shell.reload()"},
			"save":     {Summary: "Persist the active settings for future launches.", Usage: "dm.shell.save()"},
			"status":   {Summary: "Return persistence path and unsaved-change state.", Usage: "dm.shell.status()"},
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
				case "console_height":
					state.Push(lua.LNumber(settings.Values().ConsoleHeight))
				case "opacity":
					state.Push(lua.LNumber(settings.Values().Opacity))
				case "transcript_limit":
					state.Push(lua.LNumber(settings.Values().TranscriptLimit))
				case "animation_speed":
					state.Push(lua.LNumber(settings.Values().AnimationSpeed))
				default:
					state.RaiseError("unknown shell setting %q", name)
					return 0
				}
				return 1
			},
			"set": func(state *lua.LState) int {
				values := settings.Values()
				setValue(state, &values, state.CheckString(1), 2)
				if err := settings.Update(values); err != nil {
					state.RaiseError("set shell setting: %v", err)
				}
				return 0
			},
			"set_many": func(state *lua.LState) int {
				values := settings.Values()
				state.CheckTable(1).ForEach(func(key, value lua.LValue) {
					state.Push(value)
					setValue(state, &values, key.String(), state.GetTop())
					state.Pop(1)
				})
				if err := settings.Update(values); err != nil {
					state.RaiseError("set shell settings: %v", err)
				}
				return 0
			},
			"reset": func(*lua.LState) int { settings.Reset(); return 0 },
			"reload": func(state *lua.LState) int {
				if err := settings.Reload(); err != nil {
					state.RaiseError("reload shell settings: %v", err)
				}
				return 0
			},
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
