package modruntime

import (
	"strconv"
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
		"skill":           commandHelp("dm.game_data.skill(id)", "Return typed skill icon, eligibility, and localization metadata."),
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
			"skill": func(state *lua.LState) int {
				id := state.CheckInt(1)
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				skill, found := snapshot.SkillsByID[strconv.Itoa(id)]
				if !found {
					state.Push(lua.LNil)
					return 1
				}
				description, found := snapshot.SkillDescByName[skill.SkillDesc]
				if !found {
					state.Push(lua.LNil)
					return 1
				}
				icon, err := strconv.Atoi(strings.TrimSpace(description.IconCel))
				if err != nil || icon < 0 {
					state.Push(lua.LNil)
					return 1
				}
				sheets := map[string]string{
					"": "Skillicon", "ama": "AmSkillicon", "sor": "SoSkillicon", "nec": "NeSkillicon",
					"pal": "PaSkillicon", "bar": "BaSkillicon", "dru": "DrSkillicon", "ass": "AsSkillicon",
				}
				sheet, found := sheets[strings.ToLower(strings.TrimSpace(skill.CharClass))]
				if !found {
					sheet = "Skillicon"
				}
				result := state.NewTable()
				result.RawSetString("id", lua.LNumber(id))
				result.RawSetString("icon", lua.LNumber(icon))
				result.RawSetString("sheet", lua.LString("data/global/ui/SPELLS/"+sheet+".DC6"))
				result.RawSetString("name_key", lua.LString(description.StrName))
				result.RawSetString("short_key", lua.LString(description.StrShort))
				result.RawSetString("list_row", lua.LNumber(description.ListRow))
				result.RawSetString("left_allowed", lua.LBool(strings.TrimSpace(skill.LeftSkill) == "1"))
				result.RawSetString("passive", lua.LBool(strings.TrimSpace(skill.Passive) == "1"))
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
