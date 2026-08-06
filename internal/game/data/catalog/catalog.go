package gamedata

import (
	"fmt"
	"sync"

	"github.com/gravestench/dark-magic/internal/game/data/model"
	"github.com/gravestench/dark-magic/internal/game/data/store"
)

// Snapshot is an immutable, internally consistent view of admitted typed game
// tables. New tables are added here only after their surviving schema passes
// synthetic and real-archive verification.
type Snapshot struct {
	Issues                     []Issue
	CharStats                  []models.CharStats
	CharStatsByClass           map[string]models.CharStats
	Levels                     []models.LevelData
	LevelsByID                 map[int]models.LevelData
	Objects                    []models.Object
	ObjectsByClass             map[int]models.Object
	Skills                     []models.SkillData
	SkillsByID                 map[string]models.SkillData
	Sounds                     []models.SoundEntry
	SoundsByName               map[string]models.SoundEntry
	TreasureClasses            []models.TreasureClassEx
	TreasureByName             map[string]models.TreasureClassEx
	Armor                      []models.ItemArmor
	ArmorByCode                map[string]models.ItemArmor
	Weapons                    []models.ItemWeapon
	WeaponsByCode              map[string]models.ItemWeapon
	Misc                       []models.MiscItem
	MiscByCode                 map[string]models.MiscItem
	ItemTypes                  []models.ItemType
	ItemTypesByCode            map[string]models.ItemType
	ItemRatios                 []models.ItemRatio
	ItemStats                  []models.ItemStatCost
	ItemStatsByName            map[string]models.ItemStatCost
	Properties                 []models.ItemProperty
	PropertiesByCode           map[string]models.ItemProperty
	UniqueItems                []models.ItemUnique
	UniqueByIndex              map[string]models.ItemUnique
	SetItems                   []models.SetItemData
	SetItemsByIndex            map[string]models.SetItemData
	MagicPrefixes              []models.MagicPrefix
	MagicSuffixes              []models.MagicSuffix
	AutoMagic                  []models.AutoMagicData
	RarePrefixes               []models.RarePrefix
	RareSuffixes               []models.RareSuffix
	Gems                       []models.GemData
	GemsByCode                 map[string]models.GemData
	RuneWords                  []models.RuneWordData
	CubeRecipes                []models.CubeRecipe
	Sets                       []models.SetBonusData
	SetsByIndex                map[string]models.SetBonusData
	LevelTypes                 []models.LevelType
	LevelPresets               []models.LevelPreset
	LevelPresetByDef           map[int]models.LevelPreset
	LevelMazes                 []models.LevelMazeData
	LevelMazeByLevel           map[int]models.LevelMazeData
	LevelWarps                 []models.LevelWarp
	LevelSubs                  []models.LevelSubstitutionData
	Monsters                   []models.MonsterStats
	MonstersByID               map[string]models.MonsterStats
	MonsterGraphics            []models.MonsterStats2
	MonsterGfxByID             map[string]models.MonsterStats2
	MonsterLevels              []models.MonsterLevelStats
	MonsterProps               []models.MonsterProp
	MonsterPropsByID           map[string]models.MonsterProp
	MonsterSounds              []models.MonsterSounds
	MonsterSoundByID           map[string]models.MonsterSounds
	MonsterEquipment           []models.MonsterEquipment
	Missiles                   []models.Missile
	MissilesByName             map[string]models.Missile
	States                     []models.State
	StatesByName               map[string]models.State
	Overlays                   []models.Overlay
	OverlaysByName             map[string]models.Overlay
	PetTypes                   []models.PetType
	Experience                 []models.ExperienceData
	Inventories                []models.InventoryData
	InventoryByClass           map[string]models.InventoryData
	Belts                      []models.BeltData
	BeltsByName                map[string]models.BeltData
	Hirelings                  []models.Hireling
	Difficulties               []models.Difficultylevel
	DifficultyByName           map[string]models.Difficultylevel
	SkillDescriptions          []models.SkillDescData
	SkillDescByName            map[string]models.SkillDescData
	SoundEnvironments          []models.SoundEnvironment
	SoundEnvByHandle           map[string]models.SoundEnvironment
	AutoMapEntries             []models.AutoMapEntry
	NPCTrades                  []models.NPCTrade
	NPCTradesByID              map[string]models.NPCTrade
	Shrines                    []models.Shrine
	ShrinesByType              map[string]models.Shrine
	MonsterPresets             []models.MonsterPreset
	GambleItems                []models.GambleRecord
	GambleItemsByCode          map[string]models.GambleRecord
	ObjectTypes                []models.ObjectType
	ObjectTypesByName          map[string]models.ObjectType
	ObjectGroups               []models.ObjectGroup
	ObjectGroupsByName         map[string]models.ObjectGroup
	ObjectModes                []models.ObjectMode
	ObjectModesByName          map[string]models.ObjectMode
	QualityModifiers           []models.ItemHighQualityModifiers
	WeaponClasses              []models.WeaponClass
	WeaponClassByCode          map[models.WeaponClassID]models.WeaponClass
	Books                      []models.Book
	BooksByName                map[string]models.Book
	MonsterSequences           []models.MonsterSequence
	MonsterUniqueMods          []models.MonsterUniqueModifier
	MonsterUniqueByID          map[int]models.MonsterUniqueModifier
	UniqueAppellations         []models.MonsterUniqueAppellation
	UniquePrefixes             []models.UniquePrefix
	UniqueSuffixes             []models.UniqueSuffix
	ClassicTreasureClasses     []models.TreasureClass
	ClassicTreasureByName      map[string]models.TreasureClass
	HirelingDescriptions       []models.HirelingDescription
	HirelingDescriptionByID    map[int]models.HirelingDescription
	SuperUniques               []models.SuperUnique
	SuperUniquesByID           map[string]models.SuperUnique
	SuperUniquesByHardcodedID  map[int]models.SuperUnique
	LowQualityItemNames        []models.LowQualityItemName
	BodyLocations              []models.BodyLocation
	BodyLocationsByCode        map[models.ItemBodyLocation]models.BodyLocation
	StorePages                 []models.StorePage
	StorePagesByCode           map[models.ItemStorePage]models.StorePage
	CompositeComponents        []models.CompositeComponent
	CompositeComponentsByToken map[string]models.CompositeComponent
	HitClasses                 []models.HitClass
	HitClassesByCode           map[string]models.HitClass
	PlayerClasses              []models.PlayerClassCode
	PlayerClassesByCode        map[string]models.PlayerClassCode
	PlayerModes                []models.PlayerMode
	PlayerModesByToken         map[string]models.PlayerMode
	PlayerTypes                []models.PlayerType
	PlayerTypesByToken         map[string]models.PlayerType
	MonsterModes               []models.MonsterMode
	MonsterModesByToken        map[string]models.MonsterMode
	MonsterPlaces              []models.MonsterPlace
	TransformColors            []models.TransformColor
	TransformColorsByCode      map[models.ColorCode]models.TransformColor
	ComponentCodes             []models.ComponentCode
	ComponentCodesByCode       map[string]models.ComponentCode
	ElementTypes               []models.ElementType
	ElementTypesByCode         map[string]models.ElementType
	Events                     []models.UnitEvent
	EventsByName               map[string]models.UnitEvent
	MissileCalculations        []models.MissileCalculation
	MissileCalculationsByCode  map[string]models.MissileCalculation
	SkillCalculations          []models.SkillCalculation
	SkillCalculationsByCode    map[string]models.SkillCalculation
	ArmorTypes                 []models.ArmorType
	ArmorTypesByToken          map[string]models.ArmorType
	CubeModifierTypes          []models.CubeModifierType
	CubeModifierTypesByCode    map[string]models.CubeModifierType
}

