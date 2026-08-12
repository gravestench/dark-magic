package modruntime

import (
	"github.com/gravestench/dark-magic/internal/persistence"
	lua "github.com/yuin/gopher-lua"
)

// SaveModule exposes character metadata and selection without exposing save
// file ownership or mutable Go objects to Lua.
func SaveModule(store *persistence.Store) Module {
	return Module{Name: "engine.save/v1", Help: documentedModule("Create, select, inspect, and delete local characters.", map[string]CommandHelp{
		"create_named": commandHelp("engine.save.create_named(name, class)", "Create and persist a character with an explicit name."),
		"create":       commandHelp("engine.save.create(class)", "Create and persist a character with a generated name."),
		"characters":   commandHelp("engine.save.characters()", "Return all available character summaries."),
		"select":       commandHelp("engine.save.select(id)", "Select the active character by identifier."),
		"delete":       commandHelp("engine.save.delete(id)", "Delete a character by identifier."),
		"selected":     commandHelp("engine.save.selected()", "Return the currently selected character, if any."),
	}), Loader: func(state *lua.LState) int {
		characterTable := func(character persistence.Character) *lua.LTable {
			entry := state.NewTable()
			entry.RawSetString("id", lua.LString(character.ID))
			entry.RawSetString("name", lua.LString(character.Name))
			entry.RawSetString("class", lua.LString(character.Class))
			entry.RawSetString("level", lua.LNumber(character.Level))
			entry.RawSetString("expansion", lua.LBool(character.Expansion))
			entry.RawSetString("hardcore", lua.LBool(character.Hardcore))
			if character.Stats != nil {
				stats := state.NewTable()
				values := map[string]int{
					"experience": character.Stats.Experience, "next_level_experience": character.Stats.NextLevelExperience,
					"strength": character.Stats.Strength, "dexterity": character.Stats.Dexterity,
					"vitality": character.Stats.Vitality, "energy": character.Stats.Energy,
					"defense": character.Stats.Defense, "health": character.Stats.Health,
					"max_health": character.Stats.MaxHealth, "mana": character.Stats.Mana,
					"max_mana": character.Stats.MaxMana, "stamina": character.Stats.Stamina,
					"max_stamina": character.Stats.MaxStamina, "fire_resistance": character.Stats.FireResistance,
					"cold_resistance": character.Stats.ColdResistance, "lightning_resistance": character.Stats.LightningResistance,
					"poison_resistance": character.Stats.PoisonResistance,
				}
				for name, value := range values {
					stats.RawSetString(name, lua.LNumber(value))
				}
				entry.RawSetString("stats", stats)
			}
			if character.Appearance != nil {
				appearance := state.NewTable()
				appearance.RawSetString("cof", lua.LString(character.Appearance.COF))
				appearance.RawSetString("palette", lua.LString(character.Appearance.Palette))
				appearance.RawSetString("direction", lua.LNumber(character.Appearance.Direction))
				components := state.NewTable()
				for component, path := range character.Appearance.Components {
					components.RawSetString(component, lua.LString(path))
				}
				appearance.RawSetString("components", components)
				entry.RawSetString("appearance", appearance)
			}
			return entry
		}
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"create_named": func(state *lua.LState) int {
				character, err := store.CreateNamedWithOptions(
					state.CheckString(1), state.CheckString(2), state.OptBool(3, true), state.OptBool(4, false),
				)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LString(character.ID))
				return 1
			},
			"create": func(state *lua.LState) int {
				err := store.Create(persistence.Character{ID: state.CheckString(1), Name: state.CheckString(2), Class: state.CheckString(3), Level: 1})
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LTrue)
				return 1
			},
			"characters": func(state *lua.LState) int {
				result := state.NewTable()
				for _, character := range store.Characters() {
					result.Append(characterTable(character))
				}
				state.Push(result)
				return 1
			},
			"select": func(state *lua.LState) int {
				if err := store.Select(state.CheckString(1)); err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LTrue)
				return 1
			},
			"delete": func(state *lua.LState) int {
				if err := store.Delete(state.CheckString(1)); err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LTrue)
				return 1
			},
			"selected": func(state *lua.LState) int {
				character, ok := store.Selected()
				if !ok {
					state.Push(lua.LNil)
					return 1
				}
				state.Push(characterTable(character))
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
