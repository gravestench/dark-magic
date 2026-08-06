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
	if path != CharStatsTable && path != LevelsTable && path != ObjectsTable && path != SkillsTable && path != SoundsTable && path != TreasureClassExTable {
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
	return Snapshot{
		Issues:    issues,
		CharStats: characters, CharStatsByClass: charactersByClass,
		Levels: levels, LevelsByID: levelsByID,
		Objects: objects, ObjectsByClass: objectsByClass,
		Skills: skills, SkillsByID: skillsByID,
		Sounds: sounds, SoundsByName: soundsByName,
		TreasureClasses: treasure, TreasureByName: treasureByName,
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
	return result
}
