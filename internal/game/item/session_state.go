package item

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	sessionStateID      = "dm.items/v1"
	sessionStateVersion = 1
)

type sessionStateEnvelope struct {
	Version             int                 `json:"version"`
	ConfigurationDigest string              `json:"configuration_digest"`
	Owners              []sessionOwnerState `json:"owners"`
}

type sessionOwnerState struct {
	Owner   string          `json:"owner"`
	Archive json.RawMessage `json:"archive"`
}

type configurationState struct {
	Trades   []configurationTrade   `json:"trades"`
	Services []configurationService `json:"services"`
}

type configurationTrade struct {
	Vendor         string `json:"vendor"`
	BuyMultiplier  int64  `json:"buy_multiplier"`
	SellMultiplier int64  `json:"sell_multiplier"`
	MaxBuy         int64  `json:"max_buy"`
}

type configurationService struct {
	ID           string   `json:"id"`
	TargetSlot   string   `json:"target_slot"`
	ConsumeSlots []string `json:"consume_slots,omitempty"`
	GoldCost     int64    `json:"gold_cost"`
}

// StateID implements simulation.StateParticipant. The versioned identity is
// part of the replay format and must change if this payload becomes incompatible.
func (*Authority) StateID() string { return sessionStateID }

// SnapshotState captures every owner plus the identity of server-owned rules
// that affect command results. It is canonical for equal authority state.
func (authority *Authority) SnapshotState() ([]byte, error) {
	if authority == nil {
		return nil, fmt.Errorf("item: authority is required")
	}
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	digest, err := authority.configurationDigestLocked()
	if err != nil {
		return nil, err
	}
	owners := make([]string, 0, len(authority.players))
	for owner := range authority.players {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	envelope := sessionStateEnvelope{Version: sessionStateVersion, ConfigurationDigest: digest}
	for _, owner := range owners {
		archive, err := MarshalArchive(authority.players[owner])
		if err != nil {
			return nil, fmt.Errorf("item: archive owner %q: %w", owner, err)
		}
		envelope.Owners = append(envelope.Owners, sessionOwnerState{Owner: owner, Archive: archive})
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("item: encode session state: %w", err)
	}
	return encoded, nil
}

// RestoreState validates configuration and every owner archive before replacing
// the live owner set atomically. Server rule catalogs remain process-owned.
func (authority *Authority) RestoreState(encoded []byte) error {
	if authority == nil {
		return fmt.Errorf("item: authority is required")
	}
	var envelope sessionStateEnvelope
	if err := decodeStrict(encoded, &envelope); err != nil {
		return fmt.Errorf("item: decode session state: %w", err)
	}
	if envelope.Version != sessionStateVersion {
		return fmt.Errorf("item: unsupported session-state version %d", envelope.Version)
	}
	authority.mu.RLock()
	digest, err := authority.configurationDigestLocked()
	authority.mu.RUnlock()
	if err != nil {
		return err
	}
	if envelope.ConfigurationDigest != digest {
		return fmt.Errorf("item: authoritative configuration mismatch: replay=%s runtime=%s", envelope.ConfigurationDigest, digest)
	}
	restored := make(map[string]*State, len(envelope.Owners))
	previousOwner := ""
	for _, entry := range envelope.Owners {
		owner, err := normalizeOwner(entry.Owner)
		if err != nil {
			return err
		}
		if previousOwner != "" && owner <= previousOwner {
			return fmt.Errorf("item: session owners are duplicated or not canonical")
		}
		state, err := UnmarshalArchive(entry.Archive)
		if err != nil {
			return fmt.Errorf("item: restore owner %q: %w", owner, err)
		}
		restored[owner] = state
		previousOwner = owner
	}
	authority.mu.Lock()
	currentDigest, err := authority.configurationDigestLocked()
	if err != nil {
		authority.mu.Unlock()
		return err
	}
	if currentDigest != envelope.ConfigurationDigest {
		authority.mu.Unlock()
		return fmt.Errorf("item: authoritative configuration changed during restore")
	}
	authority.players = restored
	authority.mu.Unlock()
	return nil
}

func (authority *Authority) configurationDigestLocked() (string, error) {
	state := configurationState{}
	vendors := make([]string, 0, len(authority.trades))
	for vendor := range authority.trades {
		vendors = append(vendors, vendor)
	}
	sort.Strings(vendors)
	for _, vendor := range vendors {
		terms, err := authority.trades.Terms(vendor)
		if err != nil {
			return "", err
		}
		state.Trades = append(state.Trades, configurationTrade{Vendor: vendor, BuyMultiplier: terms.BuyMultiplier, SellMultiplier: terms.SellMultiplier, MaxBuy: terms.MaxBuy})
	}
	services := make([]string, 0, len(authority.services))
	for id := range authority.services {
		services = append(services, id)
	}
	sort.Strings(services)
	for _, id := range services {
		candidate := authority.services[id]
		candidate.ConsumeSlots = append([]string(nil), candidate.ConsumeSlots...)
		rule, err := (ServiceCatalog{id: candidate}).Rule(id)
		if err != nil {
			return "", err
		}
		state.Services = append(state.Services, configurationService{ID: strings.ToLower(rule.ID), TargetSlot: rule.TargetSlot, ConsumeSlots: append([]string(nil), rule.ConsumeSlots...), GoldCost: rule.GoldCost})
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("item: encode authoritative configuration: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}
