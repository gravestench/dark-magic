package simulation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

var ErrAuthoritativeState = errors.New("simulation: invalid authoritative state")

// RuntimeIdentity pins the exact rule package and engine contract used by a
// session. It is state even though it does not change during a tick: replay and
// restore must reject a different rule implementation before executing it.
type RuntimeIdentity struct {
	Recipe RuntimeRecipe `json:"recipe"`
}

// Digest returns the canonical runtime identity hash used to reject replay or
// restore attempts that would execute against different authoritative code.
func (identity RuntimeIdentity) Digest() (string, error) {
	if err := identity.Validate(); err != nil {
		return "", err
	}

	encoded, err := json.Marshal(identity)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(encoded)

	return hex.EncodeToString(sum[:]), nil
}

// RuntimeIdentityFromParticipants finds and validates the runtime identity in
// a participant snapshot, failing closed when deterministic provenance is absent.
func RuntimeIdentityFromParticipants(states []ParticipantState) (RuntimeIdentity, error) {
	for _, state := range states {
		if state.ID != (&IdentityParticipant{}).StateID() {
			continue
		}

		var identity RuntimeIdentity
		if err := json.Unmarshal(state.Data, &identity); err != nil {
			return RuntimeIdentity{}, err
		}

		if err := identity.Validate(); err != nil {
			return RuntimeIdentity{}, err
		}

		return identity, nil
	}

	return RuntimeIdentity{}, fmt.Errorf("%w: runtime identity participant is missing", ErrAuthoritativeState)
}

// Validate rejects incomplete runtime recipes before they can become replay identity.
func (identity RuntimeIdentity) Validate() error {
	if err := identity.Recipe.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrAuthoritativeState, err)
	}

	return nil
}

// IdentityParticipant includes a canonical runtime identity in every initial
// state, checkpoint, checksum, and replay without coupling the session to Lua.
type IdentityParticipant struct{ identity RuntimeIdentity }

// NewIdentityParticipant validates and defensively copies identity so callers
// cannot mutate the replay contract after a session begins.
func NewIdentityParticipant(identity RuntimeIdentity) (*IdentityParticipant, error) {
	if err := identity.Validate(); err != nil {
		return nil, err
	}

	return &IdentityParticipant{identity: cloneRuntimeIdentity(identity)}, nil
}

// StateID supplies the stable participant key shared by snapshots and restores.
func (*IdentityParticipant) StateID() string { return "engine.authoritative_runtime/v1" }

// SnapshotState serializes the immutable runtime identity into checkpoint evidence.
func (participant *IdentityParticipant) SnapshotState() ([]byte, error) {
	return json.Marshal(participant.identity)
}

// RestoreState verifies saved identity without replacing the running code identity.
func (participant *IdentityParticipant) RestoreState(data []byte) error {
	var restored RuntimeIdentity
	if err := json.Unmarshal(data, &restored); err != nil {
		return fmt.Errorf("%w: decode runtime identity: %v", ErrAuthoritativeState, err)
	}

	if err := restored.Validate(); err != nil {
		return err
	}

	expected, err := json.Marshal(participant.identity)
	if err != nil {
		return err
	}

	actual, err := json.Marshal(restored)
	if err != nil {
		return err
	}

	if sha256.Sum256(expected) != sha256.Sum256(actual) {
		return fmt.Errorf("%w: authoritative runtime identity differs", ErrAuthoritativeState)
	}

	return nil
}

// StateStore is an engine-owned collection of opaque, versioned values. Script
// runtimes decide what mutations mean; this store owns the durable bytes and
// participates atomically in session snapshot/restore.
type StateStore struct {
	mu      sync.RWMutex
	stores  map[string]RegisteredState
	version uint32
}

type RegisteredState struct {
	Schema string `json:"schema"`
	Data   []byte `json:"data"`
}

type stateStoreArchive struct {
	Version uint32                     `json:"version"`
	Stores  map[string]RegisteredState `json:"stores"`
}

