package save

import (
	"github.com/gravestench/dark-magic/internal/persistence"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// Module exposes d2legacy character metadata and selection without exposing
// persistence internals or mutable Go objects to Lua.
func Module(store *persistence.Store) modruntime.Module {
	return modruntime.Module{Name: "d2legacy.save/v1", Help: modruntime.ModuleHelp{Summary: "Create, select, inspect, and delete d2legacy characters.", Commands: map[string]modruntime.CommandHelp{
		"create_named": {Usage: "d2legacy.save.create_named(name, class)", Summary: "Create and persist a character with an explicit name."},
		"create":       {Usage: "d2legacy.save.create(id, name, class)", Summary: "Persist an imported character identity."},
		"characters":   {Usage: "d2legacy.save.characters()", Summary: "Return all available character summaries."},
		"select":       {Usage: "d2legacy.save.select(id)", Summary: "Select the active character by identifier."},
		"delete":       {Usage: "d2legacy.save.delete(id)", Summary: "Delete a character by identifier."},
		"selected":     {Usage: "d2legacy.save.selected()", Summary: "Return the currently selected character, if any."},
	}}, Loader: func(state *lua.LState) int {
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
