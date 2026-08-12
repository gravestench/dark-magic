package ecs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/gravestench/akara"
)

// ErrSnapshotVersion rejects snapshots whose canonical schema is unknown.
var ErrSnapshotVersion = fmt.Errorf("game ecs: unsupported snapshot version")

// SnapshotVersion identifies the canonical dynamic-component encoding.
const SnapshotVersion = 1

// Snapshot is the canonical, replay-checkable state of runtime-defined ECS
// components at one completed simulation tick.
type Snapshot struct {
	Version    uint32              `json:"version"`
	Tick       uint64              `json:"tick"`
	Entities   []uint64            `json:"entities"`
	Components []ComponentSnapshot `json:"components"`
}

// ComponentSnapshot retains a runtime schema and its ordered instances.
type ComponentSnapshot struct {
	Name      string             `json:"name"`
	Version   uint32             `json:"version"`
	Fields    []FieldSnapshot    `json:"fields"`
	Instances []InstanceSnapshot `json:"instances"`
}

// FieldSnapshot is the replay-relevant portion of one dynamic field schema.
type FieldSnapshot struct {
	Name string          `json:"name"`
	Kind akara.FieldKind `json:"kind"`
}

// InstanceSnapshot binds canonical values to one stable entity identity.
type InstanceSnapshot struct {
	Entity uint64          `json:"entity"`
	Values []ValueSnapshot `json:"values"`
}

// ValueSnapshot avoids interface-valued JSON and preserves float bit patterns.
type ValueSnapshot struct {
	Bool   *bool   `json:"bool,omitempty"`
	Int    *int64  `json:"int,omitempty"`
	Uint   *uint64 `json:"uint,omitempty"`
	Float  *uint64 `json:"float_bits,omitempty"`
	String *string `json:"string,omitempty"`
	Entity *uint64 `json:"entity,omitempty"`
}

// Snapshot captures a stable world view between engine updates.
func (engine *Engine) Snapshot() (Snapshot, error) {
	engine.updateMu.Lock()
	defer engine.updateMu.Unlock()
	engine.mu.RLock()
	tick := engine.tick
	engine.mu.RUnlock()
	result := Snapshot{Version: SnapshotVersion, Tick: tick}
	for _, entity := range engine.world.Entities() {
		result.Entities = append(result.Entities, uint64(entity))
	}
	for _, store := range akara.DynamicStores(engine.world) {
		schema := store.Schema()
		component := ComponentSnapshot{Name: schema.Name, Version: schema.Version}
		for _, field := range schema.Fields {
			component.Fields = append(component.Fields, FieldSnapshot{Name: field.Name, Kind: field.Kind})
		}
		for _, entity := range store.Entities() {
			reference, found := store.Get(entity)
			if !found {
				return Snapshot{}, fmt.Errorf("game ecs: snapshot %q entity %d disappeared", schema.Name, entity)
			}
			values, err := reference.Snapshot()
			if err != nil {
				return Snapshot{}, fmt.Errorf("game ecs: snapshot %q entity %d: %w", schema.Name, entity, err)
			}
			instance := InstanceSnapshot{Entity: uint64(entity)}
			for _, field := range schema.Fields {
				value, err := snapshotValue(field.Kind, values[field.Name])
				if err != nil {
					return Snapshot{}, fmt.Errorf("game ecs: snapshot %q.%s entity %d: %w", schema.Name, field.Name, entity, err)
				}
				instance.Values = append(instance.Values, value)
			}
			component.Instances = append(component.Instances, instance)
		}
		result.Components = append(result.Components, component)
	}
	return result, nil
}

// Marshal returns the canonical serialized representation.
func (snapshot Snapshot) Marshal() ([]byte, error) { return json.Marshal(snapshot) }

