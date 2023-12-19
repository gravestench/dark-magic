package recordManager

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
)

type Service struct {
	logger *slog.Logger
	cfg    configFile.Dependency
	tsv    tsvLoader.Dependency
	tbl    tblLoader.Dependency

	belts                  []models.BeltData
	charStartingAttributes []models.CharStats
	inventory              []models.InventoryData
	overlays               []models.Overlay
	petTypes               []models.PetType
	autoMapEntries         []models.AutoMapEntry
	states                 []models.State
	hirelings              []models.Hireling
	hirelingDescriptions   []models.HirelingDescription
	missiles               []models.Missile
	difficultyLevels       []models.Difficultylevel
	shrines                []models.Shrine
	gambleRecords          []models.GambleRecord
	npcTradeRecords        []models.NPCTrade
	experienceBreakpoints  []models.ExperienceData

	itemArmor         []models.ItemArmor
	itemWeapon        []models.ItemWeapon
	itemWeaponClass   []models.WeaponClass
	itemMisc          []models.MiscItem
	itemTypes         []models.ItemType
	itemAutoMagic     []models.AutoMagicData
	itemStatCost      []models.ItemStatCost
	itemRatio         []models.ItemRatio
	itemUnique        []models.ItemUnique
	itemHiQualityMods []models.ItemHighQualityModifiers
	itemProperties    []models.ItemProperty

	cubeRecipes []models.CubeRecipe
	books       []models.Book
	gems        []models.GemData
	runes       []models.RuneWordData
	setItems    []models.SetItemData
	setBonuses  []models.SetBonusData

	skills    []models.SkillData
	skillDesc []models.SkillDescData

	treasures          []models.TreasureClass
	treasuresExpansion []models.TreasureClassEx

	magicPrefixes  []models.MagicPrefix
	magicSuffixes  []models.MagicSuffix
	rarePrefixes   []models.RarePrefix
	rareSuffixes   []models.RareSuffix
	uniquePrefixes []models.UniquePrefix
	uniqueSuffixes []models.UniqueSuffix

	objects      []models.Object
	objectTypes  []models.ObjectType
	objectGroups []models.ObjectGroup
	objectModes  []models.ObjectMode
	//ObjectPresets []models.ObjectPreset // in d2resurrected only?

	sounds            []models.SoundEntry
	soundEnvironments []models.SoundEnvironment

	levelPresets       []models.LevelPreset
	levelType          []models.LevelType
	levelWarp          []models.LevelWarp
	levelDetails       []models.LevelData
	levelMaze          []models.LevelMazeData
	levelSubstitutions []models.LevelSubstitutionData
	//LevelGroups        []models.LevelGroup // d2r ?

	monsterUniqueModifiers []models.MonsterUniqueModifier
	monsterEquipment       []models.MonsterEquipment
	monsterLevelStats      []models.MonsterLevelStats
	monsterPresets         []models.MonsterPreset
	monsterProperties      []models.MonsterProp
	monsterSequences       []models.MonsterSequence
	monsterStats           []models.MonsterStats
	monsterStats2          []models.MonsterStats2
	monsterSounds          []models.MonsterSounds
	monsterUniqueNames     []models.MonsterUniqueName

	loaded bool
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.LoadRecords()
}

func (s *Service) Name() string {
	return "Record Manager"
}

func (s *Service) Ready() bool {
	for _, dependency := range []any{
		s.cfg,
	} {
		if dependency == nil {
			return false
		}
	}

	return true
}

func (s *Service) IsLoaded() bool {
	return s.loaded
}

