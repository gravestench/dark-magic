package record_manager

import (
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/tsv_loader"
)

const (
	PathLevelPreset           = "/data/global/excel/LvlPrest.txt"
	PathLevelType             = "/data/global/excel/LvlTypes.txt"
	PathObjectType            = "/data/global/excel/objtype.txt"
	PathLevelWarp             = "/data/global/excel/LvlWarp.txt"
	PathLevelDetails          = "/data/global/excel/Levels.txt"
	PathLevelMaze             = "/data/global/excel/LvlMaze.txt"
	PathLevelSubstitutions    = "/data/global/excel/LvlSub.txt"
	PathObjectDetails         = "/data/global/excel/Objects.txt"
	PathObjectMode            = "/data/global/excel/ObjMode.txt"
	PathObjectGroup           = "/data/global/excel/objgroup.txt"
	PathSoundSettings         = "/data/global/excel/Sounds.txt"
	PathItemStatCost          = "/data/global/excel/ItemStatCost.txt"
	PathItemRatio             = "/data/global/excel/itemratio.txt"
	PathItemTypes             = "/data/global/excel/ItemTypes.txt"
	PathQualityItems          = "/data/global/excel/qualityitems.txt"
	PathOverlays              = "/data/global/excel/Overlay.txt"
	PathRunes                 = "/data/global/excel/runes.txt"
	PathSets                  = "/data/global/excel/Sets.txt"
	PathSetItems              = "/data/global/excel/SetItems.txt"
	PathAutoMagic             = "/data/global/excel/automagic.txt"
	PathProperties            = "/data/global/excel/Properties.txt"
	PathHireling              = "/data/global/excel/hireling.txt"
	PathHirelingDescription   = "/data/global/excel/HireDesc.txt"
	PathDifficultyLevels      = "/data/global/excel/difficultylevels.txt"
	PathAutoMap               = "/data/global/excel/AutoMap.txt"
	PathCubeRecipes           = "/data/global/excel/cubemain.txt"
	PathSkills                = "/data/global/excel/skills.txt"
	PathSkillDesc             = "/data/global/excel/skilldesc.txt"
	PathTreasureClass         = "/data/global/excel/TreasureClass.txt"
	PathTreasureClassEx       = "/data/global/excel/TreasureClassEx.txt"
	PathStates                = "/data/global/excel/states.txt"
	PathSoundEnvirons         = "/data/global/excel/soundenviron.txt"
	PathShrines               = "/data/global/excel/shrines.txt"
	PathPetType               = "/data/global/excel/pettype.txt"
	PathNPC                   = "/data/global/excel/npc.txt"
	PathMonsterUniqueModifier = "/data/global/excel/monumod.txt"
	PathMonsterEquipment      = "/data/global/excel/monequip.txt"
	PathUniqueAppellation     = "/data/global/excel/UniqueAppellation.txt"
	PathMonsterLevel          = "/data/global/excel/monlvl.txt"
	PathMonsterPreset         = "/data/global/excel/MonPreset.txt"
	PathMonsterProperties     = "/data/global/excel/MonProp.txt"
	PathMonsterStats          = "/data/global/excel/monstats.txt"
	PathMonsterStats2         = "/data/global/excel/monstats2.txt"
	PathMonsterSound          = "/data/global/excel/monsounds.txt"
	PathMonsterSequence       = "/data/global/excel/monseq.txt"
	PathBelts                 = "/data/global/excel/belts.txt"
	PathGamble                = "/data/global/excel/gamble.txt"
	PathInventory             = "/data/global/excel/inventory.txt"
	PathWeapons               = "/data/global/excel/weapons.txt"
	PathArmor                 = "/data/global/excel/armor.txt"
	PathWeaponClass           = "/data/global/excel/WeaponClass.txt"
	PathBooks                 = "/data/global/excel/books.txt"
	PathMisc                  = "/data/global/excel/misc.txt"
	PathUniqueItems           = "/data/global/excel/UniqueItems.txt"
	PathGems                  = "/data/global/excel/gems.txt"
	PathMagicPrefix           = "/data/global/excel/MagicPrefix.txt"
	PathMagicSuffix           = "/data/global/excel/MagicSuffix.txt"
	PathRarePrefix            = "/data/global/excel/RarePrefix.txt" // these are for item names
	PathRareSuffix            = "/data/global/excel/RareSuffix.txt"
	PathUniquePrefix          = "/data/global/excel/UniquePrefix.txt"
	PathUniqueSuffix          = "/data/global/excel/UniqueSuffix.txt"
	PathExperience            = "/data/global/excel/experience.txt"
	PathCharStats             = "/data/global/excel/charstats.txt"
	PathMissiles              = "/data/global/excel/Missiles.txt"

	// possible d2 resurrected files
	// PathLevelGroups  = "/data/global/excel/LevelGroups.txt"
	// PathObjectPreset = "/data/global/excel/objpreset.txt"

	// just reference files
	// PathBodyLocations   = "/data/global/excel/bodylocs.txt"
	// PathEvents          = "/data/global/excel/events.txt"
	// PathColors          = "/data/global/excel/colors.txt"
	// PathCompCode        = "/data/global/excel/compcode.txt"
	// PathCubeModifier    = "/data/global/excel/CubeMod.txt"
	// PathElemType        = "/data/global/excel/ElemTypes.txt"
	// PathHitClass        = "/data/global/excel/HitClass.txt"
	// PathLowQualityItems = "/data/global/excel/lowqualityitems.txt"
	// PathMissileCalc     = "/data/global/excel/misscalc.txt"
	// PathPlrMode         = "/data/global/excel/PlrMode.txt"
	// PathPlayerType      = "/data/global/excel/PlrType.txt"
	// PathPlayerClass     = "/data/global/excel/PlayerClass.txt"
	// PathSkillCalc       = "/data/global/excel/skillcalc.txt"
	// PathStorePage       = "/data/global/excel/ItemStorePage.txt"
	// PathComposite       = "/data/global/excel/Composit.txt"
	// PathArmorType       = "/data/global/excel/ArmType.txt"
	// PathCubeType        = "/data/global/excel/CubeType.txt" // pretty sure this is some shit they left on accident
)

