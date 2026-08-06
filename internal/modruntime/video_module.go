package modruntime

import (
	"io/fs"
	"sync"

	"github.com/gravestench/dark-magic/internal/video"
	lua "github.com/yuin/gopher-lua"
)

const videoPlaybackType = "dm.video.playback/v1"

type ownedPlayback struct {
	playback video.Playback
	once     sync.Once
	err      error
}

func (p *ownedPlayback) release() error {
	p.once.Do(func() { p.err = p.playback.Stop() })
	return p.err
}

// VideoModule exposes scoped, backend-independent cinematic playback. The
// backend receives the VFS directly so MPQ-sized videos never cross into Lua.
func VideoModule(runtime *Runtime, backend video.Backend, source fs.FS) Module {
	if backend == nil {
		backend = video.Unavailable{}
	}
	return Module{Name: "dm.video/v1", Loader: func(state *lua.LState) int {
		registerPlaybackType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"available": func(state *lua.LState) int {
				state.Push(lua.LBool(backend.Available()))
				return 1
			},
			"play": func(state *lua.LState) int {
				scope, err := runtime.requireActiveScope()
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				name := state.CheckString(1)
				if _, err := fs.Stat(source, name); err != nil {
					state.RaiseError("opening video %q: %v", name, err)
					return 0
				}
				playback, err := backend.Play(source, name)
				if err != nil {
					state.RaiseError("playing video %q: %v", name, err)
					return 0
				}
				owned := &ownedPlayback{playback: playback}
				if err := scope.Add(owned.release); err != nil {
					_ = owned.release()
					state.RaiseError("owning video %q: %v", name, err)
					return 0
				}
				userData := state.NewUserData()
				userData.Value = owned
				state.SetMetatable(userData, state.GetTypeMetatable(videoPlaybackType))
				state.Push(userData)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func registerPlaybackType(state *lua.LState) {
	meta := state.NewTypeMetatable(videoPlaybackType)
	state.SetField(meta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"status": func(state *lua.LState) int {
			snapshot := checkPlayback(state, 1).playback.Snapshot()
			result := state.NewTable()
			result.RawSetString("state", lua.LString(snapshot.State))
			if snapshot.Error != "" {
				result.RawSetString("error", lua.LString(snapshot.Error))
			}
			state.Push(result)
			return 1
		},
		"stop": func(state *lua.LState) int {
			if err := checkPlayback(state, 1).release(); err != nil {
				state.RaiseError("stopping video: %v", err)
			}
			return 0
		},
	}))
}

func checkPlayback(state *lua.LState, index int) *ownedPlayback {
	userData := state.CheckUserData(index)
	playback, ok := userData.Value.(*ownedPlayback)
	if !ok {
		state.ArgError(index, "dm.video/v1 playback expected")
		return nil
	}
	return playback
}
