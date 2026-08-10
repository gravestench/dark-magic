package modruntime

import (
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/preferences"
	lua "github.com/yuin/gopher-lua"
)

// RenderSettingsTarget accepts renderer preferences without coupling Lua to a
// concrete platform backend.
type RenderSettingsTarget interface {
	SetResidencyDebug(bool)
	SetTextureUploadBudget(uint64)
}

// SettingsModule exposes validated client preferences. Runtime effects are
// applied immediately; persistence remains an explicit UI action.
func SettingsModule(settings *preferences.Settings, mixer *audio.Mixer, renderTargets ...RenderSettingsTarget) Module {
	apply := func(values preferences.Values) error {
		for _, bus := range []string{"sfx", "ambience", "speech"} {
			if err := mixer.SetBusVolume(bus, float32(values.SoundVolume)); err != nil {
				return err
			}
		}
		if err := mixer.SetBusVolume("music", float32(values.MusicVolume)); err != nil {
			return err
		}
		for _, target := range renderTargets {
			target.SetResidencyDebug(values.DebugTextureResidency)
			target.SetTextureUploadBudget(uint64(values.TextureUploadBudgetMB * 1024 * 1024))
		}
		return nil
	}
	_ = apply(settings.Values())
	return Module{Name: "dm.settings/v1", Help: documentedModule("Inspect, preview, and persist client game preferences.", map[string]CommandHelp{
		"get":    commandHelp("dm.settings.get(name)", "Read one active preference."),
		"set":    commandHelp("dm.settings.set(name, value)", "Validate and apply one preference immediately."),
		"save":   commandHelp("dm.settings.save()", "Persist active preferences for future launches."),
		"status": commandHelp("dm.settings.status()", "Return the persistence path and dirty state."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"get": func(state *lua.LState) int {
				values := settings.Values()
				switch name := state.CheckString(1); name {
				case "sound_volume":
					state.Push(lua.LNumber(values.SoundVolume))
				case "music_volume":
					state.Push(lua.LNumber(values.MusicVolume))
				case "debug_texture_residency":
					state.Push(lua.LBool(values.DebugTextureResidency))
				case "texture_upload_budget_mb":
					state.Push(lua.LNumber(values.TextureUploadBudgetMB))
				default:
					state.RaiseError("unknown game preference %q", name)
				}
				return 1
			},
			"set": func(state *lua.LState) int {
				values := settings.Values()
				switch name := state.CheckString(1); name {
				case "sound_volume":
					values.SoundVolume = float64(state.CheckNumber(2))
				case "music_volume":
					values.MusicVolume = float64(state.CheckNumber(2))
				case "debug_texture_residency":
					values.DebugTextureResidency = state.CheckBool(2)
				case "texture_upload_budget_mb":
					values.TextureUploadBudgetMB = float64(state.CheckNumber(2))
				default:
					state.RaiseError("unknown game preference %q", name)
				}
				if err := settings.Update(values); err != nil {
					state.RaiseError("set game preference: %v", err)
				}
				if err := apply(values); err != nil {
					state.RaiseError("apply game preference: %v", err)
				}
				return 0
			},
			"save": func(state *lua.LState) int {
				if err := settings.Save(); err != nil {
					state.RaiseError("save game preferences: %v", err)
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
		state.Push(module)
		return 1
	}}
}
