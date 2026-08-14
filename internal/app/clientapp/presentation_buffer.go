package clientapp

import (
	"sort"

	"github.com/gravestench/dark-magic/internal/app/networkclock"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const (
	presentationSnapshotCapacity      = 32
	presentationMaxExtrapolationTicks = 2.5
)

type bufferedWorldView struct {
	tick     uint64
	revision uint64
	entities map[string]playeradapter.WorldEntity
}

type sampledWorldView struct {
	moment           networkclock.Moment
	discreteTick     uint64
	discreteRevision uint64
	entities         []playeradapter.WorldEntity
	extrapolated     bool
}

// presentationBuffer retains immutable authoritative public-world samples.
// The renderer asks it for a past network-clock moment and receives values
// between two known snapshots instead of chasing the newest packet.
type presentationBuffer struct {
	capacity  int
	revision  uint64
	snapshots []bufferedWorldView
}

func newPresentationBuffer() *presentationBuffer {
	return &presentationBuffer{capacity: presentationSnapshotCapacity}
}

func (buffer *presentationBuffer) Push(view playeradapter.WorldView) bool {
	if buffer == nil || (len(buffer.snapshots) > 0 && view.Tick <= buffer.snapshots[len(buffer.snapshots)-1].tick) {
		return false
	}
	return buffer.append(view)
}

// Upsert replaces metadata for the newest tick or appends a newer sample.
// Older packets cannot rewrite history.
func (buffer *presentationBuffer) Upsert(view playeradapter.WorldView) bool {
	if buffer == nil {
		return false
	}
	if len(buffer.snapshots) == 0 || view.Tick > buffer.snapshots[len(buffer.snapshots)-1].tick {
		return buffer.append(view)
	}
	if view.Tick < buffer.snapshots[len(buffer.snapshots)-1].tick {
		return false
	}
	buffer.revision++
	buffer.snapshots[len(buffer.snapshots)-1] = buffer.snapshot(view)
	return true
}

func (buffer *presentationBuffer) append(view playeradapter.WorldView) bool {
	buffer.revision++
	buffer.snapshots = append(buffer.snapshots, buffer.snapshot(view))
	if len(buffer.snapshots) > buffer.capacity {
		copy(buffer.snapshots, buffer.snapshots[len(buffer.snapshots)-buffer.capacity:])
		buffer.snapshots = buffer.snapshots[:buffer.capacity]
	}
	return true
}

func (buffer *presentationBuffer) snapshot(view playeradapter.WorldView) bufferedWorldView {
	entities := make(map[string]playeradapter.WorldEntity, len(view.Entities))
	for _, entity := range view.Entities {
		entities[entity.ID] = cloneWorldEntity(entity)
	}
	return bufferedWorldView{tick: view.Tick, revision: buffer.revision, entities: entities}
}

func (buffer *presentationBuffer) Sample(moment networkclock.Moment) (sampledWorldView, bool) {
	if buffer == nil || len(buffer.snapshots) == 0 {
		return sampledWorldView{}, false
	}
	renderTick := float64(moment.Tick) + moment.Fraction
	upper := sort.Search(len(buffer.snapshots), func(index int) bool {
		return float64(buffer.snapshots[index].tick) > renderTick
	})
	if upper == 0 {
		return sampleSnapshot(moment, buffer.snapshots[0]), true
	}
	if upper < len(buffer.snapshots) {
		lower := buffer.snapshots[upper-1]
		next := buffer.snapshots[upper]
		alpha := (renderTick - float64(lower.tick)) / float64(next.tick-lower.tick)
		return sampleBetween(moment, lower, next, alpha), true
	}
	latest := buffer.snapshots[len(buffer.snapshots)-1]
	if len(buffer.snapshots) == 1 || renderTick <= float64(latest.tick) {
		return sampleSnapshot(moment, latest), true
	}
	previous := buffer.snapshots[len(buffer.snapshots)-2]
	ahead := min(renderTick-float64(latest.tick), presentationMaxExtrapolationTicks)
	return extrapolateSnapshot(moment, previous, latest, ahead), true
}

func sampleSnapshot(moment networkclock.Moment, snapshot bufferedWorldView) sampledWorldView {
	return sampledWorldView{moment: moment, discreteTick: snapshot.tick, discreteRevision: snapshot.revision, entities: sortedEntities(snapshot.entities)}
}

func sampleBetween(moment networkclock.Moment, lower, upper bufferedWorldView, alpha float64) sampledWorldView {
	entities := make(map[string]playeradapter.WorldEntity, len(lower.entities))
	for id, entity := range lower.entities {
		value := cloneWorldEntity(entity)
		if next, found := upper.entities[id]; found {
			value.Position.X += (next.Position.X - value.Position.X) * alpha
			value.Position.Y += (next.Position.Y - value.Position.Y) * alpha
		}
		entities[id] = value
	}
	return sampledWorldView{moment: moment, discreteTick: lower.tick, discreteRevision: lower.revision, entities: sortedEntities(entities)}
}

func extrapolateSnapshot(moment networkclock.Moment, previous, latest bufferedWorldView, ahead float64) sampledWorldView {
	span := float64(latest.tick - previous.tick)
	entities := make(map[string]playeradapter.WorldEntity, len(latest.entities))
	for id, entity := range latest.entities {
		value := cloneWorldEntity(entity)
		if before, found := previous.entities[id]; found && span > 0 {
			value.Position.X += (value.Position.X - before.Position.X) * ahead / span
			value.Position.Y += (value.Position.Y - before.Position.Y) * ahead / span
		}
		entities[id] = value
	}
	return sampledWorldView{moment: moment, discreteTick: latest.tick, discreteRevision: latest.revision, entities: sortedEntities(entities), extrapolated: ahead > 0}
}

func sortedEntities(values map[string]playeradapter.WorldEntity) []playeradapter.WorldEntity {
	result := make([]playeradapter.WorldEntity, 0, len(values))
	for _, entity := range values {
		result = append(result, cloneWorldEntity(entity))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func cloneWorldEntity(entity playeradapter.WorldEntity) playeradapter.WorldEntity {
	if entity.Health != nil {
		value := *entity.Health
		entity.Health = &value
	}
	if entity.MaxHealth != nil {
		value := *entity.MaxHealth
		entity.MaxHealth = &value
	}
	return entity
}
