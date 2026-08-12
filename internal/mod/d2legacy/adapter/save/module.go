package save

import (
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// Module exposes d2legacy character metadata and selection without exposing
// persistence internals or mutable Go objects to Lua.
func Module(store *Store) modruntime.Module {
	return modruntime.Module{Name: "d2legacy.save_store/v1", Help: modruntime.ModuleHelp{Summary: "Persist already-validated d2legacy character records.", Commands: map[string]modruntime.CommandHelp{
		"create":     {Usage: "d2legacy.save_store.create(record)", Summary: "Persist an already-validated character record."},
		"characters": {Usage: "d2legacy.save.characters()", Summary: "Return all available character summaries."},
		"select":     {Usage: "d2legacy.save.select(id)", Summary: "Select the active character by identifier."},
		"delete":     {Usage: "d2legacy.save.delete(id)", Summary: "Delete a character by identifier."},
		"selected":   {Usage: "d2legacy.save.selected()", Summary: "Return the currently selected character, if any."},
	}}, Loader: func(state *lua.LState) int {
		characterTable := func(character Character) *lua.LTable {
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
			"create": func(state *lua.LState) int {
				record := state.CheckTable(1)
				character := Character{
					ID: string(record.RawGetString("id").(lua.LString)), Name: string(record.RawGetString("name").(lua.LString)),
					Class: string(record.RawGetString("class").(lua.LString)), Level: int(lua.LVAsNumber(record.RawGetString("level"))),
					Expansion: lua.LVAsBool(record.RawGetString("expansion")), Hardcore: lua.LVAsBool(record.RawGetString("hardcore")),
				}
				err := store.Create(character)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LString(character.ID))
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
