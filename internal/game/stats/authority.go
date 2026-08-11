package stats

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"
)

// Authority owns active stat sources for every target entity. Gameplay systems
// change their own facts first, then atomically publish the matching sources.
// Readers never receive the live maps or slices.
type Authority struct {
	mu       sync.RWMutex
	entities map[EntityID]*entityState
}

type entityState struct {
	revision uint64
	sources  map[SourceID]Source
}

// NewAuthority constructs an empty source owner.
func NewAuthority() *Authority {
	return &Authority{entities: make(map[EntityID]*entityState)}
}

// Apply validates and commits one mutation. A repeated, byte-for-byte-equivalent
// replacement is a no-op, which makes equipment/state refresh paths idempotent.
// Removing an absent source is also a no-op.
func (authority *Authority) Apply(entity EntityID, mutation Mutation) (uint64, error) {
	if authority == nil {
		return 0, fmt.Errorf("stats: authority is required")
	}
	if entity == "" {
		return 0, fmt.Errorf("stats: entity ID is required")
	}
	replacements := make(map[SourceID]Source, len(mutation.Replace))
	for _, source := range mutation.Replace {
		if err := source.validate(); err != nil {
			return 0, err
		}
		if _, duplicate := replacements[source.ID]; duplicate {
			return 0, fmt.Errorf("stats: source %q is replaced more than once", source.ID)
		}
		replacements[source.ID] = source.Clone()
	}
	removals := make(map[SourceID]struct{}, len(mutation.Remove))
	for _, id := range mutation.Remove {
		if id == "" {
			return 0, fmt.Errorf("stats: removal source ID is required")
		}
		if _, replaced := replacements[id]; replaced {
			return 0, fmt.Errorf("stats: source %q cannot be replaced and removed together", id)
		}
		if _, duplicate := removals[id]; duplicate {
			return 0, fmt.Errorf("stats: source %q is removed more than once", id)
		}
		removals[id] = struct{}{}
	}

	authority.mu.Lock()
	defer authority.mu.Unlock()
	if authority.entities == nil {
		authority.entities = make(map[EntityID]*entityState)
	}
	current := authority.entities[entity]
	if current == nil {
		current = &entityState{sources: make(map[SourceID]Source)}
	}
	next := cloneSourceMap(current.sources)
	changed := false
	for id, source := range replacements {
		if previous, exists := next[id]; !exists || !reflect.DeepEqual(previous, source) {
			next[id] = source
			changed = true
		}
	}
	for id := range removals {
		if _, exists := next[id]; exists {
			delete(next, id)
			changed = true
		}
	}
	if !changed {
		return current.revision, nil
	}
	current = &entityState{revision: current.revision + 1, sources: next}
	authority.entities[entity] = current
	return current.revision, nil
}

// Effective adds every logical contribution for one full StatKey. Derived
// ItemStatCost operations, caps, and display rounding intentionally do not live
// here yet. Overflow is rejected instead of silently wrapping combat state.
func (authority *Authority) Effective(entity EntityID, key Key) (int64, error) {
	if authority == nil {
		return 0, fmt.Errorf("stats: authority is required")
	}
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	state := authority.entities[entity]
	if state == nil {
		return 0, nil
	}
	var total int64
	for _, id := range sortedSourceIDs(state.sources) {
		source := state.sources[id]
		for _, entry := range source.Entries {
			if entry.Key != key {
				continue
			}
			if entry.Value > 0 && total > math.MaxInt64-entry.Value || entry.Value < 0 && total < math.MinInt64-entry.Value {
				return 0, fmt.Errorf("stats: value overflow for stat %d parameter %d", key.ID, key.Parameter)
			}
			total += entry.Value
		}
	}
	return total, nil
}

// Snapshot returns sources in SourceID order. This makes equal state serialize
// identically even though Go deliberately randomizes map iteration.
func (authority *Authority) Snapshot(entity EntityID) Snapshot {
	if authority == nil {
		return Snapshot{Entity: entity}
	}
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	result := Snapshot{Entity: entity}
	state := authority.entities[entity]
	if state == nil {
		return result
	}
	return snapshotLocked(entity, state)
}

func snapshotLocked(entity EntityID, state *entityState) Snapshot {
	result := Snapshot{Entity: entity, Revision: state.revision}
	for _, id := range sortedSourceIDs(state.sources) {
		result.Sources = append(result.Sources, state.sources[id].Clone())
	}
	return result
}

func sortedSourceIDs(sources map[SourceID]Source) []SourceID {
	ids := make([]SourceID, 0, len(sources))
	for id := range sources {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func cloneSourceMap(source map[SourceID]Source) map[SourceID]Source {
	clone := make(map[SourceID]Source, len(source))
	for id, value := range source {
		clone[id] = value.Clone()
	}
	return clone
}
