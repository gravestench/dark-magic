package player

import (
	"reflect"
	"sort"
)

const WorldDeltaVersion uint32 = 1

type WorldDelta struct {
	Version  uint32        `json:"version"`
	BaseTick uint64        `json:"base_tick"`
	Tick     uint64        `json:"tick"`
	Reset    bool          `json:"reset"`
	Upserts  []WorldEntity `json:"upserts"`
	Removed  []string      `json:"removed"`
}

// DiffWorldView creates an idempotent, deterministic delta. A truncated base
// cannot prove removals, so either truncated side produces a reset containing
// the complete currently visible bounded set.
func DiffWorldView(previous, current WorldView) WorldDelta {
	delta := WorldDelta{Version: WorldDeltaVersion, BaseTick: previous.Tick, Tick: current.Tick, Upserts: []WorldEntity{}, Removed: []string{}}
	if previous.Truncated || current.Truncated {
		delta.Reset = true
		delta.Upserts = append(delta.Upserts, current.Entities...)
		return delta
	}
	before := make(map[string]WorldEntity, len(previous.Entities))
	for _, entity := range previous.Entities {
		before[entity.ID] = entity
	}
	after := make(map[string]struct{}, len(current.Entities))
	for _, entity := range current.Entities {
		after[entity.ID] = struct{}{}
		if old, found := before[entity.ID]; !found || !reflect.DeepEqual(publicEntity(old), publicEntity(entity)) {
			delta.Upserts = append(delta.Upserts, entity)
		}
	}
	for _, entity := range previous.Entities {
		if _, found := after[entity.ID]; !found {
			delta.Removed = append(delta.Removed, entity.ID)
		}
	}
	sort.Strings(delta.Removed)
	return delta
}

func publicEntity(entity WorldEntity) WorldEntity { entity.distance2 = 0; return entity }
