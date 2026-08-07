package modruntime

import (
	"strings"

	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	lua "github.com/yuin/gopher-lua"
)

type gameDataSnapshotter interface {
	Snapshot() (gamedata.Snapshot, error)
}

// GameDataModule exposes stable, typed projections of admitted game-data
// records. Lua receives fresh scalar tables rather than Go-owned records or
// arbitrary columns from the underlying TSV files.
func GameDataModule(catalog gameDataSnapshotter) Module {
	return Module{Name: "dm.game_data/v1", Help: documentedModule("Query typed, normalized Diablo II game data.", map[string]CommandHelp{
		"character_class": commandHelp("dm.game_data.character_class(class)", "Return the typed starting data for a character class."),
		"unique_titles":   commandHelp("dm.game_data.unique_titles()", "Return the available unique-item titles."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"character_class": func(state *lua.LState) int {
				name := state.CheckString(1)
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				record, found := snapshot.CharStatsByClass[name]
				if !found {
					for key, candidate := range snapshot.CharStatsByClass {
						if strings.EqualFold(key, name) {
							record, found = candidate, true
							break
						}
					}
				}
				if !found {
					state.Push(lua.LNil)
					state.Push(lua.LString("unknown character class: " + name))
					return 2
				}
				result := state.NewTable()
				result.RawSetString("class", lua.LString(record.Class))
				result.RawSetString("strength", lua.LNumber(record.Strength))
				result.RawSetString("dexterity", lua.LNumber(record.Dexterity))
				result.RawSetString("intelligence", lua.LNumber(record.Intelligence))
				result.RawSetString("vitality", lua.LNumber(record.Vitality))
				result.RawSetString("stamina", lua.LNumber(record.Stamina))
				state.Push(result)
				return 1
			},
			"unique_titles": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				result := state.NewTable()
				for _, record := range snapshot.UniqueTitles {
					entry := state.NewTable()
					entry.RawSetString("name", lua.LString(record.Name))
					entry.RawSetString("title", lua.LString(record.Namco))
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