// Checksum returns a lowercase SHA-256 checksum of the canonical representation.
func (snapshot Snapshot) Checksum() (string, error) {
	encoded, err := snapshot.Marshal()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

// UnmarshalSnapshot decodes one snapshot and rejects unknown fields so replay
// files cannot silently depend on data this engine version ignores.
func UnmarshalSnapshot(encoded []byte) (Snapshot, error) {
	var snapshot Snapshot
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("game ecs: decode snapshot: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("trailing value")
		}
		return Snapshot{}, fmt.Errorf("game ecs: decode snapshot: %w", err)
	}
	if snapshot.Version != SnapshotVersion {
		return Snapshot{}, fmt.Errorf("%w: %d", ErrSnapshotVersion, snapshot.Version)
	}
	return snapshot, nil
}

// RestoreSnapshot creates an engine containing the exact runtime-defined world
// state and simulation tick captured by snapshot. Systems are intentionally not
// serialized; the session composition must register the same trusted systems.
func RestoreSnapshot(snapshot Snapshot) (_ *Engine, err error) {
	if snapshot.Version != SnapshotVersion {
		return nil, fmt.Errorf("%w: %d", ErrSnapshotVersion, snapshot.Version)
	}
	engine := New()
	defer func() {
		if err != nil {
			_ = engine.Close()
		}
	}()
	for _, id := range snapshot.Entities {
		if err := engine.world.CreateEntityWithID(akara.Entity(id)); err != nil {
			return nil, fmt.Errorf("game ecs: restore entity %d: %w", id, err)
		}
	}
	for _, component := range snapshot.Components {
		schema := akara.Schema{Name: component.Name, Version: component.Version}
		for _, field := range component.Fields {
			schema.Fields = append(schema.Fields, akara.Field{Name: field.Name, Kind: field.Kind})
		}
		store, err := akara.RegisterSchema(engine.world, schema)
		if err != nil {
			return nil, fmt.Errorf("game ecs: restore schema %q: %w", component.Name, err)
		}
		for _, instance := range component.Instances {
			if len(instance.Values) != len(schema.Fields) {
				return nil, fmt.Errorf("game ecs: restore %q entity %d: %d values for %d fields", component.Name, instance.Entity, len(instance.Values), len(schema.Fields))
			}
			values := make(map[string]any, len(schema.Fields))
			for index, field := range schema.Fields {
				value, err := restoreValue(field.Kind, instance.Values[index])
				if err != nil {
					return nil, fmt.Errorf("game ecs: restore %q.%s entity %d: %w", component.Name, field.Name, instance.Entity, err)
				}
				values[field.Name] = value
			}
			if _, err := store.Set(akara.Entity(instance.Entity), values); err != nil {
				return nil, fmt.Errorf("game ecs: restore %q entity %d: %w", component.Name, instance.Entity, err)
			}
		}
	}
	engine.tick = snapshot.Tick
	return engine, nil
}

// Restore replaces the engine's world with a snapshot while preserving its
// registered systems and fixed-step configuration. Rebuilding subscriptions
// is important: query subscriptions belong to one Akara world and cannot be
// reused after a transactional rollback swaps that world out.
func (engine *Engine) Restore(snapshot Snapshot) error {
	restored, err := RestoreSnapshot(snapshot)
	if err != nil {
		return err
	}

	engine.updateMu.Lock()
	defer engine.updateMu.Unlock()
	engine.mu.Lock()
	defer engine.mu.Unlock()

	candidate := make(map[string]*registeredSystem, len(engine.systems))
	for id, current := range engine.systems {
		definition := cloneDefinition(current.definition)
		definition, remapErr := remapDynamicDefinition(definition, restored.world)
		if remapErr != nil {
			_ = restored.Close()
			return fmt.Errorf("game ecs: restore system %q: %w", id, remapErr)
		}
		options := filterOptions(definition)
		subscription, subscribeErr := restored.world.Subscribe(options...)
		if subscribeErr != nil {
			_ = restored.Close()
			for _, registered := range candidate {
				_ = registered.subscription.Close()
			}
			return fmt.Errorf("game ecs: restore subscription %q: %w", id, subscribeErr)
		}
		candidate[id] = &registeredSystem{definition: definition, subscription: subscription}
	}
	order, err := compileOrder(candidate)
	if err != nil {
		_ = restored.Close()
		return fmt.Errorf("game ecs: restore system order: %w", err)
	}

	oldWorld := engine.world
	for _, registered := range engine.systems {
		_ = registered.subscription.Close()
	}
	engine.world = restored.world
	engine.systems = candidate
	engine.order = order
	engine.tick = snapshot.Tick
	restored.world = akara.NewWorld() // keep restored.Close from closing the adopted world
	_ = restored.Close()
	return oldWorld.Close()
}

