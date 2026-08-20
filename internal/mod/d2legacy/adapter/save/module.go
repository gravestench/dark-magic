package save

import (
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

const saveStoreModuleName = "d2legacy.save_store/v1"

// luaStoreBridge captures one roster for Lua callbacks without exposing mutable Go values to scripts.
type luaStoreBridge struct {
	store *Store
}

// Module exposes validated character persistence and selection while keeping storage ownership in Go.
func Module(store *Store) modruntime.Module {
	bridge := luaStoreBridge{store: store}

	return modruntime.Module{
		Name:   saveStoreModuleName,
		Help:   saveStoreHelp(),
		Loader: bridge.load,
	}
}

// saveStoreHelp keeps the public Lua command contract readable independently from callback implementation.
func saveStoreHelp() modruntime.ModuleHelp {
	return modruntime.ModuleHelp{
		Summary: "Persist already-validated d2legacy character records.",
		Commands: map[string]modruntime.CommandHelp{
			"create": {
				Usage:   "d2legacy.save_store.create(record)",
				Summary: "Persist an already-validated character record.",
			},
			"create_selected": {
				Usage:   "d2legacy.save_store.create_selected(record)",
				Summary: "Atomically persist and select an already-validated character record.",
			},
			"characters": {
				Usage:   "d2legacy.save.characters()",
				Summary: "Return all available character summaries.",
			},
			"delete": {
				Usage:   "d2legacy.save.delete(id)",
				Summary: "Delete a character by identifier.",
			},
			"select": {
				Usage:   "d2legacy.save.select(id)",
				Summary: "Select the active character by identifier.",
			},
			"selected": {
				Usage:   "d2legacy.save.selected()",
				Summary: "Return the currently selected character, if any.",
			},
		},
	}
}

// load registers stable callback names and API version on the table returned to Lua require callers.
func (bridge luaStoreBridge) load(state *lua.LState) int {
	module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"create":          bridge.create,
		"create_selected": bridge.createSelected,
		"characters":      bridge.characters,
		"delete":          bridge.delete,
		"select":          bridge.selectCharacter,
		"selected":        bridge.selected,
	})
	module.RawSetString("api", lua.LNumber(1))
	state.Push(module)

	return 1
}

// create validates the Lua record through mod policy before crossing the policy-free persistence boundary.
func (bridge luaStoreBridge) create(state *lua.LState) int {
	character, err := characterFromTable(state.CheckTable(1))
	if err != nil {
		return pushLuaStoreError(state, err)
	}

	if err := bridge.store.Create(character); err != nil {
		return pushLuaStoreError(state, err)
	}

	state.Push(lua.LString(character.ID))

	return 1
}

// createSelected preserves atomic create-and-select behavior so an old selection cannot survive character creation.
func (bridge luaStoreBridge) createSelected(state *lua.LState) int {
	character, err := characterFromTable(state.CheckTable(1))
	if err != nil {
		return pushLuaStoreError(state, err)
	}

	if err := bridge.store.CreateSelected(character); err != nil {
		return pushLuaStoreError(state, err)
	}

	state.Push(lua.LString(character.ID))

	return 1
}

// characters returns defensive roster snapshots, preventing Lua table mutation from changing stored profiles.
func (bridge luaStoreBridge) characters(state *lua.LState) int {
	result := state.NewTable()
	for _, character := range bridge.store.Characters() {
		result.Append(characterTable(state, character))
	}

	state.Push(result)

	return 1
}

// selectCharacter changes selection by opaque identity and retains the established nil-plus-error failure result.
func (bridge luaStoreBridge) selectCharacter(state *lua.LState) int {
	if err := bridge.store.Select(state.CheckString(1)); err != nil {
		return pushLuaStoreError(state, err)
	}

	state.Push(lua.LTrue)

	return 1
}

