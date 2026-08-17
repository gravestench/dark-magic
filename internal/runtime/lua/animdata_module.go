package modruntime

import (
	"fmt"
	"strings"
	"sync"

	cof "github.com/gravestench/cof"
	cofpkg "github.com/gravestench/cof/pkg"
	lua "github.com/yuin/gopher-lua"
)

const (
	animDataPath       = "data/global/AnimData.d2"
	animDataFrameScale = 256
	animDataTickRate   = 25
)

type animDataReader interface {
	Read(string) ([]byte, error)
}

// AnimDataModule exposes immutable format facts from the session-pinned
// AnimData.d2 generation. It contains no skill, actor, or combat policy.
func AnimDataModule(source animDataReader) Module {
	var once sync.Once
	var catalog *cof.AnimationData
	var loadErr error
	load := func() (*cof.AnimationData, error) {
		once.Do(func() {
			if source == nil {
				loadErr = fmt.Errorf("animdata: authoritative binary source is unavailable")
				return
			}
			var data []byte
			data, loadErr = source.Read(animDataPath)
			if loadErr == nil {
				catalog, loadErr = cof.Load(data)
			}
			if loadErr != nil {
				loadErr = fmt.Errorf("animdata: decode %q: %w", animDataPath, loadErr)
			}
		})
		return catalog, loadErr
	}

	return Module{Name: "engine.animdata/v1", Help: documentedModule("Read session-pinned fixed-tick animation records.", map[string]CommandHelp{
		"record": commandHelp("engine.animdata.record(key)", "Resolve one typed AnimData record and its deterministic tick schedule."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"record": func(state *lua.LState) int {
				asset, err := load()
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				record := asset.GetRecord(strings.ToUpper(state.CheckString(1)))
				if record == nil {
					state.Push(lua.LNil)
					return 1
				}
				frames, speed := record.FramesPerDirection(), record.Speed()
				if frames <= 0 || speed <= 0 {
					state.Push(lua.LNil)
					state.Push(lua.LString(fmt.Sprintf("animdata: record %q has invalid frames or speed", record.Name())))
					return 2
				}
				result := state.NewTable()
				result.RawSetString("name", lua.LString(record.Name()))
				result.RawSetString("frames", lua.LNumber(frames))
				result.RawSetString("speed", lua.LNumber(speed))
				result.RawSetString("frame_scale", lua.LNumber(animDataFrameScale))
				result.RawSetString("ticks_per_second", lua.LNumber(animDataTickRate))
				result.RawSetString("complete_delay", lua.LNumber(tickForCursor(frames*animDataFrameScale, speed)))
				events := state.NewTable()
				for frame := 0; frame < frames; frame++ {
					event := record.Event(frame)
					if event == cofpkg.EventNone {
						continue
					}
					entry := state.NewTable()
					entry.RawSetString("frame", lua.LNumber(frame))
					entry.RawSetString("code", lua.LNumber(event))
					entry.RawSetString("kind", lua.LString(animDataEventName(event)))
					entry.RawSetString("delay", lua.LNumber(tickForCursor(frame*animDataFrameScale, speed)))
					events.Append(entry)
				}
				result.RawSetString("events", events)
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func tickForCursor(cursor, speed int) int {
	if cursor <= 0 {
		return 1
	}
	return (cursor + speed - 1) / speed
}

func animDataEventName(event cof.FrameEvent) string {
	switch event {
	case cofpkg.EventAttack:
		return "attack"
	case cofpkg.EventMissile:
		return "missile"
	case cofpkg.EventSound:
		return "sound"
	case cofpkg.EventSkill:
		return "skill"
	default:
		return "unknown"
	}
}
