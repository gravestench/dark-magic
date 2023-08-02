package record_manager

import (
	"fmt"
	"reflect"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ExportToLua exports the foo object to a Lua table.
func (s *Service) ExportToLua(state *lua.LState) {
	for !s.IsLoaded() {
		time.Sleep(time.Second)
	}

	table := state.NewTable()

	// Set the "Sounds" field to the Lua table containing the SoundEntry objects.
	for key, records := range map[string]any{
		"Belts":                  s.Belts,
		"CharStartingAttributes": s.CharStartingAttributes,
		"Inventory":              s.Inventory,
		"Overlays":               s.Overlays,
		"PetTypes":               s.PetTypes,
		"AutoMapEntries":         s.AutoMapEntries,
		"States":                 s.States,
		"Hirelings":              s.Hirelings,
		"HirelingDescriptions":   s.HirelingDescriptions,
		"Missiles":               s.Missiles,
		"DifficultyLevels":       s.DifficultyLevels,
		"Shrines":                s.Shrines,
		"GambleRecords":          s.GambleRecords,
		"NpcTradeRecords":        s.NpcTradeRecords,
		"ExperienceBreakpoints":  s.ExperienceBreakpoints,
		"ItemArmor":              s.ItemArmor,
		"ItemWeapon":             s.ItemWeapon,
		"ItemWeaponClass":        s.ItemWeaponClass,
		"ItemMisc":               s.ItemMisc,
		"ItemTypes":              s.ItemTypes,
		"ItemAutoMagic":          s.ItemAutoMagic,
		"ItemStatCost":           s.ItemStatCost,
		"ItemRatio":              s.ItemRatio,
		"ItemUnique":             s.ItemUnique,
		"ItemHiQualityMods":      s.ItemHiQualityMods,
		"ItemProperties":         s.ItemProperties,
		"CubeRecipes":            s.CubeRecipes,
		"Books":                  s.Books,
		"Gems":                   s.Gems,
		"Runes":                  s.Runes,
		"SetItems":               s.SetItems,
		"SetBonuses":             s.SetBonuses,
		"Skills":                 s.Skills,
		"SkillDesc":              s.SkillDesc,
		"Treasures":              s.Treasures,
		"TreasuresExpansion":     s.TreasuresExpansion,
		"MagicPrefixes":          s.MagicPrefixes,
		"MagicSuffixes":          s.MagicSuffixes,
		"RarePrefixes":           s.RarePrefixes,
		"RareSuffixes":           s.RareSuffixes,
		"UniquePrefixes":         s.UniquePrefixes,
		"UniqueSuffixes":         s.UniqueSuffixes,
		"Objects":                s.Objects,
		"ObjectTypes":            s.ObjectTypes,
		"ObjectGroups":           s.ObjectGroups,
		"ObjectModes":            s.ObjectModes,
		"Sounds":                 s.Sounds,
		"SoundEnvironments":      s.SoundEnvironments,
		"LevelPresets":           s.LevelPresets,
		"LevelType":              s.LevelType,
		"LevelWarp":              s.LevelWarp,
		"LevelDetails":           s.LevelDetails,
		"LevelMaze":              s.LevelMaze,
		"LevelSubstitutions":     s.LevelSubstitutions,
		//"LevelGroups":            s.LevelGroups, // d2r?
		"MonsterUniqueModifiers": s.MonsterUniqueModifiers,
		"MonsterEquipment":       s.MonsterEquipment,
		"MonsterLevelStats":      s.MonsterLevelStats,
		"MonsterPresets":         s.MonsterPresets,
		"MonsterProperties":      s.MonsterProperties,
		"MonsterSequences":       s.MonsterSequences,
		"MonsterStats":           s.MonsterStats,
		"MonsterStats2":          s.MonsterStats2,
		"MonsterSounds":          s.MonsterSounds,
		"MonsterUniqueNames":     s.MonsterUniqueNames,
	} {
		s.logger.Info().Msgf("exporting: %s", key)
		table.RawSetString(key, genericExportArrayToLua(records, state))
	}

	state.SetGlobal("records", table)
}

// genericExportArrayToLua exports an array of structs to a Lua table using "lua" struct tags.
func genericExportArrayToLua(arr interface{}, state *lua.LState) *lua.LTable {
	value := reflect.ValueOf(arr)
	if value.Kind() != reflect.Slice {
		panic("ExportArrayToLua: arr is not a slice")
	}

	table := state.NewTable()

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		if elem.Kind() != reflect.Struct {
			panic("ExportArrayToLua: arr contains non-struct elements")
		}

		// Export the struct to a Lua table and insert it into the array.
		table.RawSetInt(i+1, genericExportToLua(elem.Interface(), state))
	}

	return table
}

// genericExportToLua exports a given object to a Lua table using "lua" struct tags.
func genericExportToLua(obj interface{}, state *lua.LState) *lua.LTable {
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Struct {
		panic("ExportToLua: obj is not a struct")
	}

	table := state.NewTable()
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typ.Field(i)

		luaFieldName := fieldType.Tag.Get("lua")
		if luaFieldName == "" {
			luaFieldName = fieldType.Name
		}

		// Set the fields of the table using the struct's values.
		table.RawSetString(luaFieldName, toLuaValue(field.Interface(), state))
	}

	return table
}

// Helper function to convert Go values to Lua values.
func toLuaValue(value interface{}, state *lua.LState) lua.LValue {
	switch v := value.(type) {
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case fmt.Stringer:
		return lua.LString(v.String())
	case string:
		return lua.LString(v)
	case []string:
		luaTable := state.NewTable()
		for i, str := range v {
			luaTable.RawSetInt(i+1, lua.LString(str))
		}
		return luaTable
	case []any:
		return toLuaValue(v, state)
	default:
		return lua.LString(fmt.Sprintf("%+v", v))
	}
}
