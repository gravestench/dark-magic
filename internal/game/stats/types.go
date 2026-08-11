// Package stats owns the small, deterministic vocabulary used to explain why
// an authoritative entity has a value.
//
// A character's strength is not just one number. Part of it may come from the
// character, part from boots, and part from a temporary shrine. Keeping those
// pieces separate means removing the boots cannot accidentally remove the
// shrine. This package stores those pieces; legacy formulas are layered on top
// only after their arithmetic has executable evidence.
package stats

import "fmt"

// EntityID identifies the entity whose effective values are being resolved.
// It is deliberately protocol-neutral: an ECS adapter may use a stable entity
// identity while a headless test can use a readable fixture name.
type EntityID string

// StatID is the numeric ItemStatCost row identity admitted by a pinned content
// generation. The numeric identity and Parameter together form the real key.
type StatID uint16

// Key distinguishes parameterized variants of the same stat. Skill bonuses are
// a common reason that retaining only StatID would lose information.
type Key struct {
	ID        StatID `json:"id"`
	Parameter int32  `json:"parameter,omitempty"`
}

// SourceID is stable within one target entity. Reusing an ID replaces that
// source instead of applying its values twice.
type SourceID string

// SourceKind says what kind of rule supplied a source. It is semantic data,
// not a switch that performs gameplay by itself.
type SourceKind string

const (
	SourceBase        SourceKind = "base"
	SourceItem        SourceKind = "item"
	SourceSocket      SourceKind = "socket"
	SourceSet         SourceKind = "set"
	SourceCharm       SourceKind = "charm"
	SourcePassive     SourceKind = "passive"
	SourceAura        SourceKind = "aura"
	SourceState       SourceKind = "state"
	SourceMonster     SourceKind = "monster_modifier"
	SourceQuest       SourceKind = "quest"
	SourceDifficulty  SourceKind = "difficulty"
	SourceEnvironment SourceKind = "environment"
)

// Lifetime documents which owner is expected to detach a source. The stats
// package does not read clocks or inventories; those authoritative systems
// explicitly replace or remove their sources when their own state changes.
type Lifetime string

const (
	LifetimeDurable   Lifetime = "durable"
	LifetimeEquipped  Lifetime = "equipped"
	LifetimeSession   Lifetime = "session"
	LifetimeTimed     Lifetime = "timed"
	LifetimeProximity Lifetime = "proximity"
)

// OwnerRef identifies the fact that produced a source, such as an item, skill,
// state instance, or monster modifier. It is separate from the target EntityID.
type OwnerRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// Entry is one decoded logical contribution. Value is not a save bit pattern,
// displayed number, or already-derived effective value.
type Entry struct {
	Key   Key   `json:"key"`
	Value int64 `json:"value"`
}

// Source groups ordered entries that must be attached and detached together.
// StartTick and EndTick are descriptive, checkpointed facts for timed owners;
// a scheduler remains responsible for expiration.
type Source struct {
	ID        SourceID   `json:"id"`
	Kind      SourceKind `json:"kind"`
	Lifetime  Lifetime   `json:"lifetime"`
	Owner     OwnerRef   `json:"owner"`
	StartTick uint64     `json:"start_tick,omitempty"`
	EndTick   uint64     `json:"end_tick,omitempty"`
	Entries   []Entry    `json:"entries"`
}

// Clone returns a source that cannot share its mutable entry slice with the
// authority. Callers may safely reuse or edit their input after attachment.
func (source Source) Clone() Source {
	clone := source
	clone.Entries = append([]Entry(nil), source.Entries...)
	return clone
}

func (source Source) validate() error {
	if source.ID == "" {
		return fmt.Errorf("stats: source ID is required")
	}
	if source.Kind == "" {
		return fmt.Errorf("stats: source %q kind is required", source.ID)
	}
	if source.Lifetime == "" {
		return fmt.Errorf("stats: source %q lifetime is required", source.ID)
	}
	if source.Owner.Kind == "" || source.Owner.ID == "" {
		return fmt.Errorf("stats: source %q owner kind and ID are required", source.ID)
	}
	if source.EndTick != 0 && source.EndTick < source.StartTick {
		return fmt.Errorf("stats: source %q ends before it starts", source.ID)
	}
	return nil
}

// Mutation describes one all-or-nothing source-set change. Every source and
// removal is checked before any live state is replaced.
type Mutation struct {
	Replace []Source   `json:"replace,omitempty"`
	Remove  []SourceID `json:"remove,omitempty"`
}

// Snapshot is the copied, deterministic public view for one entity.
type Snapshot struct {
	Entity   EntityID `json:"entity"`
	Revision uint64   `json:"revision"`
	Sources  []Source `json:"sources"`
}
