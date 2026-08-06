package gamedata

import (
	"fmt"
	"sync"

	"github.com/gravestench/dark-magic/internal/recordstore"
	"github.com/gravestench/dark-magic/pkg/models"
)

// Snapshot is an immutable, internally consistent view of admitted typed game
// tables. New tables are added here only after their surviving schema passes
// synthetic and real-archive verification.
type Snapshot struct {
	Issues           []Issue
	CharStats        []models.CharStats
	CharStatsByClass map[string]models.CharStats
	Levels           []models.LevelData
	LevelsByID       map[int]models.LevelData
	Objects          []models.Object
	ObjectsByClass   map[int]models.Object
	Skills           []models.SkillData
	SkillsByID       map[string]models.SkillData
	Sounds           []models.SoundEntry
	SoundsByName     map[string]models.SoundEntry
	TreasureClasses  []models.TreasureClassEx
	TreasureByName   map[string]models.TreasureClassEx
	Armor            []models.ItemArmor
	ArmorByCode      map[string]models.ItemArmor
	Weapons          []models.ItemWeapon
	WeaponsByCode    map[string]models.ItemWeapon
	Misc             []models.MiscItem
	MiscByCode       map[string]models.MiscItem
	ItemTypes        []models.ItemType
	ItemTypesByCode  map[string]models.ItemType
	ItemRatios       []models.ItemRatio
	ItemStats        []models.ItemStatCost
	ItemStatsByName  map[string]models.ItemStatCost
	Properties       []models.ItemProperty
	PropertiesByCode map[string]models.ItemProperty
	UniqueItems      []models.ItemUnique
	UniqueByIndex    map[string]models.ItemUnique
	SetItems         []models.SetItemData
	SetItemsByIndex  map[string]models.SetItemData
	MagicPrefixes    []models.MagicPrefix
	MagicSuffixes    []models.MagicSuffix
	AutoMagic        []models.AutoMagicData
	RarePrefixes     []models.RarePrefix
	RareSuffixes     []models.RareSuffix
	Gems             []models.GemData
	GemsByCode       map[string]models.GemData
	RuneWords        []models.RuneWordData
	CubeRecipes      []models.CubeRecipe
	Sets             []models.SetBonusData
	SetsByIndex      map[string]models.SetBonusData
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
	if path != CharStatsTable && path != LevelsTable && path != ObjectsTable && path != SkillsTable && path != SoundsTable && path != TreasureClassExTable && path != ArmorTable && path != WeaponsTable && path != MiscTable && path != ItemTypesTable && path != ItemRatioTable && path != ItemStatCostTable && path != PropertiesTable && path != UniqueItemsTable && path != SetItemsTable && path != MagicPrefixTable && path != MagicSuffixTable && path != AutoMagicTable && path != RarePrefixTable && path != RareSuffixTable && path != GemsTable && path != RunesTable && path != CubeMainTable && path != SetsTable {
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
	}, nil
}

func cloneSnapshot(source Snapshot) Snapshot {
	result := Snapshot{
		Issues:           append([]Issue(nil), source.Issues...),
		CharStats:        append([]models.CharStats(nil), source.CharStats...),
		CharStatsByClass: make(map[string]models.CharStats, len(source.CharStatsByClass)),
		Levels:           append([]models.LevelData(nil), source.Levels...),
		LevelsByID:       make(map[int]models.LevelData, len(source.LevelsByID)),
		Objects:          append([]models.Object(nil), source.Objects...),
		ObjectsByClass:   make(map[int]models.Object, len(source.ObjectsByClass)),
		Skills:           append([]models.SkillData(nil), source.Skills...),
		SkillsByID:       make(map[string]models.SkillData, len(source.SkillsByID)),
		Sounds:           append([]models.SoundEntry(nil), source.Sounds...),
		SoundsByName:     make(map[string]models.SoundEntry, len(source.SoundsByName)),
		TreasureClasses:  append([]models.TreasureClassEx(nil), source.TreasureClasses...),
		TreasureByName:   make(map[string]models.TreasureClassEx, len(source.TreasureByName)),
		Armor:            append([]models.ItemArmor(nil), source.Armor...),
		ArmorByCode:      make(map[string]models.ItemArmor, len(source.ArmorByCode)),
		Weapons:          append([]models.ItemWeapon(nil), source.Weapons...),
		WeaponsByCode:    make(map[string]models.ItemWeapon, len(source.WeaponsByCode)),
		Misc:             append([]models.MiscItem(nil), source.Misc...),
		MiscByCode:       make(map[string]models.MiscItem, len(source.MiscByCode)),
		ItemTypes:        append([]models.ItemType(nil), source.ItemTypes...),
		ItemTypesByCode:  make(map[string]models.ItemType, len(source.ItemTypesByCode)),
		ItemRatios:       append([]models.ItemRatio(nil), source.ItemRatios...),
		ItemStats:        append([]models.ItemStatCost(nil), source.ItemStats...),
		ItemStatsByName:  make(map[string]models.ItemStatCost, len(source.ItemStatsByName)),
		Properties:       append([]models.ItemProperty(nil), source.Properties...),
		PropertiesByCode: make(map[string]models.ItemProperty, len(source.PropertiesByCode)),
		UniqueItems:      append([]models.ItemUnique(nil), source.UniqueItems...),
		UniqueByIndex:    make(map[string]models.ItemUnique, len(source.UniqueByIndex)),
		SetItems:         append([]models.SetItemData(nil), source.SetItems...),
		SetItemsByIndex:  make(map[string]models.SetItemData, len(source.SetItemsByIndex)),
		MagicPrefixes:    append([]models.MagicPrefix(nil), source.MagicPrefixes...),
		MagicSuffixes:    append([]models.MagicSuffix(nil), source.MagicSuffixes...),
		AutoMagic:        append([]models.AutoMagicData(nil), source.AutoMagic...),
		RarePrefixes:     append([]models.RarePrefix(nil), source.RarePrefixes...),
		RareSuffixes:     append([]models.RareSuffix(nil), source.RareSuffixes...),
		Gems:             append([]models.GemData(nil), source.Gems...),
		GemsByCode:       make(map[string]models.GemData, len(source.GemsByCode)),
		RuneWords:        append([]models.RuneWordData(nil), source.RuneWords...),
		CubeRecipes:      append([]models.CubeRecipe(nil), source.CubeRecipes...),
		Sets:             append([]models.SetBonusData(nil), source.Sets...),
		SetsByIndex:      make(map[string]models.SetBonusData, len(source.SetsByIndex)),
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
	return result
}
