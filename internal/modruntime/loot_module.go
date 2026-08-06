package modruntime

import (
	"bytes"
	"io/fs"

	"github.com/gravestench/dark-magic/pkg/loot"
	lua "github.com/yuin/gopher-lua"
)

// LootModule exposes deterministic treasure-class rolling over layered content.
func LootModule(source fs.FS) Module {
	return Module{Name: "dm.loot/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"event_seed": func(state *lua.LState) int {
				seed, err := loot.EventSeed(uint64(state.CheckNumber(1)), loot.Event{Kind: loot.EventKind(state.CheckString(2)), EntityID: uint64(state.CheckNumber(3)), Sequence: uint64(state.CheckNumber(4))})
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LNumber(seed))
				return 1
			},
			"roll_tsv": func(state *lua.LState) int {
				fileName := state.CheckString(1)
				className := state.CheckString(2)
				seed := uint64(state.CheckNumber(3))
				data, err := fs.ReadFile(source, fileName)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				catalog, err := loot.ParseTreasureClassTSV(bytes.NewReader(data))
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				drops, err := loot.New(catalog, seed).Roll(className)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				result := state.NewTable()
				for _, drop := range drops {
					entry := state.NewTable()
					entry.RawSetString("code", lua.LString(drop.Code))
					path := state.NewTable()
					for _, segment := range drop.Path {
						path.Append(lua.LString(segment))
					}
					entry.RawSetString("path", path)
					quality := state.NewTable()
					for _, modifier := range drop.Quality {
						value := state.NewTable()
						value.RawSetString("unique", lua.LNumber(modifier.Unique))
						value.RawSetString("set", lua.LNumber(modifier.Set))
						value.RawSetString("rare", lua.LNumber(modifier.Rare))
						value.RawSetString("magic", lua.LNumber(modifier.Magic))
						quality.Append(value)
					}
					entry.RawSetString("quality", quality)
					result.Append(entry)
				}
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
