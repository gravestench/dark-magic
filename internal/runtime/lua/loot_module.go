package modruntime

import (
	"github.com/gravestench/dark-magic/internal/game/loot"
	lua "github.com/yuin/gopher-lua"
)

// LootModule exposes deterministic rolling over typed treasure-class records.
func LootModule(records loot.TreasureClassRecords) Module {
	return Module{Name: "engine.loot/v1", Help: documentedModule("Perform deterministic loot-table rolls from game data.", map[string]CommandHelp{
		"event_seed": commandHelp("engine.loot.event_seed(...) ", "Derive a deterministic seed for a loot event."),
		"roll":       commandHelp("engine.loot.roll(table, seed)", "Roll a treasure class deterministically."),
	}), Loader: func(state *lua.LState) int {
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
			"roll": func(state *lua.LState) int {
				className := state.CheckString(1)
				seed := uint64(state.CheckNumber(2))
				catalog, err := loot.CatalogFromRecords(records)
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