// Catalog owns typed record decoding on top of the shared generic row store.
// Snapshot rebuilds are atomic: readers receive either the previous complete
// generation or the next complete generation, never a partially reloaded set.
type Catalog struct {
	store *recordstore.Store
	mu    sync.RWMutex
	data  *Snapshot
}

func New(store *recordstore.Store) *Catalog {
	return &Catalog{store: store}
}

// Snapshot returns a defensive copy of the current typed game-data generation.
func (c *Catalog) Snapshot() (Snapshot, error) {
	if c == nil || c.store == nil {
		return Snapshot{}, fmt.Errorf("gamedata: catalog has no record store")
	}
	c.mu.RLock()
	data := c.data
	c.mu.RUnlock()
	if data != nil {
		return cloneSnapshot(*data), nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data == nil {
		loaded, err := c.load()
		if err != nil {
			return Snapshot{}, err
		}
		c.data = &loaded
	}
	return cloneSnapshot(*c.data), nil
}

// Invalidate clears a changed generic table and drops the typed snapshot when
// that table participates in it. The next reader atomically rebuilds all typed
// indexes from one layered-content generation.
func (c *Catalog) Invalidate(path string) {
	if c == nil || c.store == nil {
		return
	}
	c.store.Invalidate(path)
	if !isAdmittedTable(path) {
		return
	}
	c.mu.Lock()
	c.data = nil
	c.mu.Unlock()
}

func (c *Catalog) load() (Snapshot, error) {
	var issues []Issue
	characters, err := Load[models.CharStats](c.store, CharStatsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	charactersByClass, found, err := ObservedIndex(CharStatsTable, characters, func(record models.CharStats) string { return record.Class })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index charstats: %w", err)
	}
	issues = append(issues, found...)
	levels, err := Load[models.LevelData](c.store, LevelsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	levelsByID, found, err := ObservedIndex(LevelsTable, levels, func(record models.LevelData) int { return record.Id })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index levels: %w", err)
	}
	issues = append(issues, found...)
	objects, err := Load[models.Object](c.store, ObjectsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	objectsByClass, found, err := ObservedIndex(ObjectsTable, objects, func(record models.Object) int { return record.Class })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index objects: %w", err)
	}
	issues = append(issues, found...)
	skills, err := Load[models.SkillData](c.store, SkillsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	skillsByID, found, err := ObservedIndex(SkillsTable, skills, func(record models.SkillData) string { return record.ID })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index skills: %w", err)
	}
	issues = append(issues, found...)
	sounds, err := Load[models.SoundEntry](c.store, SoundsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	soundsByName, found, err := ObservedIndex(SoundsTable, sounds, func(record models.SoundEntry) string { return record.Sound })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index sounds: %w", err)
	}
	issues = append(issues, found...)
	treasure, err := Load[models.TreasureClassEx](c.store, TreasureClassExTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	treasureByName, found, err := ObservedIndex(TreasureClassExTable, treasure, func(record models.TreasureClassEx) string { return record.TreasureClass })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index treasure classes: %w", err)
	}
	issues = append(issues, found...)
	armor, err := Load[models.ItemArmor](c.store, ArmorTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	armorByCode, found, err := ObservedIndex(ArmorTable, armor, func(record models.ItemArmor) string { return record.Code })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index armor: %w", err)
	}
	issues = append(issues, found...)
	weapons, err := Load[models.ItemWeapon](c.store, WeaponsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	weaponsByCode, found, err := ObservedIndex(WeaponsTable, weapons, func(record models.ItemWeapon) string { return record.Code })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index weapons: %w", err)
	}
	issues = append(issues, found...)
	misc, err := Load[models.MiscItem](c.store, MiscTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	miscByCode, found, err := ObservedIndex(MiscTable, misc, func(record models.MiscItem) string { return record.Code })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index misc items: %w", err)
	}
	issues = append(issues, found...)
	itemTypes, err := Load[models.ItemType](c.store, ItemTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	itemTypesByCode, found, err := ObservedIndex(ItemTypesTable, itemTypes, func(record models.ItemType) string { return record.Code })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index item types: %w", err)
	}
	issues = append(issues, found...)
	itemRatios, err := Load[models.ItemRatio](c.store, ItemRatioTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	itemStats, err := Load[models.ItemStatCost](c.store, ItemStatCostTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	itemStatsByName, found, err := ObservedIndex(ItemStatCostTable, itemStats, func(record models.ItemStatCost) string { return record.Name })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index item stats: %w", err)
	}
	issues = append(issues, found...)
	properties, err := Load[models.ItemProperty](c.store, PropertiesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	propertiesByCode, found, err := ObservedIndex(PropertiesTable, properties, func(record models.ItemProperty) string { return record.Code })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index properties: %w", err)
	}
	issues = append(issues, found...)
	uniqueItems, err := Load[models.ItemUnique](c.store, UniqueItemsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	uniqueByIndex, found, err := ObservedIndex(UniqueItemsTable, uniqueItems, func(record models.ItemUnique) string { return record.Index })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index unique items: %w", err)
	}
	issues = append(issues, found...)
	setItems, err := Load[models.SetItemData](c.store, SetItemsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	setItemsByIndex, found, err := ObservedIndex(SetItemsTable, setItems, func(record models.SetItemData) string { return record.Index })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index set items: %w", err)
	}
	issues = append(issues, found...)
	magicPrefixes, err := Load[models.MagicPrefix](c.store, MagicPrefixTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	magicSuffixes, err := Load[models.MagicSuffix](c.store, MagicSuffixTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	autoMagic, err := Load[models.AutoMagicData](c.store, AutoMagicTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	rarePrefixes, err := Load[models.RarePrefix](c.store, RarePrefixTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	rareSuffixes, err := Load[models.RareSuffix](c.store, RareSuffixTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	gems, err := Load[models.GemData](c.store, GemsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	gemsByCode, found, err := ObservedIndex(GemsTable, gems, func(record models.GemData) string { return record.Code })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index gems: %w", err)
	}
	issues = append(issues, found...)
	runeWords, err := Load[models.RuneWordData](c.store, RunesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	cubeRecipes, err := Load[models.CubeRecipe](c.store, CubeMainTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	sets, err := Load[models.SetBonusData](c.store, SetsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	setsByIndex, found, err := ObservedIndex(SetsTable, sets, func(record models.SetBonusData) string { return record.Index })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index sets: %w", err)
	}
	issues = append(issues, found...)
	levelTypes, err := Load[models.LevelType](c.store, LevelTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	levelPresets, err := Load[models.LevelPreset](c.store, LevelPresetsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	levelPresetByDef, found, err := ObservedIndex(LevelPresetsTable, levelPresets, func(record models.LevelPreset) int { return record.Def })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index level presets: %w", err)
	}
	issues = append(issues, found...)
	levelMazes, err := Load[models.LevelMazeData](c.store, LevelMazeTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	levelMazeByLevel, found, err := ObservedIndex(LevelMazeTable, levelMazes, func(record models.LevelMazeData) int { return record.Level })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index level mazes: %w", err)
	}
	issues = append(issues, found...)
	levelWarps, err := Load[models.LevelWarp](c.store, LevelWarpTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	levelSubs, err := Load[models.LevelSubstitutionData](c.store, LevelSubTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsters, err := Load[models.MonsterStats](c.store, MonsterStatsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monstersByID, found, err := ObservedIndex(MonsterStatsTable, monsters, func(record models.MonsterStats) string { return record.Id })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index monsters: %w", err)
	}
	issues = append(issues, found...)
	monsterGraphics, err := Load[models.MonsterStats2](c.store, MonsterStats2Table)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterGfxByID, found, err := ObservedIndex(MonsterStats2Table, monsterGraphics, func(record models.MonsterStats2) string { return record.Id })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index monster graphics: %w", err)
	}
	issues = append(issues, found...)
	monsterLevels, err := Load[models.MonsterLevelStats](c.store, MonsterLevelsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterProps, err := Load[models.MonsterProp](c.store, MonsterPropsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterPropsByID, found, err := ObservedIndex(MonsterPropsTable, monsterProps, func(record models.MonsterProp) string { return record.ID })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index monster properties: %w", err)
	}
	issues = append(issues, found...)
	monsterSounds, err := Load[models.MonsterSounds](c.store, MonsterSoundsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterSoundByID, found, err := ObservedIndex(MonsterSoundsTable, monsterSounds, func(record models.MonsterSounds) string { return record.ID })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index monster sounds: %w", err)
	}
	issues = append(issues, found...)
	monsterEquipment, err := Load[models.MonsterEquipment](c.store, MonsterEquipTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	missiles, err := Load[models.Missile](c.store, MissilesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	missilesByName, found, err := ObservedIndex(MissilesTable, missiles, func(record models.Missile) string { return record.Missile })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index missiles: %w", err)
	}
	issues = append(issues, found...)
	states, err := Load[models.State](c.store, StatesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	statesByName, found, err := ObservedIndex(StatesTable, states, func(record models.State) string { return record.State })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index states: %w", err)
	}
	issues = append(issues, found...)
	overlays, err := Load[models.Overlay](c.store, OverlaysTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	overlaysByName, found, err := ObservedIndex(OverlaysTable, overlays, func(record models.Overlay) string { return record.Overlay })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index overlays: %w", err)
	}
	issues = append(issues, found...)
	petTypes, err := Load[models.PetType](c.store, PetTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	experience, err := Load[models.ExperienceData](c.store, ExperienceTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	inventories, err := Load[models.InventoryData](c.store, InventoryTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	inventoryByClass, found, err := ObservedIndex(InventoryTable, inventories, func(record models.InventoryData) string { return record.Class })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	belts, err := Load[models.BeltData](c.store, BeltsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	beltsByName, found, err := ObservedIndex(BeltsTable, belts, func(record models.BeltData) string { return record.Name })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	hirelings, err := Load[models.Hireling](c.store, HirelingTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	difficulties, err := Load[models.Difficultylevel](c.store, DifficultyTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	difficultyByName, found, err := ObservedIndex(DifficultyTable, difficulties, func(record models.Difficultylevel) string { return record.Name })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	skillDescriptions, err := Load[models.SkillDescData](c.store, SkillDescTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	skillDescByName, found, err := ObservedIndex(SkillDescTable, skillDescriptions, func(record models.SkillDescData) string { return record.SkillDesc })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	soundEnvironments, err := Load[models.SoundEnvironment](c.store, SoundEnvironTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	soundEnvByHandle, found, err := ObservedIndex(SoundEnvironTable, soundEnvironments, func(record models.SoundEnvironment) string { return record.Handle })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	autoMapEntries, err := Load[models.AutoMapEntry](c.store, AutoMapTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	npcTrades, err := Load[models.NPCTrade](c.store, NPCTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	npcTradesByID, found, err := ObservedIndex(NPCTable, npcTrades, func(record models.NPCTrade) string { return record.NPC })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	shrines, err := Load[models.Shrine](c.store, ShrinesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	shrinesByType, found, err := ObservedIndex(ShrinesTable, shrines, func(record models.Shrine) string { return record.ShrineType })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	monsterPresets, err := Load[models.MonsterPreset](c.store, MonsterPresetsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	gambleItems, err := Load[models.GambleRecord](c.store, GambleTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	gambleItemsByCode, found, err := ObservedIndex(GambleTable, gambleItems, func(record models.GambleRecord) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	objectTypes, err := Load[models.ObjectType](c.store, ObjectTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	objectTypesByName, found, err := ObservedIndex(ObjectTypesTable, objectTypes, func(record models.ObjectType) string { return record.Name })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	objectGroups, err := Load[models.ObjectGroup](c.store, ObjectGroupsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	objectGroupsByName, found, err := ObservedIndex(ObjectGroupsTable, objectGroups, func(record models.ObjectGroup) string { return record.GroupName })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	objectModes, err := Load[models.ObjectMode](c.store, ObjectModesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	objectModesByName, found, err := ObservedIndex(ObjectModesTable, objectModes, func(record models.ObjectMode) string { return record.Name })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	qualityModifiers, err := Load[models.ItemHighQualityModifiers](c.store, QualityItemsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	weaponClasses, err := Load[models.WeaponClass](c.store, WeaponClassTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	weaponClassByCode, found, err := ObservedIndex(WeaponClassTable, weaponClasses, func(record models.WeaponClass) models.WeaponClassID { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	books, err := Load[models.Book](c.store, BooksTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	booksByName, found, err := ObservedIndex(BooksTable, books, func(record models.Book) string { return record.Name })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	monsterSequences, err := Load[models.MonsterSequence](c.store, MonsterSequencesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterUniqueMods, err := Load[models.MonsterUniqueModifier](c.store, MonsterUniqueModsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterUniqueByID, found, err := ObservedIndex(MonsterUniqueModsTable, monsterUniqueMods, func(record models.MonsterUniqueModifier) int { return record.ID })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	uniqueAppellations, err := Load[models.MonsterUniqueAppellation](c.store, UniqueAppellationsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	uniquePrefixes, err := Load[models.UniquePrefix](c.store, UniquePrefixesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	uniqueSuffixes, err := Load[models.UniqueSuffix](c.store, UniqueSuffixesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	classicTreasure, err := Load[models.TreasureClass](c.store, TreasureClassTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	classicTreasureByName, found, err := ObservedIndex(TreasureClassTable, classicTreasure, func(record models.TreasureClass) string { return record.TreasureClass })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	hirelingDescriptions, err := Load[models.HirelingDescription](c.store, HirelingDescriptionsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	hirelingDescriptionByID, found, err := ObservedIndex(HirelingDescriptionsTable, hirelingDescriptions, func(record models.HirelingDescription) int { return record.ID })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	superUniques, err := Load[models.SuperUnique](c.store, SuperUniquesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	superUniquesByID, found, err := ObservedIndex(SuperUniquesTable, superUniques, func(record models.SuperUnique) string { return record.ID })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	superUniquesByHardcodedID, found, err := ObservedIndex(SuperUniquesTable, superUniques, func(record models.SuperUnique) int { return record.HardcodedID })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	lowQualityItemNames, err := Load[models.LowQualityItemName](c.store, LowQualityItemsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	bodyLocations, err := Load[models.BodyLocation](c.store, BodyLocationsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	bodyLocationsByCode, found, err := ObservedIndex(BodyLocationsTable, bodyLocations, func(record models.BodyLocation) models.ItemBodyLocation { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	storePages, err := Load[models.StorePage](c.store, StorePagesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	storePagesByCode, found, err := ObservedIndex(StorePagesTable, storePages, func(record models.StorePage) models.ItemStorePage { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	compositeComponents, err := Load[models.CompositeComponent](c.store, CompositeComponentsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	compositeComponentsByToken, found, err := ObservedIndex(CompositeComponentsTable, compositeComponents, func(record models.CompositeComponent) string { return record.Token })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	hitClasses, err := Load[models.HitClass](c.store, HitClassesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	hitClassesByCode, found, err := ObservedIndex(HitClassesTable, hitClasses, func(record models.HitClass) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	playerClasses, err := Load[models.PlayerClassCode](c.store, PlayerClassesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	playerClassesByCode, found, err := ObservedIndex(PlayerClassesTable, playerClasses, func(record models.PlayerClassCode) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	playerModes, err := Load[models.PlayerMode](c.store, PlayerModesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	playerModesByToken, found, err := ObservedIndex(PlayerModesTable, playerModes, func(record models.PlayerMode) string { return record.Token })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	playerTypes, err := Load[models.PlayerType](c.store, PlayerTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	playerTypesByToken, found, err := ObservedIndex(PlayerTypesTable, playerTypes, func(record models.PlayerType) string { return record.Token })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	monsterModes, err := Load[models.MonsterMode](c.store, MonsterModesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	monsterModesByToken, found, err := ObservedIndex(MonsterModesTable, monsterModes, func(record models.MonsterMode) string { return record.Token })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	monsterPlaces, err := Load[models.MonsterPlace](c.store, MonsterPlacesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	transformColors, err := Load[models.TransformColor](c.store, ColorsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	transformColorsByCode, found, err := ObservedIndex(ColorsTable, transformColors, func(record models.TransformColor) models.ColorCode { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	componentCodes, err := Load[models.ComponentCode](c.store, ComponentCodesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	componentCodesByCode, found, err := ObservedIndex(ComponentCodesTable, componentCodes, func(record models.ComponentCode) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	elementTypes, err := Load[models.ElementType](c.store, ElementTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	elementTypesByCode, found, err := ObservedIndex(ElementTypesTable, elementTypes, func(record models.ElementType) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	events, err := Load[models.UnitEvent](c.store, EventsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	eventsByName, found, err := ObservedIndex(EventsTable, events, func(record models.UnitEvent) string { return record.Event })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	missileCalculations, err := Load[models.MissileCalculation](c.store, MissileCalculationsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	missileCalculationsByCode, found, err := ObservedIndex(MissileCalculationsTable, missileCalculations, func(record models.MissileCalculation) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	skillCalculations, err := Load[models.SkillCalculation](c.store, SkillCalculationsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	skillCalculationsByCode, found, err := ObservedIndex(SkillCalculationsTable, skillCalculations, func(record models.SkillCalculation) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	armorTypes, err := Load[models.ArmorType](c.store, ArmorTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	armorTypesByToken, found, err := ObservedIndex(ArmorTypesTable, armorTypes, func(record models.ArmorType) string { return record.Token })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	cubeModifierTypes, err := Load[models.CubeModifierType](c.store, CubeModifierTypesTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	cubeModifierTypesByCode, found, err := ObservedIndex(CubeModifierTypesTable, cubeModifierTypes, func(record models.CubeModifierType) string { return record.Code })
	if err != nil {
		return Snapshot{}, err
	}
	issues = append(issues, found...)
	return Snapshot{
		Issues:    issues,
		CharStats: characters, CharStatsByClass: charactersByClass,
		Levels: levels, LevelsByID: levelsByID,
		Objects: objects, ObjectsByClass: objectsByClass,
		Skills: skills, SkillsByID: skillsByID,
		Sounds: sounds, SoundsByName: soundsByName,
		TreasureClasses: treasure, TreasureByName: treasureByName,
		Armor: armor, ArmorByCode: armorByCode,
		Weapons: weapons, WeaponsByCode: weaponsByCode,
		Misc: misc, MiscByCode: miscByCode,
		ItemTypes: itemTypes, ItemTypesByCode: itemTypesByCode,
		ItemRatios: itemRatios,
		ItemStats:  itemStats, ItemStatsByName: itemStatsByName,
		Properties: properties, PropertiesByCode: propertiesByCode,
		UniqueItems: uniqueItems, UniqueByIndex: uniqueByIndex,
		SetItems: setItems, SetItemsByIndex: setItemsByIndex,
		MagicPrefixes: magicPrefixes, MagicSuffixes: magicSuffixes,
		AutoMagic: autoMagic, RarePrefixes: rarePrefixes, RareSuffixes: rareSuffixes,
		Gems: gems, GemsByCode: gemsByCode, RuneWords: runeWords, CubeRecipes: cubeRecipes,
		Sets: sets, SetsByIndex: setsByIndex,
		LevelTypes: levelTypes, LevelPresets: levelPresets, LevelPresetByDef: levelPresetByDef,
		LevelMazes: levelMazes, LevelMazeByLevel: levelMazeByLevel, LevelWarps: levelWarps, LevelSubs: levelSubs,
		Monsters: monsters, MonstersByID: monstersByID, MonsterGraphics: monsterGraphics, MonsterGfxByID: monsterGfxByID,
		MonsterLevels: monsterLevels, MonsterProps: monsterProps, MonsterPropsByID: monsterPropsByID,
		MonsterSounds: monsterSounds, MonsterSoundByID: monsterSoundByID, MonsterEquipment: monsterEquipment,
		Missiles: missiles, MissilesByName: missilesByName, States: states, StatesByName: statesByName,
		Overlays: overlays, OverlaysByName: overlaysByName, PetTypes: petTypes,
		Experience: experience, Inventories: inventories, InventoryByClass: inventoryByClass,
		Belts: belts, BeltsByName: beltsByName, Hirelings: hirelings, Difficulties: difficulties, DifficultyByName: difficultyByName,
		SkillDescriptions: skillDescriptions, SkillDescByName: skillDescByName,
		SoundEnvironments: soundEnvironments, SoundEnvByHandle: soundEnvByHandle, AutoMapEntries: autoMapEntries,
		NPCTrades: npcTrades, NPCTradesByID: npcTradesByID, Shrines: shrines, ShrinesByType: shrinesByType,
		MonsterPresets: monsterPresets, GambleItems: gambleItems, GambleItemsByCode: gambleItemsByCode,
		ObjectTypes: objectTypes, ObjectTypesByName: objectTypesByName,
		ObjectGroups: objectGroups, ObjectGroupsByName: objectGroupsByName,
		ObjectModes: objectModes, ObjectModesByName: objectModesByName,
		QualityModifiers: qualityModifiers, WeaponClasses: weaponClasses, WeaponClassByCode: weaponClassByCode,
		Books: books, BooksByName: booksByName,
		MonsterSequences: monsterSequences, MonsterUniqueMods: monsterUniqueMods, MonsterUniqueByID: monsterUniqueByID,
		UniqueAppellations: uniqueAppellations, UniquePrefixes: uniquePrefixes, UniqueSuffixes: uniqueSuffixes,
		ClassicTreasureClasses: classicTreasure, ClassicTreasureByName: classicTreasureByName,
		HirelingDescriptions: hirelingDescriptions, HirelingDescriptionByID: hirelingDescriptionByID,
		SuperUniques: superUniques, SuperUniquesByID: superUniquesByID, SuperUniquesByHardcodedID: superUniquesByHardcodedID,
		LowQualityItemNames: lowQualityItemNames, BodyLocations: bodyLocations, BodyLocationsByCode: bodyLocationsByCode,
		StorePages: storePages, StorePagesByCode: storePagesByCode,
		CompositeComponents: compositeComponents, CompositeComponentsByToken: compositeComponentsByToken,
		HitClasses: hitClasses, HitClassesByCode: hitClassesByCode,
		PlayerClasses: playerClasses, PlayerClassesByCode: playerClassesByCode,
		PlayerModes: playerModes, PlayerModesByToken: playerModesByToken,
		PlayerTypes: playerTypes, PlayerTypesByToken: playerTypesByToken,
		MonsterModes: monsterModes, MonsterModesByToken: monsterModesByToken, MonsterPlaces: monsterPlaces,
		TransformColors: transformColors, TransformColorsByCode: transformColorsByCode,
		ComponentCodes: componentCodes, ComponentCodesByCode: componentCodesByCode,
		ElementTypes: elementTypes, ElementTypesByCode: elementTypesByCode, Events: events, EventsByName: eventsByName,
		MissileCalculations: missileCalculations, MissileCalculationsByCode: missileCalculationsByCode,
		SkillCalculations: skillCalculations, SkillCalculationsByCode: skillCalculationsByCode,
		ArmorTypes: armorTypes, ArmorTypesByToken: armorTypesByToken,
		CubeModifierTypes: cubeModifierTypes, CubeModifierTypesByCode: cubeModifierTypesByCode,
	}, nil
}

func cloneSnapshot(source Snapshot) Snapshot {
	result := Snapshot{
		Issues:                     append([]Issue(nil), source.Issues...),
		CharStats:                  append([]models.CharStats(nil), source.CharStats...),
		CharStatsByClass:           make(map[string]models.CharStats, len(source.CharStatsByClass)),
		Levels:                     append([]models.LevelData(nil), source.Levels...),
		LevelsByID:                 make(map[int]models.LevelData, len(source.LevelsByID)),
		Objects:                    append([]models.Object(nil), source.Objects...),
		ObjectsByClass:             make(map[int]models.Object, len(source.ObjectsByClass)),
		Skills:                     append([]models.SkillData(nil), source.Skills...),
		SkillsByID:                 make(map[string]models.SkillData, len(source.SkillsByID)),
		Sounds:                     append([]models.SoundEntry(nil), source.Sounds...),
		SoundsByName:               make(map[string]models.SoundEntry, len(source.SoundsByName)),
		TreasureClasses:            append([]models.TreasureClassEx(nil), source.TreasureClasses...),
		TreasureByName:             make(map[string]models.TreasureClassEx, len(source.TreasureByName)),
		Armor:                      append([]models.ItemArmor(nil), source.Armor...),
		ArmorByCode:                make(map[string]models.ItemArmor, len(source.ArmorByCode)),
		Weapons:                    append([]models.ItemWeapon(nil), source.Weapons...),
		WeaponsByCode:              make(map[string]models.ItemWeapon, len(source.WeaponsByCode)),
		Misc:                       append([]models.MiscItem(nil), source.Misc...),
		MiscByCode:                 make(map[string]models.MiscItem, len(source.MiscByCode)),
		ItemTypes:                  append([]models.ItemType(nil), source.ItemTypes...),
		ItemTypesByCode:            make(map[string]models.ItemType, len(source.ItemTypesByCode)),
		ItemRatios:                 append([]models.ItemRatio(nil), source.ItemRatios...),
		ItemStats:                  append([]models.ItemStatCost(nil), source.ItemStats...),
		ItemStatsByName:            make(map[string]models.ItemStatCost, len(source.ItemStatsByName)),
		Properties:                 append([]models.ItemProperty(nil), source.Properties...),
		PropertiesByCode:           make(map[string]models.ItemProperty, len(source.PropertiesByCode)),
		UniqueItems:                append([]models.ItemUnique(nil), source.UniqueItems...),
		UniqueByIndex:              make(map[string]models.ItemUnique, len(source.UniqueByIndex)),
		SetItems:                   append([]models.SetItemData(nil), source.SetItems...),
		SetItemsByIndex:            make(map[string]models.SetItemData, len(source.SetItemsByIndex)),
		MagicPrefixes:              append([]models.MagicPrefix(nil), source.MagicPrefixes...),
		MagicSuffixes:              append([]models.MagicSuffix(nil), source.MagicSuffixes...),
		AutoMagic:                  append([]models.AutoMagicData(nil), source.AutoMagic...),
		RarePrefixes:               append([]models.RarePrefix(nil), source.RarePrefixes...),
		RareSuffixes:               append([]models.RareSuffix(nil), source.RareSuffixes...),
		Gems:                       append([]models.GemData(nil), source.Gems...),
		GemsByCode:                 make(map[string]models.GemData, len(source.GemsByCode)),
		RuneWords:                  append([]models.RuneWordData(nil), source.RuneWords...),
		CubeRecipes:                append([]models.CubeRecipe(nil), source.CubeRecipes...),
		Sets:                       append([]models.SetBonusData(nil), source.Sets...),
		SetsByIndex:                make(map[string]models.SetBonusData, len(source.SetsByIndex)),
		LevelTypes:                 append([]models.LevelType(nil), source.LevelTypes...),
		LevelPresets:               append([]models.LevelPreset(nil), source.LevelPresets...),
		LevelPresetByDef:           make(map[int]models.LevelPreset, len(source.LevelPresetByDef)),
		LevelMazes:                 append([]models.LevelMazeData(nil), source.LevelMazes...),
		LevelMazeByLevel:           make(map[int]models.LevelMazeData, len(source.LevelMazeByLevel)),
		LevelWarps:                 append([]models.LevelWarp(nil), source.LevelWarps...),
		LevelSubs:                  append([]models.LevelSubstitutionData(nil), source.LevelSubs...),
		Monsters:                   append([]models.MonsterStats(nil), source.Monsters...),
		MonstersByID:               make(map[string]models.MonsterStats, len(source.MonstersByID)),
		MonsterGraphics:            append([]models.MonsterStats2(nil), source.MonsterGraphics...),
		MonsterGfxByID:             make(map[string]models.MonsterStats2, len(source.MonsterGfxByID)),
		MonsterLevels:              append([]models.MonsterLevelStats(nil), source.MonsterLevels...),
		MonsterProps:               append([]models.MonsterProp(nil), source.MonsterProps...),
		MonsterPropsByID:           make(map[string]models.MonsterProp, len(source.MonsterPropsByID)),
		MonsterSounds:              append([]models.MonsterSounds(nil), source.MonsterSounds...),
		MonsterSoundByID:           make(map[string]models.MonsterSounds, len(source.MonsterSoundByID)),
		MonsterEquipment:           append([]models.MonsterEquipment(nil), source.MonsterEquipment...),
		Missiles:                   append([]models.Missile(nil), source.Missiles...),
		MissilesByName:             make(map[string]models.Missile, len(source.MissilesByName)),
		States:                     append([]models.State(nil), source.States...),
		StatesByName:               make(map[string]models.State, len(source.StatesByName)),
		Overlays:                   append([]models.Overlay(nil), source.Overlays...),
		OverlaysByName:             make(map[string]models.Overlay, len(source.OverlaysByName)),
		PetTypes:                   append([]models.PetType(nil), source.PetTypes...),
		Experience:                 append([]models.ExperienceData(nil), source.Experience...),
		Inventories:                append([]models.InventoryData(nil), source.Inventories...),
		InventoryByClass:           make(map[string]models.InventoryData, len(source.InventoryByClass)),
		Belts:                      append([]models.BeltData(nil), source.Belts...),
		BeltsByName:                make(map[string]models.BeltData, len(source.BeltsByName)),
		Hirelings:                  append([]models.Hireling(nil), source.Hirelings...),
		Difficulties:               append([]models.Difficultylevel(nil), source.Difficulties...),
		DifficultyByName:           make(map[string]models.Difficultylevel, len(source.DifficultyByName)),
		SkillDescriptions:          append([]models.SkillDescData(nil), source.SkillDescriptions...),
		SkillDescByName:            make(map[string]models.SkillDescData, len(source.SkillDescByName)),
		SoundEnvironments:          append([]models.SoundEnvironment(nil), source.SoundEnvironments...),
		SoundEnvByHandle:           make(map[string]models.SoundEnvironment, len(source.SoundEnvByHandle)),
		AutoMapEntries:             append([]models.AutoMapEntry(nil), source.AutoMapEntries...),
		NPCTrades:                  append([]models.NPCTrade(nil), source.NPCTrades...),
		NPCTradesByID:              make(map[string]models.NPCTrade, len(source.NPCTradesByID)),
		Shrines:                    append([]models.Shrine(nil), source.Shrines...),
		ShrinesByType:              make(map[string]models.Shrine, len(source.ShrinesByType)),
		MonsterPresets:             append([]models.MonsterPreset(nil), source.MonsterPresets...),
		GambleItems:                append([]models.GambleRecord(nil), source.GambleItems...),
		GambleItemsByCode:          make(map[string]models.GambleRecord, len(source.GambleItemsByCode)),
		ObjectTypes:                append([]models.ObjectType(nil), source.ObjectTypes...),
		ObjectTypesByName:          make(map[string]models.ObjectType, len(source.ObjectTypesByName)),
		ObjectGroups:               append([]models.ObjectGroup(nil), source.ObjectGroups...),
		ObjectGroupsByName:         make(map[string]models.ObjectGroup, len(source.ObjectGroupsByName)),
		ObjectModes:                append([]models.ObjectMode(nil), source.ObjectModes...),
		ObjectModesByName:          make(map[string]models.ObjectMode, len(source.ObjectModesByName)),
		QualityModifiers:           append([]models.ItemHighQualityModifiers(nil), source.QualityModifiers...),
		WeaponClasses:              append([]models.WeaponClass(nil), source.WeaponClasses...),
		WeaponClassByCode:          make(map[models.WeaponClassID]models.WeaponClass, len(source.WeaponClassByCode)),
		Books:                      append([]models.Book(nil), source.Books...),
		BooksByName:                make(map[string]models.Book, len(source.BooksByName)),
		MonsterSequences:           append([]models.MonsterSequence(nil), source.MonsterSequences...),
		MonsterUniqueMods:          append([]models.MonsterUniqueModifier(nil), source.MonsterUniqueMods...),
		MonsterUniqueByID:          make(map[int]models.MonsterUniqueModifier, len(source.MonsterUniqueByID)),
		UniqueAppellations:         append([]models.MonsterUniqueAppellation(nil), source.UniqueAppellations...),
		UniquePrefixes:             append([]models.UniquePrefix(nil), source.UniquePrefixes...),
		UniqueSuffixes:             append([]models.UniqueSuffix(nil), source.UniqueSuffixes...),
		ClassicTreasureClasses:     append([]models.TreasureClass(nil), source.ClassicTreasureClasses...),
		ClassicTreasureByName:      make(map[string]models.TreasureClass, len(source.ClassicTreasureByName)),
		HirelingDescriptions:       append([]models.HirelingDescription(nil), source.HirelingDescriptions...),
		HirelingDescriptionByID:    make(map[int]models.HirelingDescription, len(source.HirelingDescriptionByID)),
		SuperUniques:               append([]models.SuperUnique(nil), source.SuperUniques...),
		SuperUniquesByID:           make(map[string]models.SuperUnique, len(source.SuperUniquesByID)),
		SuperUniquesByHardcodedID:  make(map[int]models.SuperUnique, len(source.SuperUniquesByHardcodedID)),
		LowQualityItemNames:        append([]models.LowQualityItemName(nil), source.LowQualityItemNames...),
		BodyLocations:              append([]models.BodyLocation(nil), source.BodyLocations...),
		BodyLocationsByCode:        make(map[models.ItemBodyLocation]models.BodyLocation, len(source.BodyLocationsByCode)),
		StorePages:                 append([]models.StorePage(nil), source.StorePages...),
		StorePagesByCode:           make(map[models.ItemStorePage]models.StorePage, len(source.StorePagesByCode)),
		CompositeComponents:        append([]models.CompositeComponent(nil), source.CompositeComponents...),
		CompositeComponentsByToken: make(map[string]models.CompositeComponent, len(source.CompositeComponentsByToken)),
		HitClasses:                 append([]models.HitClass(nil), source.HitClasses...),
		HitClassesByCode:           make(map[string]models.HitClass, len(source.HitClassesByCode)),
		PlayerClasses:              append([]models.PlayerClassCode(nil), source.PlayerClasses...),
		PlayerClassesByCode:        make(map[string]models.PlayerClassCode, len(source.PlayerClassesByCode)),
		PlayerModes:                append([]models.PlayerMode(nil), source.PlayerModes...),
		PlayerModesByToken:         make(map[string]models.PlayerMode, len(source.PlayerModesByToken)),
		PlayerTypes:                append([]models.PlayerType(nil), source.PlayerTypes...),
		PlayerTypesByToken:         make(map[string]models.PlayerType, len(source.PlayerTypesByToken)),
		MonsterModes:               append([]models.MonsterMode(nil), source.MonsterModes...),
		MonsterModesByToken:        make(map[string]models.MonsterMode, len(source.MonsterModesByToken)),
		MonsterPlaces:              append([]models.MonsterPlace(nil), source.MonsterPlaces...),
		TransformColors:            append([]models.TransformColor(nil), source.TransformColors...),
		TransformColorsByCode:      make(map[models.ColorCode]models.TransformColor, len(source.TransformColorsByCode)),
		ComponentCodes:             append([]models.ComponentCode(nil), source.ComponentCodes...),
		ComponentCodesByCode:       make(map[string]models.ComponentCode, len(source.ComponentCodesByCode)),
		ElementTypes:               append([]models.ElementType(nil), source.ElementTypes...),
		ElementTypesByCode:         make(map[string]models.ElementType, len(source.ElementTypesByCode)),
		Events:                     append([]models.UnitEvent(nil), source.Events...),
		EventsByName:               make(map[string]models.UnitEvent, len(source.EventsByName)),
		MissileCalculations:        append([]models.MissileCalculation(nil), source.MissileCalculations...),
		MissileCalculationsByCode:  make(map[string]models.MissileCalculation, len(source.MissileCalculationsByCode)),
		SkillCalculations:          append([]models.SkillCalculation(nil), source.SkillCalculations...),
		SkillCalculationsByCode:    make(map[string]models.SkillCalculation, len(source.SkillCalculationsByCode)),
		ArmorTypes:                 append([]models.ArmorType(nil), source.ArmorTypes...),
		ArmorTypesByToken:          make(map[string]models.ArmorType, len(source.ArmorTypesByToken)),
		CubeModifierTypes:          append([]models.CubeModifierType(nil), source.CubeModifierTypes...),
		CubeModifierTypesByCode:    make(map[string]models.CubeModifierType, len(source.CubeModifierTypesByCode)),
	}
	for key, value := range source.CharStatsByClass {
		result.CharStatsByClass[key] = value
	}
	for key, value := range source.LevelsByID {
		result.LevelsByID[key] = value
	}
	for key, value := range source.ObjectsByClass {
		result.ObjectsByClass[key] = value
	}
	for key, value := range source.SkillsByID {
		result.SkillsByID[key] = value
	}
	for key, value := range source.SoundsByName {
		result.SoundsByName[key] = value
	}
	for key, value := range source.TreasureByName {
		result.TreasureByName[key] = value
	}
	for key, value := range source.ArmorByCode {
		result.ArmorByCode[key] = value
	}
	for key, value := range source.WeaponsByCode {
		result.WeaponsByCode[key] = value
	}
	for key, value := range source.MiscByCode {
		result.MiscByCode[key] = value
	}
	for key, value := range source.ItemTypesByCode {
		result.ItemTypesByCode[key] = value
	}
	for key, value := range source.ItemStatsByName {
		result.ItemStatsByName[key] = value
	}
	for key, value := range source.PropertiesByCode {
		result.PropertiesByCode[key] = value
	}
	for key, value := range source.UniqueByIndex {
		result.UniqueByIndex[key] = value
	}
	for key, value := range source.SetItemsByIndex {
		result.SetItemsByIndex[key] = value
	}
	for key, value := range source.GemsByCode {
		result.GemsByCode[key] = value
	}
	for key, value := range source.SetsByIndex {
		result.SetsByIndex[key] = value
	}
	for key, value := range source.LevelPresetByDef {
		result.LevelPresetByDef[key] = value
	}
	for key, value := range source.LevelMazeByLevel {
		result.LevelMazeByLevel[key] = value
	}
	for key, value := range source.MonstersByID {
		result.MonstersByID[key] = value
	}
	for key, value := range source.MonsterGfxByID {
		result.MonsterGfxByID[key] = value
	}
	for key, value := range source.MonsterPropsByID {
		result.MonsterPropsByID[key] = value
	}
	for key, value := range source.MonsterSoundByID {
		result.MonsterSoundByID[key] = value
	}
	for key, value := range source.MissilesByName {
		result.MissilesByName[key] = value
	}
	for key, value := range source.StatesByName {
		result.StatesByName[key] = value
	}
	for key, value := range source.OverlaysByName {
		result.OverlaysByName[key] = value
	}
	for key, value := range source.InventoryByClass {
		result.InventoryByClass[key] = value
	}
	for key, value := range source.BeltsByName {
		result.BeltsByName[key] = value
	}
	for key, value := range source.DifficultyByName {
		result.DifficultyByName[key] = value
	}
	for key, value := range source.SkillDescByName {
		result.SkillDescByName[key] = value
	}
	for key, value := range source.SoundEnvByHandle {
		result.SoundEnvByHandle[key] = value
	}
	for key, value := range source.NPCTradesByID {
		result.NPCTradesByID[key] = value
	}
	for key, value := range source.ShrinesByType {
		result.ShrinesByType[key] = value
	}
	for key, value := range source.GambleItemsByCode {
		result.GambleItemsByCode[key] = value
	}
	for key, value := range source.ObjectTypesByName {
		result.ObjectTypesByName[key] = value
	}
	for key, value := range source.ObjectGroupsByName {
		result.ObjectGroupsByName[key] = value
	}
	for key, value := range source.ObjectModesByName {
		result.ObjectModesByName[key] = value
	}
	for key, value := range source.WeaponClassByCode {
		result.WeaponClassByCode[key] = value
	}
	for key, value := range source.BooksByName {
		result.BooksByName[key] = value
	}
	for key, value := range source.MonsterUniqueByID {
		result.MonsterUniqueByID[key] = value
	}
	for key, value := range source.ClassicTreasureByName {
		result.ClassicTreasureByName[key] = value
	}
	for key, value := range source.HirelingDescriptionByID {
		result.HirelingDescriptionByID[key] = value
	}
	for key, value := range source.SuperUniquesByID {
		result.SuperUniquesByID[key] = value
	}
	for key, value := range source.SuperUniquesByHardcodedID {
		result.SuperUniquesByHardcodedID[key] = value
	}
	for key, value := range source.BodyLocationsByCode {
		result.BodyLocationsByCode[key] = value
	}
	for key, value := range source.StorePagesByCode {
		result.StorePagesByCode[key] = value
	}
	for key, value := range source.CompositeComponentsByToken {
		result.CompositeComponentsByToken[key] = value
	}
	for key, value := range source.HitClassesByCode {
		result.HitClassesByCode[key] = value
	}
	for key, value := range source.PlayerClassesByCode {
		result.PlayerClassesByCode[key] = value
	}
	for key, value := range source.PlayerModesByToken {
		result.PlayerModesByToken[key] = value
	}
	for key, value := range source.PlayerTypesByToken {
		result.PlayerTypesByToken[key] = value
	}
	for key, value := range source.MonsterModesByToken {
		result.MonsterModesByToken[key] = value
	}
	for key, value := range source.TransformColorsByCode {
		result.TransformColorsByCode[key] = value
	}
	for key, value := range source.ComponentCodesByCode {
		result.ComponentCodesByCode[key] = value
	}
	for key, value := range source.ElementTypesByCode {
		result.ElementTypesByCode[key] = value
	}
	for key, value := range source.EventsByName {
		result.EventsByName[key] = value
	}
	for key, value := range source.MissileCalculationsByCode {
		result.MissileCalculationsByCode[key] = value
	}
	for key, value := range source.SkillCalculationsByCode {
		result.SkillCalculationsByCode[key] = value
	}
	for key, value := range source.ArmorTypesByToken {
		result.ArmorTypesByToken[key] = value
	}
	for key, value := range source.CubeModifierTypesByCode {
		result.CubeModifierTypesByCode[key] = value
	}
	return result
}