// delete removes by opaque identity and retains the established nil-plus-error failure result.
func (bridge luaStoreBridge) delete(state *lua.LState) int {
	if err := bridge.store.Delete(state.CheckString(1)); err != nil {
		return pushLuaStoreError(state, err)
	}

	state.Push(lua.LTrue)

	return 1
}

// selected returns nil when no selection exists instead of fabricating an empty character table.
func (bridge luaStoreBridge) selected(state *lua.LState) int {
	character, found := bridge.store.Selected()
	if !found {
		state.Push(lua.LNil)
		return 1
	}

	state.Push(characterTable(state, character))

	return 1
}

// pushLuaStoreError preserves the two-value nil/error convention used by fallible save-store commands.
func pushLuaStoreError(state *lua.LState, err error) int {
	state.Push(lua.LNil)
	state.Push(lua.LString(err.Error()))

	return 2
}

// characterTable converts a defensive character snapshot into the stable table schema consumed by presentation Lua.
func characterTable(state *lua.LState, character Character) *lua.LTable {
	entry := state.NewTable()
	entry.RawSetString("id", lua.LString(character.ID))
	entry.RawSetString("name", lua.LString(character.Name))
	entry.RawSetString("class", lua.LString(character.Class))
	entry.RawSetString("level", lua.LNumber(character.Level))
	entry.RawSetString("expansion", lua.LBool(character.Expansion))
	entry.RawSetString("hardcore", lua.LBool(character.Hardcore))

	if character.Stats != nil {
		entry.RawSetString("stats", characterStatsTable(state, *character.Stats))
	}

	if character.Appearance != nil {
		entry.RawSetString("appearance", characterAppearanceTable(state, *character.Appearance))
	}

	return entry
}

// characterStatsTable preserves the complete snake-case Lua schema while copying only scalar snapshot values.
func characterStatsTable(state *lua.LState, stats Stats) *lua.LTable {
	table := state.NewTable()

	values := map[string]int{
		"experience":            stats.Experience,
		"next_level_experience": stats.NextLevelExperience,
		"strength":              stats.Strength,
		"dexterity":             stats.Dexterity,
		"vitality":              stats.Vitality,
		"energy":                stats.Energy,
		"defense":               stats.Defense,
		"health":                stats.Health,
		"max_health":            stats.MaxHealth,
		"mana":                  stats.Mana,
		"max_mana":              stats.MaxMana,
		"stamina":               stats.Stamina,
		"max_stamina":           stats.MaxStamina,
		"fire_resistance":       stats.FireResistance,
		"cold_resistance":       stats.ColdResistance,
		"lightning_resistance":  stats.LightningResistance,
		"poison_resistance":     stats.PoisonResistance,
	}
	for name, value := range values {
		table.RawSetString(name, lua.LNumber(value))
	}

	return table
}

// characterAppearanceTable copies component paths so Lua cannot retain a reference to the Go-owned component map.
func characterAppearanceTable(state *lua.LState, appearance Appearance) *lua.LTable {
	table := state.NewTable()
	table.RawSetString("cof", lua.LString(appearance.COF))
	table.RawSetString("palette", lua.LString(appearance.Palette))
	table.RawSetString("direction", lua.LNumber(appearance.Direction))

	components := state.NewTable()
	for component, path := range appearance.Components {
		components.RawSetString(component, lua.LString(path))
	}

	table.RawSetString("components", components)

	return table
}

// characterFromTable admits only creation choices; NewCharacter supplies progression and canonical class defaults.
func characterFromTable(record *lua.LTable) (Character, error) {
	return NewCharacter(CharacterRequest{
		ID:        string(lua.LVAsString(record.RawGetString("id"))),
		Name:      string(lua.LVAsString(record.RawGetString("name"))),
		Class:     string(lua.LVAsString(record.RawGetString("class"))),
		Expansion: lua.LVAsBool(record.RawGetString("expansion")),
		Hardcore:  lua.LVAsBool(record.RawGetString("hardcore")),
	})
}
