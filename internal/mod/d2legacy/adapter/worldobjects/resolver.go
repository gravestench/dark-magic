// Package worldobjects joins Diablo DS1 act-local identities to recovered and
// decoded d2legacy records.
package worldobjects

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
)

// Records supplies generation-pinned table rows used to resolve dynamic object identities.
type Records interface {
	LoadRows(string) ([]recordstore.Row, error)
}

// StaticDefinition keeps the DS1 act-local identity separate from the authored
// global Objects.txt numeric ID and the derived lookup key.
type StaticDefinition struct {
	ActLocalID  recordstore.ActLocalID
	ObjectID    recordstore.NumericID
	Description string
	RowOrdinal  recordstore.RowOrdinal
	Source      recordstore.Provenance
	IndexKey    recordstore.DerivedIndex
}

// DynamicDefinition keeps MonPreset's global source row separate from its
// positional per-act DS1 ID and authored monster code.
type DynamicDefinition struct {
	ActLocalID recordstore.ActLocalID
	Place      recordstore.SymbolicID
	RowOrdinal recordstore.RowOrdinal
	Source     recordstore.Provenance
	IndexKey   recordstore.DerivedIndex
}

// Resolver is an immutable O(1) lookup built once per game-data generation.
type Resolver struct {
	mu      sync.RWMutex
	static  map[recordstore.DerivedIndex]StaticDefinition
	dynamic map[recordstore.DerivedIndex]DynamicDefinition
}

// New freezes both data generations into act-local lookup indexes. Dynamic
// monster preset IDs are positional, so their authored per-act order is kept.
func New(recoveredData recovered.Snapshot, records Records) (*Resolver, error) {
	resolver := &Resolver{}
	if err := resolver.Update(recoveredData, records); err != nil {
		return nil, err
	}

	return resolver, nil
}

// Update atomically replaces lookup indexes after the mounted package recipe
// changes while preserving the pointer captured by Lua/native capabilities.
func (resolver *Resolver) Update(recoveredData recovered.Snapshot, records Records) error {
	presets, err := records.LoadRows("data/global/excel/monpreset.txt")
	if err != nil {
		return fmt.Errorf("d2legacy world objects: load MonPreset.txt: %w", err)
	}

	static := indexStaticDefinitions(recoveredData.MapObjects)

	dynamic, err := indexDynamicDefinitions(presets)
	if err != nil {
		return err
	}

	// Build both indexes before taking the lock so readers never observe a partial data-generation replacement.
	resolver.mu.Lock()
	resolver.static = static
	resolver.dynamic = dynamic
	resolver.mu.Unlock()

	return nil
}

// indexStaticDefinitions projects recovered DS1 identities into the act-local key space used by map objects.
func indexStaticDefinitions(entries []recovered.MapObject) map[recordstore.DerivedIndex]StaticDefinition {
	definitions := make(map[recordstore.DerivedIndex]StaticDefinition, len(entries))

	for _, entry := range entries {
		actLocalID := recordstore.ActLocalID{Act: entry.Act, Index: entry.ID}
		indexKey := key(actLocalID)
		definitions[indexKey] = StaticDefinition{
			ActLocalID:  actLocalID,
			ObjectID:    recordstore.NumericID(entry.ObjectID),
			Description: entry.Description,
			RowOrdinal:  entry.RowOrdinal,
			Source:      entry.Source,
			IndexKey:    indexKey,
		}
	}

	return definitions
}

// indexDynamicDefinitions preserves authored per-act row order because DS1 dynamic IDs are positional indexes.
func indexDynamicDefinitions(rows []recordstore.Row) (map[recordstore.DerivedIndex]DynamicDefinition, error) {
	definitions := make(map[recordstore.DerivedIndex]DynamicDefinition, len(rows))
	actOffsets := make(map[int]int, 5)

	for _, row := range rows {
		act, err := strconv.Atoi(row.Values["Act"])
		if err != nil {
			return nil, fmt.Errorf(
				"d2legacy world objects: MonPreset.txt row %d has invalid Act %q",
				int(row.Ordinal)+2,
				row.Values["Act"],
			)
		}

		actLocalID := recordstore.ActLocalID{Act: act, Index: actOffsets[act]}
		indexKey := key(actLocalID)
		definitions[indexKey] = DynamicDefinition{
			ActLocalID: actLocalID,
			Place:      recordstore.SymbolicID(row.Values["Place"]),
			RowOrdinal: row.Ordinal,
			Source:     row.Provenance,
			IndexKey:   indexKey,
		}
		actOffsets[act]++
	}

	return definitions, nil
}

// ResolveStaticObject maps a DS1 static ID to Objects.txt and its recovered
// human-readable description.
func (resolver *Resolver) ResolveStaticObject(act, id int) (int, string, bool) {
	entry, found := resolver.StaticDefinition(act, id)
	return int(entry.ObjectID), entry.Description, found
}

// StaticDefinition exposes the authored numeric target and source coordinates
// without conflating either with the DS1 act-local lookup identity.
func (resolver *Resolver) StaticDefinition(act, id int) (StaticDefinition, bool) {
	if resolver == nil {
		return StaticDefinition{}, false
	}

	resolver.mu.RLock()
	defer resolver.mu.RUnlock()

	entry, found := resolver.static[key(recordstore.ActLocalID{Act: act, Index: id})]

	return entry, found
}

// ResolveDynamicObject maps a DS1 dynamic ID to the act-local MonPreset place
// token. It returns identity only; spawning remains authoritative game logic.
func (resolver *Resolver) ResolveDynamicObject(act, id int) (string, bool) {
	entry, found := resolver.DynamicDefinition(act, id)
	return string(entry.Place), found
}

// DynamicDefinition exposes the source ordinal, authored monster code, and
// per-act positional ID as separate values for diagnostics and consumers.
func (resolver *Resolver) DynamicDefinition(act, id int) (DynamicDefinition, bool) {
	if resolver == nil {
		return DynamicDefinition{}, false
	}

	resolver.mu.RLock()
	defer resolver.mu.RUnlock()

	entry, found := resolver.dynamic[key(recordstore.ActLocalID{Act: act, Index: id})]

	return entry, found
}

// key keeps identical numeric IDs from different acts in independent lookup namespaces.
func key(id recordstore.ActLocalID) recordstore.DerivedIndex {
	return recordstore.DerivedIndex(fmt.Sprintf("%d:%d", id.Act, id.Index))
}
