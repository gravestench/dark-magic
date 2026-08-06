package gamedata

import (
	"os"
	"slices"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/recordstore"
	"github.com/gravestench/dark-magic/pkg/models"
)

func TestRealArchivesDecodeTypedCoreTables(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	store := recordstore.New(assets)
	characters, err := Load[models.CharStats](store, CharStatsTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(characters) < 7 {
		t.Fatalf("typed charstats rows = %d, want at least seven classes", len(characters))
	}
	byClass, err := Index(characters, func(record models.CharStats) string { return record.Class })
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := byClass["Amazon"]; !exists {
		t.Fatal("typed charstats index is missing Amazon")
	}
	sounds, err := Load[models.SoundEntry](store, SoundsTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(sounds) == 0 {
		t.Fatal("typed sounds table is empty")
	}
	catalog := New(store)
	snapshot, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.LevelsByID) == 0 || len(snapshot.ObjectsByClass) == 0 || len(snapshot.SkillsByID) == 0 || len(snapshot.TreasureByName) == 0 {
		t.Fatal("typed core catalog indexes are incomplete")
	}
	if len(snapshot.ArmorByCode) == 0 || len(snapshot.WeaponsByCode) == 0 || len(snapshot.MiscByCode) == 0 || len(snapshot.ItemTypesByCode) == 0 {
		t.Fatal("typed base-item catalog indexes are incomplete")
	}
	if len(snapshot.ItemRatios) == 0 || len(snapshot.ItemStatsByName) == 0 || len(snapshot.PropertiesByCode) == 0 || len(snapshot.UniqueByIndex) == 0 || len(snapshot.SetItemsByIndex) == 0 {
		t.Fatal("typed item-rule catalog indexes are incomplete")
	}
	if len(snapshot.MagicPrefixes) == 0 || len(snapshot.MagicSuffixes) == 0 || len(snapshot.AutoMagic) == 0 || len(snapshot.RarePrefixes) == 0 || len(snapshot.RareSuffixes) == 0 {
		t.Fatal("typed affix tables are incomplete")
	}
	if len(snapshot.GemsByCode) == 0 || len(snapshot.RuneWords) == 0 || len(snapshot.CubeRecipes) == 0 || len(snapshot.SetsByIndex) == 0 {
		t.Fatal("typed socketing and crafting tables are incomplete")
	}
	if len(snapshot.LevelTypes) == 0 || len(snapshot.LevelPresetByDef) == 0 || len(snapshot.LevelMazeByLevel) == 0 || len(snapshot.LevelWarps) == 0 || len(snapshot.LevelSubs) == 0 {
		t.Fatal("typed world-generation tables are incomplete")
	}
	if !slices.ContainsFunc(snapshot.LevelTypes, func(record models.LevelType) bool { return record.Act > 0 }) ||
		!slices.ContainsFunc(snapshot.LevelPresets, func(record models.LevelPreset) bool { return record.Def > 0 && record.Files > 0 }) ||
		!slices.ContainsFunc(snapshot.LevelMazes, func(record models.LevelMazeData) bool { return record.Rooms > 0 && record.SizeX > 0 }) ||
		!slices.ContainsFunc(snapshot.LevelWarps, func(record models.LevelWarp) bool { return record.Id != "" && record.Tiles > 0 }) ||
		!slices.ContainsFunc(snapshot.LevelSubs, func(record models.LevelSubstitutionData) bool { return record.Type > 0 && record.File != "" }) {
		t.Fatal("typed world-generation fields did not bind representative authored values")
	}
	if len(snapshot.MonstersByID) == 0 || len(snapshot.MonsterGfxByID) == 0 || len(snapshot.MonsterLevels) == 0 || len(snapshot.MonsterPropsByID) == 0 || len(snapshot.MonsterSoundByID) == 0 || len(snapshot.MonsterEquipment) == 0 {
		t.Fatal("typed monster foundation tables are incomplete")
	}
	if len(snapshot.MissilesByName) == 0 || len(snapshot.StatesByName) == 0 || len(snapshot.OverlaysByName) == 0 || len(snapshot.PetTypes) == 0 {
		t.Fatal("typed combat presentation tables are incomplete")
	}
	if !slices.ContainsFunc(snapshot.PetTypes, func(record models.PetType) bool { return record.MClass != [4]int{} || record.MIcon != [4]string{} }) {
		t.Fatal("typed pet grouped fields did not bind authored values")
	}
	if len(snapshot.Experience) == 0 || len(snapshot.InventoryByClass) == 0 || len(snapshot.BeltsByName) == 0 || len(snapshot.Hirelings) == 0 || len(snapshot.DifficultyByName) == 0 {
		t.Fatal("typed character configuration tables are incomplete")
	}
	if len(snapshot.SkillDescByName) == 0 || len(snapshot.SoundEnvByHandle) == 0 || len(snapshot.AutoMapEntries) == 0 {
		t.Fatal("typed presentation support tables are incomplete")
	}
	if len(snapshot.NPCTradesByID) == 0 || len(snapshot.ShrinesByType) == 0 || len(snapshot.MonsterPresets) == 0 || len(snapshot.GambleItemsByCode) == 0 {
		t.Fatal("typed world-interaction tables are incomplete")
	}
	if len(snapshot.ObjectTypesByName) == 0 || len(snapshot.ObjectGroupsByName) == 0 || len(snapshot.ObjectModesByName) == 0 {
		t.Fatal("typed object metadata tables are incomplete")
	}
	if len(snapshot.QualityModifiers) == 0 || len(snapshot.WeaponClassByCode) == 0 || len(snapshot.BooksByName) == 0 {
		t.Fatal("typed item support tables are incomplete")
	}
	if len(snapshot.MonsterSequences) == 0 || len(snapshot.MonsterUniqueByID) == 0 || len(snapshot.UniqueAppellations) == 0 || len(snapshot.UniquePrefixes) == 0 || len(snapshot.UniqueSuffixes) == 0 {
		t.Fatal("typed monster identity tables are incomplete")
	}
	if len(snapshot.ClassicTreasureByName) == 0 || len(snapshot.HirelingDescriptionByID) == 0 {
		t.Fatal("typed classic and hireling support tables are incomplete")
	}
	if len(snapshot.SuperUniquesByID) == 0 || len(snapshot.SuperUniquesByHardcodedID) == 0 {
		t.Fatal("typed super-unique table is incomplete")
	}
	if !slices.ContainsFunc(snapshot.SuperUniques, func(record models.SuperUnique) bool {
		return record.Class != "" && record.MaxGroup >= record.MinGroup && record.TreasureClassNormal != ""
	}) {
		t.Fatal("typed super-unique guide fields did not bind representative authored values")
	}
	if len(snapshot.LowQualityItemNames) == 0 || len(snapshot.BodyLocationsByCode) == 0 || len(snapshot.StorePagesByCode) == 0 || len(snapshot.CompositeComponentsByToken) == 0 || len(snapshot.HitClassesByCode) == 0 {
		t.Fatal("typed item lookup tables are incomplete")
	}
	if len(snapshot.PlayerClassesByCode) == 0 || len(snapshot.PlayerModesByToken) == 0 || len(snapshot.PlayerTypesByToken) == 0 || len(snapshot.MonsterModesByToken) == 0 || len(snapshot.MonsterPlaces) == 0 {
		t.Fatal("typed actor lookup tables are incomplete")
	}
	if len(snapshot.TransformColorsByCode) == 0 || len(snapshot.ComponentCodesByCode) == 0 || len(snapshot.ElementTypesByCode) == 0 || len(snapshot.EventsByName) == 0 || len(snapshot.MissileCalculationsByCode) == 0 || len(snapshot.SkillCalculationsByCode) == 0 {
		t.Fatal("typed calculation lookup tables are incomplete")
	}
	if len(snapshot.Issues) == 0 {
		t.Fatal("expected shipped-data diagnostics for known duplicate/sentinel records")
	}
	for name, load := range map[string]func() (int, error){
		"levels": func() (int, error) {
			records, err := Load[models.LevelData](store, LevelsTable)
			return len(records), err
		},
		"objects": func() (int, error) {
			records, err := Load[models.Object](store, ObjectsTable)
			return len(records), err
		},
		"skills": func() (int, error) {
			records, err := Load[models.SkillData](store, SkillsTable)
			return len(records), err
		},
		"treasure classes": func() (int, error) {
			records, err := Load[models.TreasureClassEx](store, TreasureClassExTable)
			return len(records), err
		},
		"armor": func() (int, error) {
			records, err := Load[models.ItemArmor](store, ArmorTable)
			return len(records), err
		},
		"weapons": func() (int, error) {
			records, err := Load[models.ItemWeapon](store, WeaponsTable)
			return len(records), err
		},
		"misc items": func() (int, error) {
			records, err := Load[models.MiscItem](store, MiscTable)
			return len(records), err
		},
		"item types": func() (int, error) {
			records, err := Load[models.ItemType](store, ItemTypesTable)
			return len(records), err
		},
		"item ratios": func() (int, error) {
			records, err := Load[models.ItemRatio](store, ItemRatioTable)
			return len(records), err
		},
		"item stat costs": func() (int, error) {
			records, err := Load[models.ItemStatCost](store, ItemStatCostTable)
			return len(records), err
		},
		"properties": func() (int, error) {
			records, err := Load[models.ItemProperty](store, PropertiesTable)
			return len(records), err
		},
		"unique items": func() (int, error) {
			records, err := Load[models.ItemUnique](store, UniqueItemsTable)
			return len(records), err
		},
		"set items": func() (int, error) {
			records, err := Load[models.SetItemData](store, SetItemsTable)
			return len(records), err
		},
		"magic prefixes": func() (int, error) {
			records, err := Load[models.MagicPrefix](store, MagicPrefixTable)
			return len(records), err
		},
		"magic suffixes": func() (int, error) {
			records, err := Load[models.MagicSuffix](store, MagicSuffixTable)
			return len(records), err
		},
		"automagic": func() (int, error) {
			records, err := Load[models.AutoMagicData](store, AutoMagicTable)
			return len(records), err
		},
		"rare prefixes": func() (int, error) {
			records, err := Load[models.RarePrefix](store, RarePrefixTable)
			return len(records), err
		},
		"rare suffixes": func() (int, error) {
			records, err := Load[models.RareSuffix](store, RareSuffixTable)
			return len(records), err
		},
		"gems": func() (int, error) {
			records, err := Load[models.GemData](store, GemsTable)
			return len(records), err
		},
		"rune words": func() (int, error) {
			records, err := Load[models.RuneWordData](store, RunesTable)
			return len(records), err
		},
		"cube recipes": func() (int, error) {
			records, err := Load[models.CubeRecipe](store, CubeMainTable)
			return len(records), err
		},
		"sets": func() (int, error) {
			records, err := Load[models.SetBonusData](store, SetsTable)
			return len(records), err
		},
		"level types": func() (int, error) {
			records, err := Load[models.LevelType](store, LevelTypesTable)
			return len(records), err
		},
		"level presets": func() (int, error) {
			records, err := Load[models.LevelPreset](store, LevelPresetsTable)
			return len(records), err
		},
		"level mazes": func() (int, error) {
			records, err := Load[models.LevelMazeData](store, LevelMazeTable)
			return len(records), err
		},
		"level warps": func() (int, error) {
			records, err := Load[models.LevelWarp](store, LevelWarpTable)
			return len(records), err
		},
		"level substitutions": func() (int, error) {
			records, err := Load[models.LevelSubstitutionData](store, LevelSubTable)
			return len(records), err
		},
		"monster stats": func() (int, error) {
			records, err := Load[models.MonsterStats](store, MonsterStatsTable)
			return len(records), err
		},
		"monster graphics": func() (int, error) {
			records, err := Load[models.MonsterStats2](store, MonsterStats2Table)
			return len(records), err
		},
		"monster levels": func() (int, error) {
			records, err := Load[models.MonsterLevelStats](store, MonsterLevelsTable)
			return len(records), err
		},
		"monster properties": func() (int, error) {
			records, err := Load[models.MonsterProp](store, MonsterPropsTable)
			return len(records), err
		},
		"monster sounds": func() (int, error) {
			records, err := Load[models.MonsterSounds](store, MonsterSoundsTable)
			return len(records), err
		},
		"monster equipment": func() (int, error) {
			records, err := Load[models.MonsterEquipment](store, MonsterEquipTable)
			return len(records), err
		},
		"missiles": func() (int, error) {
			records, err := Load[models.Missile](store, MissilesTable)
			return len(records), err
		},
		"states": func() (int, error) { records, err := Load[models.State](store, StatesTable); return len(records), err },
		"overlays": func() (int, error) {
			records, err := Load[models.Overlay](store, OverlaysTable)
			return len(records), err
		},
		"pet types": func() (int, error) {
			records, err := Load[models.PetType](store, PetTypesTable)
			return len(records), err
		},
	} {
		count, err := load()
		if err != nil {
			t.Errorf("decode typed %s: %v", name, err)
		} else if count == 0 {
			t.Errorf("typed %s table is empty", name)
		}
	}
}
