package modruntime

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/audiocore"
	lua "github.com/yuin/gopher-lua"
)

const audioSoundType = "dm.audio.sound/v1"

type ownedSound struct {
	mixer *audiocore.Mixer
	id    audiocore.SoundID
	once  sync.Once
	err   error
}

func (s *ownedSound) release() error {
	s.once.Do(func() { s.err = s.mixer.Stop(s.id) })
	return s.err
}

// AudioModule exposes scoped archive-backed sound playback.
func AudioModule(runtime *Runtime, mixer *audiocore.Mixer, source fs.FS) Module {
	catalog := audiocore.NewCatalog(source)
	return Module{Name: "dm.audio/v1", Loader: func(state *lua.LState) int {
		registerSoundType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
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
				options := audiocore.PlayOptions{Bus: "sfx", Volume: 1}
				if table := state.OptTable(2, nil); table != nil {
					if value := table.RawGetString("bus"); value != lua.LNil {
						options.Bus = lua.LVAsString(value)
					}
					if value := table.RawGetString("volume"); value != lua.LNil {
						options.Volume = float32(lua.LVAsNumber(value))
					}
					if value := table.RawGetString("pan"); value != lua.LNil {
						options.Pan = float32(lua.LVAsNumber(value))
					}
					if value := table.RawGetString("loop"); value != lua.LNil {
						options.Loop = lua.LVAsBool(value)
					}
					if value := table.RawGetString("group"); value != lua.LNil {
						options.Group = lua.LVAsString(value)
					}
					if value := table.RawGetString("stream"); value != lua.LNil {
						options.Stream = lua.LVAsBool(value)
					}
				}
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
