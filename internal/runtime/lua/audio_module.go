package modruntime

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/audio"
	lua "github.com/yuin/gopher-lua"
)

const audioSoundType = "dm.audio.sound/v1"

type ownedSound struct {
	mixer *audio.Mixer
	id    audio.SoundID
	once  sync.Once
	err   error
}

func luaPlayOptions(state *lua.LState, index int) audio.PlayOptions {
	options := audio.PlayOptions{Bus: "sfx", Volume: 1}
	if table := state.OptTable(index, nil); table != nil {
		for name, apply := range map[string]func(lua.LValue){
			"bus":    func(value lua.LValue) { options.Bus = lua.LVAsString(value) },
			"volume": func(value lua.LValue) { options.Volume = float32(lua.LVAsNumber(value)) },
			"pan":    func(value lua.LValue) { options.Pan = float32(lua.LVAsNumber(value)) },
			"loop":   func(value lua.LValue) { options.Loop = lua.LVAsBool(value) },
			"group":  func(value lua.LValue) { options.Group = lua.LVAsString(value) },
			"stream": func(value lua.LValue) { options.Stream = lua.LVAsBool(value) },
		} {
			if value := table.RawGetString(name); value != lua.LNil {
				apply(value)
			}
		}
	}
	return options
}

func (s *ownedSound) release() error {
	s.once.Do(func() { s.err = s.mixer.Stop(s.id) })
	return s.err
}

// AudioModule exposes scoped archive-backed sound playback.
func AudioModule(runtime *Runtime, mixer *audio.Mixer, source fs.FS, records audio.SoundRecords) Module {
	catalog := audio.NewCatalog(source, records)
	return Module{Name: "dm.audio/v1", Loader: func(state *lua.LState) int {
		registerSoundType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"diagnostics": func(state *lua.LState) int {
				diagnostics := mixer.Diagnostics()
				result := state.NewTable()
				result.RawSetString("active", lua.LNumber(diagnostics.Active))
				result.RawSetString("pending", lua.LNumber(diagnostics.Pending))
				result.RawSetString("slots", lua.LNumber(diagnostics.Slots))
				buses := state.NewTable()
				for bus, volume := range diagnostics.BusVolumes {
					buses.RawSetString(bus, lua.LNumber(volume))
				}
				result.RawSetString("buses", buses)
				state.Push(result)
				return 1
			},
			"exists": func(state *lua.LState) int {
				_, err := fs.Stat(source, state.CheckString(1))
				state.Push(lua.LBool(err == nil))
				return 1
			},
			"play": func(state *lua.LState) int {
				scope, err := runtime.requireActiveScope()
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				fileName := state.CheckString(1)
				data, err := fs.ReadFile(source, fileName)
				if err != nil {
					state.RaiseError("reading sound %q: %v", fileName, err)
					return 0
				}
				format := strings.ToLower(filepath.Ext(fileName))
				options := luaPlayOptions(state, 2)
				id, err := mixer.PlayWithOptions(format, data, options)
				if err != nil {
					state.RaiseError("playing sound %q: %v", fileName, err)
					return 0
				}
				sound := &ownedSound{mixer: mixer, id: id}
				if err := scope.Add(sound.release); err != nil {
					_ = sound.release()
					state.RaiseError("owning sound %q: %v", fileName, err)
					return 0
				}
				userData := state.NewUserData()
				userData.Value = sound
				state.SetMetatable(userData, state.GetTypeMetatable(audioSoundType))
				state.Push(userData)
				return 1
			},
			"play_persistent": func(state *lua.LState) int {
				fileName := state.CheckString(1)
				options := luaPlayOptions(state, 2)
				if options.Group == "" {
					state.ArgError(2, "persistent audio requires a group")
					return 0
				}
				data, err := fs.ReadFile(source, fileName)
				if err != nil {
					state.RaiseError("reading persistent sound %q: %v", fileName, err)
					return 0
				}
				if err := mixer.StopGroup(options.Group); err != nil {
					state.RaiseError("replacing persistent audio group %q: %v", options.Group, err)
					return 0
				}
				if _, err := mixer.PlayWithOptions(strings.ToLower(filepath.Ext(fileName)), data, options); err != nil {
					state.RaiseError("playing persistent sound %q: %v", fileName, err)
				}
				return 0
			},
			"set_bus_volume": func(state *lua.LState) int {
				if err := mixer.SetBusVolume(state.CheckString(1), float32(state.CheckNumber(2))); err != nil {
					state.RaiseError("setting bus volume: %v", err)
				}
				return 0
			},
			"play_record": func(state *lua.LState) int {
				scope, err := runtime.requireActiveScope()
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				definition, err := catalog.Resolve(state.CheckString(1), uint64(state.OptInt64(2, 0)))
				if err != nil {
					state.RaiseError("resolving sound record: %v", err)
					return 0
				}
				id, err := mixer.PlayWithOptions(definition.Format, definition.Data, definition.Options)
				if err != nil {
					state.RaiseError("playing sound record %q: %v", definition.Name, err)
					return 0
				}
				sound := &ownedSound{mixer: mixer, id: id}
				if err := scope.Add(sound.release); err != nil {
					_ = sound.release()
					state.RaiseError("owning sound record %q: %v", definition.Name, err)
					return 0
				}
				userData := state.NewUserData()
				userData.Value = sound
				state.SetMetatable(userData, state.GetTypeMetatable(audioSoundType))
				state.Push(userData)
				return 1
			},
			"stop_group": func(state *lua.LState) int {
				if err := mixer.StopGroup(state.CheckString(1)); err != nil {
					state.RaiseError("stopping audio group: %v", err)
				}
				return 0
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func registerSoundType(state *lua.LState) {
	meta := state.NewTypeMetatable(audioSoundType)
	state.SetField(meta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"set_volume": func(state *lua.LState) int {
			sound := checkSound(state, 1)
			if err := sound.mixer.SetVolume(sound.id, float32(state.CheckNumber(2))); err != nil {
				state.RaiseError("setting sound volume: %v", err)
			}
			return 0
		},
		"set_pan": func(state *lua.LState) int {
			sound := checkSound(state, 1)
			if err := sound.mixer.SetPan(sound.id, float32(state.CheckNumber(2))); err != nil {
				state.RaiseError("setting sound pan: %v", err)
			}
			return 0
		},
		"fade_to": func(state *lua.LState) int {
			sound := checkSound(state, 1)
			if err := sound.mixer.Fade(sound.id, float32(state.CheckNumber(2)), time.Duration(state.CheckNumber(3))*time.Millisecond); err != nil {
				state.RaiseError("fading sound: %v", err)
			}
			return 0
		},
		"stop": func(state *lua.LState) int {
			if err := checkSound(state, 1).release(); err != nil {
				state.RaiseError("stopping sound: %v", err)
			}
			return 0
		},
	}))
}

func checkSound(state *lua.LState, index int) *ownedSound {
	userData := state.CheckUserData(index)
	sound, ok := userData.Value.(*ownedSound)
	if !ok {
		state.ArgError(index, "dm.audio/v1 sound expected")
		return nil
	}
	return sound
}
