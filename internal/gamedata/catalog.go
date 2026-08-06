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
	CharStats        []models.CharStats
	CharStatsByClass map[string]models.CharStats
	Sounds           []models.SoundEntry
	SoundsByName     map[string]models.SoundEntry
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
	if path != CharStatsTable && path != SoundsTable {
		return
	}
	c.mu.Lock()
	c.data = nil
	c.mu.Unlock()
}

func (c *Catalog) load() (Snapshot, error) {
	characters, err := Load[models.CharStats](c.store, CharStatsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	charactersByClass, err := Index(characters, func(record models.CharStats) string { return record.Class })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index charstats: %w", err)
	}
	sounds, err := Load[models.SoundEntry](c.store, SoundsTable)
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: build catalog: %w", err)
	}
	soundsByName, err := Index(sounds, func(record models.SoundEntry) string { return record.Sound })
	if err != nil {
		return Snapshot{}, fmt.Errorf("gamedata: index sounds: %w", err)
	}
	return Snapshot{CharStats: characters, CharStatsByClass: charactersByClass, Sounds: sounds, SoundsByName: soundsByName}, nil
}

func cloneSnapshot(source Snapshot) Snapshot {
	result := Snapshot{
		CharStats:        append([]models.CharStats(nil), source.CharStats...),
		CharStatsByClass: make(map[string]models.CharStats, len(source.CharStatsByClass)),
		Sounds:           append([]models.SoundEntry(nil), source.Sounds...),
		SoundsByName:     make(map[string]models.SoundEntry, len(source.SoundsByName)),
	}
	for key, value := range source.CharStatsByClass {
		result.CharStatsByClass[key] = value
	}
	for key, value := range source.SoundsByName {
		result.SoundsByName[key] = value
	}
	return result
}
