package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

const (
	sessionStateID      = "engine.stats/v1"
	sessionStateVersion = 1
)

type sessionState struct {
	Version  int        `json:"version"`
	Entities []Snapshot `json:"entities"`
}

// StateID implements simulation.StateParticipant without importing simulation.
// The interface is satisfied structurally at the session composition root.
func (*Authority) StateID() string { return sessionStateID }

// SnapshotState serializes every target in canonical order. Revisions are
// future-affecting cache/invalidation facts and therefore remain checkpointed.
func (authority *Authority) SnapshotState() ([]byte, error) {
	if authority == nil {
		return nil, fmt.Errorf("stats: authority is required")
	}
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	entities := make([]string, 0, len(authority.entities))
	for entity := range authority.entities {
		entities = append(entities, string(entity))
	}
	sort.Strings(entities)
	state := sessionState{Version: sessionStateVersion}
	for _, entity := range entities {
		id := EntityID(entity)
		state.Entities = append(state.Entities, snapshotLocked(id, authority.entities[id]))
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("stats: encode session state: %w", err)
	}
	return encoded, nil
}

// RestoreState validates the complete payload before replacing live state, so
// a malformed checkpoint cannot leave half of the entities restored.
func (authority *Authority) RestoreState(encoded []byte) error {
	if authority == nil {
		return fmt.Errorf("stats: authority is required")
	}
	var state sessionState
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return fmt.Errorf("stats: decode session state: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("stats: decode session state: trailing data")
	}
	if state.Version != sessionStateVersion {
		return fmt.Errorf("stats: unsupported session-state version %d", state.Version)
	}
	restored := make(map[EntityID]*entityState, len(state.Entities))
	previous := ""
	for _, snapshot := range state.Entities {
		if snapshot.Entity == "" || previous != "" && string(snapshot.Entity) <= previous {
			return fmt.Errorf("stats: session entities are empty, duplicated, or not canonical")
		}
		sources := make(map[SourceID]Source, len(snapshot.Sources))
		previousSource := ""
		for _, source := range snapshot.Sources {
			if err := source.validate(); err != nil {
				return err
			}
			if previousSource != "" && string(source.ID) <= previousSource {
				return fmt.Errorf("stats: sources for entity %q are duplicated or not canonical", snapshot.Entity)
			}
			sources[source.ID] = source.Clone()
			previousSource = string(source.ID)
		}
		restored[snapshot.Entity] = &entityState{revision: snapshot.Revision, sources: sources}
		previous = string(snapshot.Entity)
	}
	authority.mu.Lock()
	authority.entities = restored
	authority.mu.Unlock()
	return nil
}