func remapDynamicDefinition(definition Definition, world *akara.World) (Definition, error) {
	remap := func(source []akara.ComponentType) ([]akara.ComponentType, error) {
		result := make([]akara.ComponentType, 0, len(source))
		for _, component := range source {
			store, dynamic := component.(*akara.DynamicStore)
			if !dynamic {
				return nil, fmt.Errorf("snapshot rollback cannot restore non-dynamic component filter %T", component)
			}
			restored, found := akara.GetDynamicStore(world, store.Schema().Name)
			if !found {
				return nil, fmt.Errorf("restored component %q is missing", store.Schema().Name)
			}
			result = append(result, restored)
		}
		return result, nil
	}
	var err error
	if definition.All, err = remap(definition.All); err != nil {
		return Definition{}, err
	}
	if definition.Any, err = remap(definition.Any); err != nil {
		return Definition{}, err
	}
	if definition.None, err = remap(definition.None); err != nil {
		return Definition{}, err
	}
	if definition.Read, err = remap(definition.Read); err != nil {
		return Definition{}, err
	}
	if definition.Write, err = remap(definition.Write); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func snapshotValue(kind akara.FieldKind, value any) (ValueSnapshot, error) {
	result := ValueSnapshot{}
	switch kind {
	case akara.FieldBool:
		typed, ok := value.(bool)
		if !ok {
			return result, fmt.Errorf("expected bool, got %T", value)
		}
		result.Bool = &typed
	case akara.FieldInt64:
		typed, ok := value.(int64)
		if !ok {
			return result, fmt.Errorf("expected int64, got %T", value)
		}
		result.Int = &typed
	case akara.FieldUint64:
		typed, ok := value.(uint64)
		if !ok {
			return result, fmt.Errorf("expected uint64, got %T", value)
		}
		result.Uint = &typed
	case akara.FieldFloat64:
		typed, ok := value.(float64)
		if !ok {
			return result, fmt.Errorf("expected float64, got %T", value)
		}
		bits := math.Float64bits(typed)
		result.Float = &bits
	case akara.FieldString:
		typed, ok := value.(string)
		if !ok {
			return result, fmt.Errorf("expected string, got %T", value)
		}
		result.String = &typed
	case akara.FieldEntity:
		typed, ok := value.(akara.Entity)
		if !ok {
			return result, fmt.Errorf("expected entity, got %T", value)
		}
		entity := uint64(typed)
		result.Entity = &entity
	default:
		return result, fmt.Errorf("unsupported field kind %d", kind)
	}
	return result, nil
}

func restoreValue(kind akara.FieldKind, value ValueSnapshot) (any, error) {
	set := 0
	for _, present := range []bool{value.Bool != nil, value.Int != nil, value.Uint != nil, value.Float != nil, value.String != nil, value.Entity != nil} {
		if present {
			set++
		}
	}
	if set != 1 {
		return nil, fmt.Errorf("expected exactly one encoded value, got %d", set)
	}
	switch kind {
	case akara.FieldBool:
		if value.Bool != nil {
			return *value.Bool, nil
		}
	case akara.FieldInt64:
		if value.Int != nil {
			return *value.Int, nil
		}
	case akara.FieldUint64:
		if value.Uint != nil {
			return *value.Uint, nil
		}
	case akara.FieldFloat64:
		if value.Float != nil {
			return math.Float64frombits(*value.Float), nil
		}
	case akara.FieldString:
		if value.String != nil {
			return *value.String, nil
		}
	case akara.FieldEntity:
		if value.Entity != nil {
			return akara.Entity(*value.Entity), nil
		}
	default:
		return nil, fmt.Errorf("unsupported field kind %d", kind)
	}
	return nil, fmt.Errorf("missing value for field kind %d", kind)
}
