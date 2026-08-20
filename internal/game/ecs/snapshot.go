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

// Snapshot captures a stable world view between engine updates. It takes the update lock so no system can mutate the
// world while entity identities, schemas, and positional values are copied into the canonical representation.
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

	// Akara exposes dynamic stores and their entities in canonical order; retaining that order makes Marshal and
	// Checksum deterministic without another sorting pass here.
	for _, store := range akara.DynamicStores(engine.world) {
		component, err := snapshotDynamicStore(store)
		if err != nil {
			return Snapshot{}, err
		}

		result.Components = append(result.Components, component)
	}

	return result, nil
}

// snapshotDynamicStore captures a schema and its instances in the order supplied by Akara. Field values remain
// positional, so preserving schema order is required for restore and checksum compatibility.
func snapshotDynamicStore(store *akara.DynamicStore) (ComponentSnapshot, error) {
	schema := store.Schema()
	component := ComponentSnapshot{Name: schema.Name, Version: schema.Version}

	for _, field := range schema.Fields {
		component.Fields = append(component.Fields, FieldSnapshot{Name: field.Name, Kind: field.Kind})
	}

	for _, entity := range store.Entities() {
		instance, err := snapshotComponentInstance(store, schema, entity)
		if err != nil {
			return ComponentSnapshot{}, err
		}

		component.Instances = append(component.Instances, instance)
	}

	return component, nil
}

// snapshotComponentInstance reads one stable dynamic component and encodes values in schema order. A missing reference
// is treated as an invariant violation rather than silently dropping an entity from the canonical snapshot.
func snapshotComponentInstance(
	store *akara.DynamicStore,
	schema akara.Schema,
	entity akara.Entity,
) (InstanceSnapshot, error) {
	reference, found := store.Get(entity)
	if !found {
		return InstanceSnapshot{}, fmt.Errorf("game ecs: snapshot %q entity %d disappeared", schema.Name, entity)
	}

	values, err := reference.Snapshot()
	if err != nil {
		return InstanceSnapshot{}, fmt.Errorf("game ecs: snapshot %q entity %d: %w", schema.Name, entity, err)
	}

	instance := InstanceSnapshot{Entity: uint64(entity)}

	for _, field := range schema.Fields {
		value, err := snapshotValue(field.Kind, values[field.Name])
		if err != nil {
			return InstanceSnapshot{}, fmt.Errorf(
				"game ecs: snapshot %q.%s entity %d: %w",
				schema.Name,
				field.Name,
				entity,
				err,
			)
		}

		instance.Values = append(instance.Values, value)
	}

	return instance, nil
}

// Marshal returns the canonical JSON representation. Callers may persist or hash these bytes because snapshot capture
// preserves deterministic entity, component, field, and instance ordering.
func (snapshot Snapshot) Marshal() ([]byte, error) { return json.Marshal(snapshot) }

// Checksum returns a lowercase SHA-256 checksum of the canonical representation, allowing peers and replay tooling to
// compare simulation state without retaining the full encoded snapshot.
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

	// A second decode must reach EOF; accepting trailing JSON would make one byte stream describe multiple states.
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

// snapshotValue converts one typed component value into a JSON-safe union. Float values are stored by bits so NaNs,
// signed zero, and other replay-significant representations survive round trips exactly.
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

// restoreValue validates that the encoded union contains exactly one member before converting it to the schema's kind.
// Rejecting ambiguous unions keeps malformed snapshots from producing checksum-valid but ill-defined state.
func restoreValue(kind akara.FieldKind, value ValueSnapshot) (any, error) {
	set := encodedValueCount(value)
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

// encodedValueCount reports how many union members are present. Restore requires exactly one member so the selected
// schema kind cannot mask extra, contradictory data in an untrusted snapshot.
func encodedValueCount(value ValueSnapshot) int {
	presentValues := [...]bool{
		value.Bool != nil,
		value.Int != nil,
		value.Uint != nil,
		value.Float != nil,
		value.String != nil,
		value.Entity != nil,
	}

	count := 0

	for _, present := range presentValues {
		if present {
			count++
		}
	}

	return count
}