func (s *Service) LoadRecords() error {
	s.logger.Info("loading records")
	for path, destination := range map[string]any{
		PathLevelPreset:        &s.levelPresets,
		PathLevelType:          &s.levelType,
		PathLevelWarp:          &s.levelWarp,
		PathLevelDetails:       &s.levelDetails,
		PathLevelMaze:          &s.levelMaze,
		PathLevelSubstitutions: &s.levelSubstitutions,
		//PathLevelGroups:        &s.levelGroups, // d2r ?

		PathObjectDetails: &s.objects,
		PathObjectType:    &s.objectTypes,
		PathObjectGroup:   &s.objectGroups,
		PathObjectMode:    &s.objectModes,
		// PathObjectPreset:  &s.objectPresets, // in d2resurrected only?

		PathSoundSettings: &s.sounds,
		PathSoundEnvirons: &s.soundEnvironments,

		PathOverlays:            &s.overlays,
		PathHireling:            &s.hirelings,
		PathHirelingDescription: &s.hirelingDescriptions,
		PathAutoMap:             &s.autoMapEntries,
		PathDifficultyLevels:    &s.difficultyLevels,
		PathShrines:             &s.shrines,
		PathBelts:               &s.belts,
		PathGamble:              &s.gambleRecords,
		PathInventory:           &s.inventory,
		PathStates:              &s.states,
		PathPetType:             &s.petTypes,
		PathNPC:                 &s.npcTradeRecords,
		PathMissiles:            &s.missiles,

		PathItemStatCost: &s.itemStatCost,
		PathItemRatio:    &s.itemRatio,
		PathItemTypes:    &s.itemTypes,
		PathAutoMagic:    &s.itemAutoMagic,
		PathWeapons:      &s.itemWeapon,
		PathArmor:        &s.itemArmor,
		PathWeaponClass:  &s.itemWeaponClass,
		PathBooks:        &s.books,
		PathMisc:         &s.itemMisc,
		PathUniqueItems:  &s.itemUnique,
		PathGems:         &s.gems,
		PathRunes:        &s.runes,
		PathSets:         &s.setBonuses,
		PathSetItems:     &s.setItems,

		PathExperience:   &s.experienceBreakpoints,
		PathCharStats:    &s.charStartingAttributes,
		PathSkills:       &s.skills,
		PathSkillDesc:    &s.skillDesc,
		PathQualityItems: &s.itemHiQualityMods,
		PathProperties:   &s.itemProperties,

		PathCubeRecipes: &s.cubeRecipes,

		PathTreasureClass:   &s.treasures,
		PathTreasureClassEx: &s.treasuresExpansion,

		PathMonsterUniqueModifier: &s.monsterUniqueModifiers,
		PathMonsterEquipment:      &s.monsterEquipment,
		PathMonsterLevel:          &s.monsterLevelStats,
		PathMonsterPreset:         &s.monsterPresets,
		PathMonsterProperties:     &s.monsterProperties,
		PathMonsterSequence:       &s.monsterSequences,
		PathMonsterStats:          &s.monsterStats,
		PathMonsterStats2:         &s.monsterStats2,
		PathMonsterSound:          &s.monsterSounds,
		PathUniqueAppellation:     &s.monsterUniqueNames,

		PathMagicPrefix:  &s.magicPrefixes,
		PathMagicSuffix:  &s.magicSuffixes,
		PathRarePrefix:   &s.rarePrefixes,
		PathRareSuffix:   &s.rareSuffixes,
		PathUniquePrefix: &s.uniquePrefixes,
		PathUniqueSuffix: &s.uniqueSuffixes,
	} {
		if destination == nil {
			continue
		}

		if err := s.tsv.Unmarshal(path, destination); err != nil {
			s.logger.Error("parsing records for %q: %v", path, err)
		}
	}

	s.loaded = true
	s.logger.Info("done")

	return nil
}

func (s *Service) Belts() []models.BeltData {
	return s.belts
}

func (s *Service) CharStartingAttributes() []models.CharStats {
	return s.charStartingAttributes
}

func (s *Service) Inventory() []models.InventoryData {
	return s.inventory
}

func (s *Service) Overlays() []models.Overlay {
	return s.overlays
}

func (s *Service) PetTypes() []models.PetType {
	return s.petTypes
}

func (s *Service) AutoMapEntries() []models.AutoMapEntry {
	return s.autoMapEntries
}

func (s *Service) States() []models.State {
	return s.states
}

func (s *Service) Hirelings() []models.Hireling {
	return s.hirelings
}

func (s *Service) HirelingDescriptions() []models.HirelingDescription {
	return s.hirelingDescriptions
}

func (s *Service) Missiles() []models.Missile {
	return s.missiles
}

func (s *Service) DifficultyLevels() []models.Difficultylevel {
	return s.difficultyLevels
}

func (s *Service) Shrines() []models.Shrine {
	return s.shrines
}

func (s *Service) GambleRecords() []models.GambleRecord {
	return s.gambleRecords
}

func (s *Service) NpcTradeRecords() []models.NPCTrade {
	return s.npcTradeRecords
}

func (s *Service) ExperienceBreakpoints() []models.ExperienceData {
	return s.experienceBreakpoints
}

func (s *Service) ItemArmor() []models.ItemArmor {
	return s.itemArmor
}

func (s *Service) ItemWeapon() []models.ItemWeapon {
	return s.itemWeapon
}

func (s *Service) ItemWeaponClass() []models.WeaponClass {
	return s.itemWeaponClass
}