type Service struct {
	logger *zerolog.Logger
	cfg    config_file.Dependency
	tsv    tsv_loader.Dependency

	Belts                  []models.BeltData
	CharStartingAttributes []models.CharStats
	Inventory              []models.InventoryData
	Overlays               []models.Overlay
	PetTypes               []models.PetType
	AutoMapEntries         []models.AutoMapEntry
	States                 []models.State
	Hirelings              []models.Hireling
	HirelingDescriptions   []models.HirelingDescription
	Missiles               []models.Missile
	DifficultyLevels       []models.Difficultylevel
	Shrines                []models.Shrine
	GambleRecords          []models.GambleRecord
	NpcTradeRecords        []models.NPCTrade
	ExperienceBreakpoints  []models.ExperienceData

	ItemArmor         []models.ItemArmor
	ItemWeapon        []models.ItemWeapon
	ItemWeaponClass   []models.WeaponClass
	ItemMisc          []models.MiscItem
	ItemTypes         []models.ItemType
	ItemAutoMagic     []models.AutoMagicData
	ItemStatCost      []models.ItemStatCost
	ItemRatio         []models.ItemRatio
	ItemUnique        []models.ItemUnique
	ItemHiQualityMods []models.ItemHighQualityModifiers
	ItemProperties    []models.ItemProperty

	CubeRecipes []models.CubeRecipe
	Books       []models.Book
	Gems        []models.GemData
	Runes       []models.RuneWordData
	SetItems    []models.SetItemData
	SetBonuses  []models.SetBonusData

	Skills    []models.SkillData
	SkillDesc []models.SkillDescData

	Treasures          []models.TreasureClass
	TreasuresExpansion []models.TreasureClassEx

	MagicPrefixes  []models.MagicPrefix
	MagicSuffixes  []models.MagicSuffix
	RarePrefixes   []models.RarePrefix
	RareSuffixes   []models.RareSuffix
	UniquePrefixes []models.UniquePrefix
	UniqueSuffixes []models.UniqueSuffix

	Objects      []models.Object
	ObjectTypes  []models.ObjectType
	ObjectGroups []models.ObjectGroup
	ObjectModes  []models.ObjectMode
	//ObjectPresets []models.ObjectPreset // in d2resurrected only?

	Sounds            []models.SoundEntry
	SoundEnvironments []models.SoundEnvironment

	LevelPresets       []models.LevelPreset
	LevelType          []models.LevelType
	LevelWarp          []models.LevelWarp
	LevelDetails       []models.LevelData
	LevelMaze          []models.LevelMazeData
	LevelSubstitutions []models.LevelSubstitutionData
	LevelGroups        []models.LevelGroup

	MonsterUniqueModifiers []models.MonsterUniqueModifier
	MonsterEquipment       []models.MonsterEquipment
	MonsterLevelStats      []models.MonsterLevelStats
	MonsterPresets         []models.MonsterPreset
	MonsterProperties      []models.MonsterProp
	MonsterSequences       []models.MonsterSequence
	MonsterStats           []models.MonsterStats
	MonsterStats2          []models.MonsterStats2
	MonsterSounds          []models.MonsterSounds
	MonsterUniqueNames     []models.MonsterUniqueName
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) DependenciesResolved() bool {
	if s.cfg == nil {
		return false
	}

	if s.tsv == nil {
		return false
	}

	if !s.tsv.(runtime.HasDependencies).DependenciesResolved() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(runtime runtime.R) {
	for _, service := range runtime.Services() {
		switch candidate := service.(type) {
		case config_file.Dependency:
			s.cfg = candidate
		case tsv_loader.Dependency:
			s.tsv = candidate
		}
	}
}

func (s *Service) Init(rt runtime.R) {
	s.LoadRecords()
}

func (s *Service) Name() string {
	return "Diablo II Record Manager"
}

func (s *Service) LoadRecords() error {
	for path, destination := range map[string]any{
		PathLevelPreset:        &s.LevelPresets,
		PathLevelType:          &s.LevelType,
		PathLevelWarp:          &s.LevelWarp,
		PathLevelDetails:       &s.LevelDetails,
		PathLevelMaze:          &s.LevelMaze,
		PathLevelSubstitutions: &s.LevelSubstitutions,

		PathObjectDetails: &s.Objects,
		PathObjectType:    &s.ObjectTypes,
		PathObjectGroup:   &s.ObjectGroups,
		PathObjectMode:    &s.ObjectModes,
		// PathObjectPreset:  &s.ObjectPresets, // in d2resurrected only?

		PathSoundSettings: &s.Sounds,
		PathSoundEnvirons: &s.SoundEnvironments,

		PathOverlays:            &s.Overlays,
		PathHireling:            &s.Hirelings,
		PathHirelingDescription: &s.HirelingDescriptions,
		PathAutoMap:             &s.AutoMapEntries,
		PathDifficultyLevels:    &s.DifficultyLevels,
		PathShrines:             &s.Shrines,
		PathBelts:               &s.Belts,
		PathGamble:              &s.GambleRecords,
		PathInventory:           &s.Inventory,
		PathStates:              &s.States,
		PathPetType:             &s.PetTypes,
		PathNPC:                 &s.NpcTradeRecords,
		PathMissiles:            &s.Missiles,

		PathItemStatCost: &s.ItemStatCost,
		PathItemRatio:    &s.ItemRatio,
		PathItemTypes:    &s.ItemTypes,
		PathAutoMagic:    &s.ItemAutoMagic,
		PathWeapons:      &s.ItemWeapon,
		PathArmor:        &s.ItemArmor,
		PathWeaponClass:  &s.ItemWeaponClass,
		PathBooks:        &s.Books,
		PathMisc:         &s.ItemMisc,
		PathUniqueItems:  &s.ItemUnique,
		PathGems:         &s.Gems,
		PathRunes:        &s.Runes,
		PathSets:         &s.SetBonuses,
		PathSetItems:     &s.SetItems,

		PathExperience:   &s.ExperienceBreakpoints,
		PathCharStats:    &s.CharStartingAttributes,
		PathSkills:       &s.Skills,
		PathSkillDesc:    &s.SkillDesc,
		PathQualityItems: &s.ItemHiQualityMods,
		PathProperties:   &s.ItemProperties,

		PathCubeRecipes: &s.CubeRecipes,

		PathTreasureClass:   &s.Treasures,
		PathTreasureClassEx: &s.TreasuresExpansion,

		PathMonsterUniqueModifier: &s.MonsterUniqueModifiers,
		PathMonsterEquipment:      &s.MonsterEquipment,
		PathMonsterLevel:          &s.MonsterLevelStats,
		PathMonsterPreset:         &s.MonsterPresets,
		PathMonsterProperties:     &s.MonsterProperties,
		PathMonsterSequence:       &s.MonsterSequences,
		PathMonsterStats:          &s.MonsterStats,
		PathMonsterStats2:         &s.MonsterStats2,
		PathMonsterSound:          &s.MonsterSounds,
		PathUniqueAppellation:     &s.MonsterUniqueNames,

		PathMagicPrefix:  &s.MagicPrefixes,
		PathMagicSuffix:  &s.MagicSuffixes,
		PathRarePrefix:   &s.RarePrefixes,
		PathRareSuffix:   &s.RareSuffixes,
		PathUniquePrefix: &s.UniquePrefixes,
		PathUniqueSuffix: &s.UniqueSuffixes,
	} {
		if destination == nil {
			continue
		}

		if err := s.tsv.Load(path, destination); err != nil {
			s.logger.Error().Msgf("parsing records for '%s': %v", path, err)
		}
	}

	s.logger.Info().Msg("done")

	return nil
}
