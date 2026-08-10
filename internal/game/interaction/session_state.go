package interaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

const (
	sessionStateID      = "dm.interaction/v1"
	sessionStateVersion = 1
)

type stateEnvelope struct {
	Version             int          `json:"version"`
	ConfigurationDigest string       `json:"configuration_digest"`
	Owners              []ownerState `json:"owners"`
}

type ownerState struct {
	Owner   string  `json:"owner"`
	Context Context `json:"context"`
}

func (*Authority) StateID() string { return sessionStateID }

func (authority *Authority) SnapshotState() ([]byte, error) {
	if authority == nil {
		return nil, fmt.Errorf("interaction: authority is required")
	}
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	digest, err := authority.configurationDigestLocked()
	if err != nil {
		return nil, err
	}
	owners := make([]string, 0, len(authority.owners))
	for owner := range authority.owners {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	envelope := stateEnvelope{Version: sessionStateVersion, ConfigurationDigest: digest}
	for _, owner := range owners {
		envelope.Owners = append(envelope.Owners, ownerState{Owner: owner, Context: cloneContext(authority.owners[owner])})
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("interaction: encode session state: %w", err)
	}
	return encoded, nil
}

func (authority *Authority) RestoreState(encoded []byte) error {
	if authority == nil {
		return fmt.Errorf("interaction: authority is required")
	}
	var envelope stateEnvelope
	if err := decodeStrict(encoded, &envelope); err != nil {
		return fmt.Errorf("interaction: decode session state: %w", err)
	}
	if envelope.Version != sessionStateVersion {
		return fmt.Errorf("interaction: unsupported session-state version %d", envelope.Version)
	}
	authority.mu.RLock()
	digest, err := authority.configurationDigestLocked()
	authority.mu.RUnlock()
	if err != nil {
		return err
	}
	if digest != envelope.ConfigurationDigest {
		return fmt.Errorf("interaction: authoritative configuration mismatch")
	}
	restored := make(map[string]Context, len(envelope.Owners))
	previous := ""
	for _, entry := range envelope.Owners {
		if entry.Owner == "" || previous != "" && entry.Owner <= previous {
			return fmt.Errorf("interaction: owners are empty, duplicated, or not canonical")
		}
		context := Context{}
		if entry.Context.TargetID != "" {
			authority.mu.RLock()
			context, err = authority.resolveLocked(entry.Context.TargetID)
			authority.mu.RUnlock()
			if err != nil {
				return err
			}
			if context.NPC != entry.Context.NPC || context.Vendor != entry.Context.Vendor {
				return fmt.Errorf("interaction: context for %q does not match target", entry.Owner)
			}
		}
		restored[entry.Owner] = context
		previous = entry.Owner
	}
	authority.mu.Lock()
	currentDigest, err := authority.configurationDigestLocked()
	if err == nil && currentDigest == envelope.ConfigurationDigest {
		authority.owners = restored
	} else if err == nil {
		err = fmt.Errorf("interaction: configuration changed during restore")
	}
	authority.mu.Unlock()
	return err
}

func (authority *Authority) configurationDigestLocked() (string, error) {
	ids := make([]string, 0, len(authority.targets))
	for id := range authority.targets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	targets := make([]Target, 0, len(ids))
	for _, id := range ids {
		targets = append(targets, authority.targets[id])
	}
	encoded, err := json.Marshal(targets)
	if err != nil {
		return "", fmt.Errorf("interaction: encode target configuration: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func decodeStrict(encoded []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("trailing data")
	}
	return nil
}
