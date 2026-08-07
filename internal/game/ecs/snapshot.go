package ecs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/gravestench/akara"
)

const SnapshotVersion = 1

// Snapshot is the canonical, replay-checkable state of runtime-defined ECS
// components at one completed simulation tick.
type Snapshot struct {
	Version    uint32              `json:"version"`
	Tick       uint64              `json:"tick"`
	Entities   []uint64            `json:"entities"`
	Components []ComponentSnapshot `json:"components"`
}

type ComponentSnapshot struct {
	Name      string             `json:"name"`
	Version   uint32             `json:"version"`
	Fields    []FieldSnapshot    `json:"fields"`
	Instances []InstanceSnapshot `json:"instances"`
}

type FieldSnapshot struct {
	Name string          `json:"name"`
	Kind akara.FieldKind `json:"kind"`
}

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