// NewStateStore creates an empty version-one registry for authoritative opaque state.
func NewStateStore() *StateStore {
	return &StateStore{stores: make(map[string]RegisteredState), version: 1}
}

// StateID supplies the stable participant key used by replay restoration.
func (*StateStore) StateID() string { return "engine.authoritative_state/v1" }

// Register fixes a state's schema and initial bytes before authoritative updates begin.
func (store *StateStore) Register(id, schema string, initial []byte) error {
	id, schema = strings.TrimSpace(id), strings.TrimSpace(schema)
	if id == "" || schema == "" {
		return fmt.Errorf("%w: state ID and schema are required", ErrAuthoritativeState)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.stores[id]; exists {
		return fmt.Errorf("%w: duplicate state %q", ErrAuthoritativeState, id)
	}

	store.stores[id] = RegisteredState{Schema: schema, Data: append([]byte(nil), initial...)}

	return nil
}

// Read returns a defensive copy so readers cannot mutate authoritative bytes outside Replace.
func (store *StateStore) Read(id string) (RegisteredState, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, found := store.stores[id]
	value.Data = append([]byte(nil), value.Data...)

	return value, found
}

// Replace commits one complete value. Callers compute policy before this point;
// errors cannot leave a partially written store entry.
func (store *StateStore) Replace(id, schema string, data []byte) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	current, found := store.stores[id]
	if !found {
		return fmt.Errorf("%w: unknown state %q", ErrAuthoritativeState, id)
	}

	if current.Schema != schema {
		return fmt.Errorf("%w: state %q schema is %q, not %q", ErrAuthoritativeState, id, current.Schema, schema)
	}

	store.stores[id] = RegisteredState{Schema: schema, Data: append([]byte(nil), data...)}

	return nil
}

// SnapshotState captures one lock-consistent archive of all registered state.
func (store *StateStore) SnapshotState() ([]byte, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return json.Marshal(stateStoreArchive{Version: store.version, Stores: cloneRegisteredStates(store.stores)})
}

// RestoreState validates the full registration topology before atomically replacing bytes.
func (store *StateStore) RestoreState(data []byte) error {
	var archive stateStoreArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		return fmt.Errorf("%w: decode state store: %v", ErrAuthoritativeState, err)
	}

	if archive.Version != 1 {
		return fmt.Errorf("%w: state store version %d", ErrAuthoritativeState, archive.Version)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(archive.Stores) != len(store.stores) {
		return fmt.Errorf("%w: registered state count differs", ErrAuthoritativeState)
	}

	for id, current := range store.stores {
		restored, found := archive.Stores[id]
		if !found || restored.Schema != current.Schema {
			return fmt.Errorf("%w: state %q registration or schema differs", ErrAuthoritativeState, id)
		}
	}

	store.stores = cloneRegisteredStates(archive.Stores)

	return nil
}

// Digest hashes the canonical archive used to compare authoritative state across peers.
func (store *StateStore) Digest() (string, error) {
	data, err := store.SnapshotState()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:]), nil
}

// cloneRegisteredStates copies map entries and byte slices so snapshots never alias live state.
func cloneRegisteredStates(source map[string]RegisteredState) map[string]RegisteredState {
	result := make(map[string]RegisteredState, len(source))

	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := source[key]
		result[key] = RegisteredState{Schema: value.Schema, Data: append([]byte(nil), value.Data...)}
	}

	return result
}

// cloneRuntimeIdentity copies mutable slices and maps to freeze the session's provenance.
func cloneRuntimeIdentity(identity RuntimeIdentity) RuntimeIdentity {
	result := identity
	result.Recipe.Packages.Extensions = append([]RuntimePackage(nil), identity.Recipe.Packages.Extensions...)

	result.Recipe.CapabilityVersions = make(map[string]string, len(identity.Recipe.CapabilityVersions))
	for key, value := range identity.Recipe.CapabilityVersions {
		result.Recipe.CapabilityVersions[key] = value
	}

	return result
}
