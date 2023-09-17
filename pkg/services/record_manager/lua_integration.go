package record_manager

import (
	"fmt"
	"reflect"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (s *Service) ExportToLua(state *lua.LState) {
	for !s.IsLoaded() {
		time.Sleep(time.Second)
	}

	table := state.NewTable()

	// for each set of records, create an array. Each element is
	// a lua table for the record, with all fields exported and available.
	for key, records := range map[string]any{
		"Belts":                  s.belts,
		"CharStartingAttributes": s.charStartingAttributes,
		"Inventory":              s.inventory,
		"Overlays":               s.overlays,
		"PetTypes":               s.petTypes,
		"AutoMapEntries":         s.autoMapEntries,
		"States":                 s.states,
		"Hirelings":              s.hirelings,
		"HirelingDescriptions":   s.hirelingDescriptions,
		"Missiles":               s.missiles,
		"DifficultyLevels":       s.difficultyLevels,
		"Shrines":                s.shrines,
		"GambleRecords":          s.gambleRecords,
		"NpcTradeRecords":        s.npcTradeRecords,
		"ExperienceBreakpoints":  s.experienceBreakpoints,
		"ItemArmor":              s.itemArmor,
		"ItemWeapon":             s.itemWeapon,
		"ItemWeaponClass":        s.itemWeaponClass,
		"ItemMisc":               s.itemMisc,
		"ItemTypes":              s.itemTypes,
		"ItemAutoMagic":          s.itemAutoMagic,
		"ItemStatCost":           s.itemStatCost,
		"ItemRatio":              s.itemRatio,
		"ItemUnique":             s.itemUnique,
		"ItemHiQualityMods":      s.itemHiQualityMods,
		"ItemProperties":         s.itemProperties,
		"CubeRecipes":            s.cubeRecipes,
		"Books":                  s.books,
		"Gems":                   s.gems,
		"Runes":                  s.runes,
		"SetItems":               s.setItems,
		"SetBonuses":             s.setBonuses,
		"Skills":                 s.skills,
		"SkillDesc":              s.skillDesc,
		"Treasures":              s.treasures,
		"TreasuresEx":            s.treasuresExpansion,
		"MagicPrefixes":          s.magicPrefixes,
		"MagicSuffixes":          s.magicSuffixes,
		"RarePrefixes":           s.rarePrefixes,
		"RareSuffixes":           s.rareSuffixes,
		"UniquePrefixes":         s.uniquePrefixes,
		"UniqueSuffixes":         s.uniqueSuffixes,
		"Objects":                s.objects,
		"ObjectTypes":            s.objectTypes,
		"ObjectGroups":           s.objectGroups,
		"ObjectModes":            s.objectModes,
		"Sounds":                 s.sounds,
		"SoundEnvironments":      s.soundEnvironments,
		"LevelPresets":           s.levelPresets,
		"LevelType":              s.levelType,
		"LevelWarp":              s.levelWarp,
		"LevelDetails":           s.levelDetails,
		"LevelMaze":              s.levelMaze,
		"LevelSubstitutions":     s.levelSubstitutions,
		//"LevelGroups":            s.levelGroups, // d2r?
		"MonsterUniqueModifiers": s.monsterUniqueModifiers,
		"MonsterEquipment":       s.monsterEquipment,
		"MonsterLevelStats":      s.monsterLevelStats,
		"MonsterPresets":         s.monsterPresets,
		"MonsterProperties":      s.monsterProperties,
		"MonsterSequences":       s.monsterSequences,
		"MonsterStats":           s.monsterStats,
		"MonsterStats2":          s.monsterStats2,
		"MonsterSounds":          s.monsterSounds,
		"MonsterUniqueNames":     s.monsterUniqueNames,
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
	default:
		return lua.LString(fmt.Sprintf("%+v", v))
	}
}