func (s *Service) ItemMisc() []models.MiscItem {
	return s.itemMisc
}

func (s *Service) ItemTypes() []models.ItemType {
	return s.itemTypes
}

func (s *Service) ItemAutoMagic() []models.AutoMagicData {
	return s.itemAutoMagic
}

func (s *Service) ItemStatCost() []models.ItemStatCost {
	return s.itemStatCost
}

func (s *Service) ItemRatio() []models.ItemRatio {
	return s.itemRatio
}

func (s *Service) ItemUnique() []models.ItemUnique {
	return s.itemUnique
}

func (s *Service) ItemHiQualityMods() []models.ItemHighQualityModifiers {
	return s.itemHiQualityMods
}

func (s *Service) ItemProperties() []models.ItemProperty {
	return s.itemProperties
}

func (s *Service) CubeRecipes() []models.CubeRecipe {
	return s.cubeRecipes
}

func (s *Service) Books() []models.Book {
	return s.books
}

func (s *Service) Gems() []models.GemData {
	return s.gems
}

func (s *Service) Runes() []models.RuneWordData {
	return s.runes
}

func (s *Service) SetItems() []models.SetItemData {
	return s.setItems
}

func (s *Service) SetBonuses() []models.SetBonusData {
	return s.setBonuses
}

func (s *Service) Skills() []models.SkillData {
	return s.skills
}

func (s *Service) SkillDesc() []models.SkillDescData {
	return s.skillDesc
}

func (s *Service) Treasures() []models.TreasureClass {
	return s.treasures
}

func (s *Service) TreasuresExpansion() []models.TreasureClassEx {
	return s.treasuresExpansion
}

func (s *Service) MagicPrefixes() []models.MagicPrefix {
	return s.magicPrefixes
}

func (s *Service) MagicSuffixes() []models.MagicSuffix {
	return s.magicSuffixes
}

func (s *Service) RarePrefixes() []models.RarePrefix {
	return s.rarePrefixes
}

func (s *Service) RareSuffixes() []models.RareSuffix {
	return s.rareSuffixes
}

func (s *Service) UniquePrefixes() []models.UniquePrefix {
	return s.uniquePrefixes
}

func (s *Service) UniqueSuffixes() []models.UniqueSuffix {
	return s.uniqueSuffixes
}

func (s *Service) Objects() []models.Object {
	return s.objects
}

func (s *Service) ObjectTypes() []models.ObjectType {
	return s.objectTypes
}

func (s *Service) ObjectGroups() []models.ObjectGroup {
	return s.objectGroups
}

func (s *Service) ObjectModes() []models.ObjectMode {
	return s.objectModes
}

func (s *Service) Sounds() []models.SoundEntry {
	return s.sounds
}

func (s *Service) SoundEnvironments() []models.SoundEnvironment {
	return s.soundEnvironments
}

func (s *Service) LevelPresets() []models.LevelPreset {
	return s.levelPresets
}

func (s *Service) LevelType() []models.LevelType {
	return s.levelType
}

func (s *Service) LevelWarp() []models.LevelWarp {
	return s.levelWarp
}

func (s *Service) LevelDetails() []models.LevelData {
	return s.levelDetails
}

func (s *Service) LevelMaze() []models.LevelMazeData {
	return s.levelMaze
}

func (s *Service) LevelSubstitutions() []models.LevelSubstitutionData {
	return s.levelSubstitutions
}

func (s *Service) MonsterUniqueModifiers() []models.MonsterUniqueModifier {
	return s.monsterUniqueModifiers
}

func (s *Service) MonsterEquipment() []models.MonsterEquipment {
	return s.monsterEquipment
}

func (s *Service) MonsterLevelStats() []models.MonsterLevelStats {
	return s.monsterLevelStats
}

func (s *Service) MonsterPresets() []models.MonsterPreset {
	return s.monsterPresets
}

func (s *Service) MonsterProperties() []models.MonsterProp {
	return s.monsterProperties
}

func (s *Service) MonsterSequences() []models.MonsterSequence {
	return s.monsterSequences
}

func (s *Service) MonsterStats() []models.MonsterStats {
	return s.monsterStats
}

func (s *Service) MonsterStats2() []models.MonsterStats2 {
	return s.monsterStats2
}

func (s *Service) MonsterSounds() []models.MonsterSounds {
	return s.monsterSounds
}

func (s *Service) MonsterUniqueNames() []models.MonsterUniqueName {
	return s.monsterUniqueNames
}
