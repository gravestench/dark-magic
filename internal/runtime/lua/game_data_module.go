package modruntime

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	gamemissile "github.com/gravestench/dark-magic/internal/game/missile"
	gamemonster "github.com/gravestench/dark-magic/internal/game/monster"
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
		"monsters":        commandHelp("dm.game_data.monsters()", "Return coherent ordinary-monster presentation recipes for development inspection."),
		"missiles":        commandHelp("dm.game_data.missiles()", "Return coherent missile presentation recipes for development inspection."),
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
			"monsters": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				ids := make([]string, 0, len(snapshot.MonstersByID))
				for id := range snapshot.MonstersByID {
					ids = append(ids, id)
				}
				sort.Strings(ids)
				result := state.NewTable()
				for _, id := range ids {
					definition, err := gamemonster.FromCatalog(snapshot, id, gamemonster.Normal)
					if err != nil || len(definition.Components) == 0 {
						continue
					}
					entry := state.NewTable()
					entry.RawSetString("id", lua.LString(definition.ID))
					entry.RawSetString("token", lua.LString(definition.Token))
					entry.RawSetString("weapon_class", lua.LString(definition.WeaponClass))
					entry.RawSetString("name_key", lua.LString(definition.NameKey))
					keys := make([]string, 0, len(definition.Components))
					for key := range definition.Components {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					parts := make([]string, 0, len(keys))
					for _, key := range keys {
						parts = append(parts, key+"="+definition.Components[key])
					}
					entry.RawSetString("components", lua.LString(strings.Join(parts, ",")))
					result.Append(entry)
				}
				state.Push(result)
				return 1
			},
			"missiles": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				ids := make([]string, 0, len(snapshot.MissilesByName))
				for id := range snapshot.MissilesByName {
					ids = append(ids, id)
				}
				sort.Strings(ids)
				result := state.NewTable()
				for _, id := range ids {
					recipe, err := gamemissile.PresentationFromCatalog(snapshot, id)
					if err != nil {
						continue
					}
					entry := state.NewTable()
					entry.RawSetString("id", lua.LString(recipe.MissileID))
					entry.RawSetString("dcc", lua.LString(recipe.DCC))
					entry.RawSetString("palette", lua.LString(recipe.Palette))
					entry.RawSetString("directions", lua.LNumber(recipe.Directions))
					entry.RawSetString("frames_per_second", lua.LNumber(recipe.FramesPerSecond))
					entry.RawSetString("loop", lua.LBool(recipe.Loop))
					entry.RawSetString("travel_sound", lua.LString(recipe.TravelSound))
					entry.RawSetString("hit_sound", lua.LString(recipe.HitSound))
					entry.RawSetString("offset_x", lua.LNumber(recipe.OffsetX))
					entry.RawSetString("offset_y", lua.LNumber(recipe.OffsetY))
					entry.RawSetString("offset_z", lua.LNumber(recipe.OffsetZ))
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
