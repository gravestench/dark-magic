package gamedata

// Canonical layered-content paths for the first typed catalog slice. Additional
// tables belong here as their schemas are verified and admitted to snapshots.
const (
	CharStatsTable            = "data/global/excel/charstats.txt"
	LevelsTable               = "data/global/excel/Levels.txt"
	ObjectsTable              = "data/global/excel/Objects.txt"
	SkillsTable               = "data/global/excel/skills.txt"
	SoundsTable               = "data/global/excel/Sounds.txt"
	TreasureClassExTable      = "data/global/excel/TreasureClassEx.txt"
	ArmorTable                = "data/global/excel/armor.txt"
	WeaponsTable              = "data/global/excel/weapons.txt"
	MiscTable                 = "data/global/excel/misc.txt"
	ItemTypesTable            = "data/global/excel/ItemTypes.txt"
	ItemRatioTable            = "data/global/excel/ItemRatio.txt"
	ItemStatCostTable         = "data/global/excel/ItemStatCost.txt"
	PropertiesTable           = "data/global/excel/Properties.txt"
	UniqueItemsTable          = "data/global/excel/UniqueItems.txt"
	SetItemsTable             = "data/global/excel/SetItems.txt"
	MagicPrefixTable          = "data/global/excel/MagicPrefix.txt"
	MagicSuffixTable          = "data/global/excel/MagicSuffix.txt"
	AutoMagicTable            = "data/global/excel/AutoMagic.txt"
	RarePrefixTable           = "data/global/excel/RarePrefix.txt"
	RareSuffixTable           = "data/global/excel/RareSuffix.txt"
	GemsTable                 = "data/global/excel/Gems.txt"
	RunesTable                = "data/global/excel/Runes.txt"
	CubeMainTable             = "data/global/excel/CubeMain.txt"
	SetsTable                 = "data/global/excel/Sets.txt"
	LevelTypesTable           = "data/global/excel/LvlTypes.txt"
	LevelPresetsTable         = "data/global/excel/LvlPrest.txt"
	LevelMazeTable            = "data/global/excel/LvlMaze.txt"
	LevelWarpTable            = "data/global/excel/LvlWarp.txt"
	LevelSubTable             = "data/global/excel/LvlSub.txt"
	MonsterStatsTable         = "data/global/excel/MonStats.txt"
	MonsterStats2Table        = "data/global/excel/MonStats2.txt"
	MonsterLevelsTable        = "data/global/excel/MonLvl.txt"
	MonsterPropsTable         = "data/global/excel/MonProp.txt"
	MonsterSoundsTable        = "data/global/excel/MonSounds.txt"
	MonsterEquipTable         = "data/global/excel/MonEquip.txt"
	MissilesTable             = "data/global/excel/Missiles.txt"
	StatesTable               = "data/global/excel/States.txt"
	OverlaysTable             = "data/global/excel/Overlay.txt"
	PetTypesTable             = "data/global/excel/PetType.txt"
	ExperienceTable           = "data/global/excel/Experience.txt"
	InventoryTable            = "data/global/excel/Inventory.txt"
	BeltsTable                = "data/global/excel/Belts.txt"
	HirelingTable             = "data/global/excel/Hireling.txt"
	DifficultyTable           = "data/global/excel/Difficultylevels.txt"
	SkillDescTable            = "data/global/excel/SkillDesc.txt"
	SoundEnvironTable         = "data/global/excel/SoundEnviron.txt"
	AutoMapTable              = "data/global/excel/AutoMap.txt"
	NPCTable                  = "data/global/excel/Npc.txt"
	ShrinesTable              = "data/global/excel/Shrines.txt"
	MonsterPresetsTable       = "data/global/excel/MonPreset.txt"
	GambleTable               = "data/global/excel/Gamble.txt"
	ObjectTypesTable          = "data/global/excel/ObjType.txt"
	ObjectGroupsTable         = "data/global/excel/ObjGroup.txt"
	ObjectModesTable          = "data/global/excel/ObjMode.txt"
	QualityItemsTable         = "data/global/excel/QualityItems.txt"
	WeaponClassTable          = "data/global/excel/WeaponClass.txt"
	BooksTable                = "data/global/excel/Books.txt"
	MonsterSequencesTable     = "data/global/excel/MonSeq.txt"
	MonsterUniqueModsTable    = "data/global/excel/MonUMod.txt"
	UniqueAppellationsTable   = "data/global/excel/UniqueAppellation.txt"
	UniquePrefixesTable       = "data/global/excel/UniquePrefix.txt"
	UniqueSuffixesTable       = "data/global/excel/UniqueSuffix.txt"
	TreasureClassTable        = "data/global/excel/TreasureClass.txt"
	HirelingDescriptionsTable = "data/global/excel/HireDesc.txt"
	SuperUniquesTable         = "data/global/excel/SuperUniques.txt"
	LowQualityItemsTable      = "data/global/excel/LowQualityItems.txt"
	BodyLocationsTable        = "data/global/excel/BodyLocs.txt"
	StorePagesTable           = "data/global/excel/StorePage.txt"
	CompositeComponentsTable  = "data/global/excel/Composit.txt"
	HitClassesTable           = "data/global/excel/HitClass.txt"
)

func isAdmittedTable(path string) bool {
	switch path {
	case CharStatsTable, LevelsTable, ObjectsTable, SkillsTable, SoundsTable,
		TreasureClassExTable, ArmorTable, WeaponsTable, MiscTable, ItemTypesTable,
		ItemRatioTable, ItemStatCostTable, PropertiesTable, UniqueItemsTable,
		SetItemsTable, MagicPrefixTable, MagicSuffixTable, AutoMagicTable,
		RarePrefixTable, RareSuffixTable, GemsTable, RunesTable, CubeMainTable,
		SetsTable, LevelTypesTable, LevelPresetsTable, LevelMazeTable,
		LevelWarpTable, LevelSubTable, MonsterStatsTable, MonsterStats2Table,
		MonsterLevelsTable, MonsterPropsTable, MonsterSoundsTable,
		MonsterEquipTable, MissilesTable, StatesTable, OverlaysTable,
		PetTypesTable, ExperienceTable, InventoryTable, BeltsTable, HirelingTable,
		DifficultyTable, SkillDescTable, SoundEnvironTable, AutoMapTable, NPCTable,
		ShrinesTable, MonsterPresetsTable, GambleTable, ObjectTypesTable,
		ObjectGroupsTable, ObjectModesTable, QualityItemsTable, WeaponClassTable,
		BooksTable, MonsterSequencesTable, MonsterUniqueModsTable,
		UniqueAppellationsTable, UniquePrefixesTable, UniqueSuffixesTable,
		TreasureClassTable, HirelingDescriptionsTable, SuperUniquesTable,
		LowQualityItemsTable, BodyLocationsTable, StorePagesTable,
		CompositeComponentsTable, HitClassesTable:
		return true
	default:
		return false
	}
}
