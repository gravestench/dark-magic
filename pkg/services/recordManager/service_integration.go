package recordManager

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/models"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasDependencies = &Service{}
	//_ configFile.HasDefaultConfig = &Service{}
	_ LoadsDiablo2Records = &Service{}
)

type Dependency = LoadsDiablo2Records

type LoadsDiablo2Records interface {
	LoadRecords() error
	Belts() []models.BeltData
	CharStartingAttributes() []models.CharStats
	Inventory() []models.InventoryData
	Overlays() []models.Overlay
	PetTypes() []models.PetType
	AutoMapEntries() []models.AutoMapEntry
	States() []models.State
	Hirelings() []models.Hireling
	HirelingDescriptions() []models.HirelingDescription
	Missiles() []models.Missile
	DifficultyLevels() []models.Difficultylevel
	Shrines() []models.Shrine
	GambleRecords() []models.GambleRecord
	NpcTradeRecords() []models.NPCTrade
	ExperienceBreakpoints() []models.ExperienceData
	ItemArmor() []models.ItemArmor
	ItemWeapon() []models.ItemWeapon
	ItemWeaponClass() []models.WeaponClass
	ItemMisc() []models.MiscItem
	ItemTypes() []models.ItemType
	ItemAutoMagic() []models.AutoMagicData
	ItemStatCost() []models.ItemStatCost
	ItemRatio() []models.ItemRatio
	ItemUnique() []models.ItemUnique
	ItemHiQualityMods() []models.ItemHighQualityModifiers
	ItemProperties() []models.ItemProperty
	CubeRecipes() []models.CubeRecipe
	Books() []models.Book
	Gems() []models.GemData
	Runes() []models.RuneWordData
	SetItems() []models.SetItemData
	SetBonuses() []models.SetBonusData
	Skills() []models.SkillData
	SkillDesc() []models.SkillDescData
	Treasures() []models.TreasureClass
	TreasuresExpansion() []models.TreasureClassEx
	MagicPrefixes() []models.MagicPrefix
	MagicSuffixes() []models.MagicSuffix
	RarePrefixes() []models.RarePrefix
	RareSuffixes() []models.RareSuffix
	UniquePrefixes() []models.UniquePrefix
	UniqueSuffixes() []models.UniqueSuffix
	Objects() []models.Object
	ObjectTypes() []models.ObjectType
	ObjectGroups() []models.ObjectGroup
	ObjectModes() []models.ObjectMode
	Sounds() []models.SoundEntry
	SoundEnvironments() []models.SoundEnvironment
	LevelPresets() []models.LevelPreset
	LevelType() []models.LevelType
	LevelWarp() []models.LevelWarp
	LevelDetails() []models.LevelData
	LevelMaze() []models.LevelMazeData
	LevelSubstitutions() []models.LevelSubstitutionData
	MonsterUniqueModifiers() []models.MonsterUniqueModifier
	MonsterEquipment() []models.MonsterEquipment
	MonsterLevelStats() []models.MonsterLevelStats
	MonsterPresets() []models.MonsterPreset
	MonsterProperties() []models.MonsterProp
	MonsterSequences() []models.MonsterSequence
	MonsterStats() []models.MonsterStats
	MonsterStats2() []models.MonsterStats2
	MonsterSounds() []models.MonsterSounds
	MonsterUniqueNames() []models.MonsterUniqueName
}
