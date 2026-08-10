package item

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Authority owns each player's container state. Session commands are the only
// writers; snapshots are copies for presentation, persistence, and networking.
type Authority struct {
	mu       sync.RWMutex
	players  map[string]*State
	trades   TradeCatalog
	services ServiceCatalog
}

// NewAuthority creates an empty owner registry. Rule catalogs and initial owner
// states must be installed before command registration captures replay state.
func NewAuthority() *Authority { return &Authority{players: make(map[string]*State)} }

// SetTradeCatalog replaces server-owned vendor pricing rules with a defensive
// copy. Clients submit vendor identity and item intent, never price results.
func (authority *Authority) SetTradeCatalog(catalog TradeCatalog) {
	copyCatalog := make(TradeCatalog, len(catalog))
	for vendor, terms := range catalog {
		copyCatalog[strings.ToLower(strings.TrimSpace(vendor))] = terms
	}
	authority.mu.Lock()
	authority.trades = copyCatalog
	authority.mu.Unlock()
}

// SetServiceCatalog replaces server-owned item-service recipes defensively.
func (authority *Authority) SetServiceCatalog(catalog ServiceCatalog) {
	copyCatalog := make(ServiceCatalog, len(catalog))
	for id, rule := range catalog {
		rule.ID = strings.TrimSpace(rule.ID)
		rule.ConsumeSlots = append([]string(nil), rule.ConsumeSlots...)
		copyCatalog[strings.ToLower(strings.TrimSpace(id))] = rule
	}
	authority.mu.Lock()
	authority.services = copyCatalog
	authority.mu.Unlock()
}

// Register installs a validated copy of one owner's initial item state. Sharing
// the caller's State pointer would create a mutation path around session commands.
func (authority *Authority) Register(owner string, state *State) error {
	owner = strings.TrimSpace(owner)
	if owner == "" || state == nil {
		return fmt.Errorf("item: owner and state are required")
	}
	layout, itemsByID, placements := state.Snapshot()
	items := make([]Item, 0, len(itemsByID))
	for _, candidate := range itemsByID {
		items = append(items, candidate)
	}
	sort.Slice(items, func(left, right int) bool { return items[left].ID < items[right].ID })
	owned, err := NewState(layout, items, placements)
	if err != nil {
		return err
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, exists := authority.players[owner]; exists {
		return fmt.Errorf("item: owner %q is already registered", owner)
	}
	authority.players[owner] = owned
	return nil
}

func (authority *Authority) move(owner, itemID string, destination Placement, placeHeld bool) (string, error) {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	state, found := authority.players[owner]
	if !found {
		return "", fmt.Errorf("item: unknown owner %q", owner)
	}
	if placeHeld {
		return state.PlaceHeld(itemID, destination)
	}
	return "", state.Move(itemID, destination)
}

func (authority *Authority) selectWeaponSet(owner string, set int) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	state, found := authority.players[owner]
	if !found {
		return fmt.Errorf("item: unknown owner %q", owner)
	}
	return state.SelectWeaponSet(set)
}

func (authority *Authority) sellHeld(owner, itemID, vendor, category string) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	state, found := authority.players[owner]
	if !found {
		return fmt.Errorf("item: unknown owner %q", owner)
	}
	terms, err := authority.trades.Terms(vendor)
	if err != nil {
		return err
	}
	_, err = state.sellHeldForGold(itemID, category, terms)
	return err
}

func (authority *Authority) buyToHeld(owner, itemID, vendor string) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	state, found := authority.players[owner]
	if !found {
		return fmt.Errorf("item: unknown owner %q", owner)
	}
	terms, err := authority.trades.Terms(vendor)
	if err != nil {
		return err
	}
	_, err = state.buyToHeldForGold(itemID, terms)
	return err
}

func (authority *Authority) completeService(owner, service string) error {
	authority.mu.Lock()
	defer authority.mu.Unlock()
	state, found := authority.players[owner]
	if !found {
		return fmt.Errorf("item: unknown owner %q", owner)
	}
	rule, err := authority.services.Rule(service)
	if err != nil {
		return err
	}
	_, err = state.completeService(rule)
	return err
}

// Snapshot returns copied facts for presentation, persistence, and networking.
// Mutating the returned maps or layout never changes authority.
func (authority *Authority) Snapshot(owner string) (Layout, map[string]Item, map[string]Placement, error) {
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	state, found := authority.players[owner]
	if !found {
		return Layout{}, nil, nil, fmt.Errorf("item: unknown owner %q", owner)
	}
	layout, items, placements := state.Snapshot()
	return layout, items, placements, nil
}

// Export creates a self-contained archive suitable for disconnect recovery,
// durable saves, or transfer to another realm process.
func (authority *Authority) Export(owner string) ([]byte, error) {
	owner, err := normalizeOwner(owner)
	if err != nil {
		return nil, err
	}
	authority.mu.RLock()
	defer authority.mu.RUnlock()
	state, found := authority.players[owner]
	if !found {
		return nil, fmt.Errorf("item: unknown owner %q", owner)
	}
	return MarshalArchive(state)
}

// Restore validates a complete archive before atomically installing it. The
// old owner state remains untouched if verification or reconstruction fails.
func (authority *Authority) Restore(owner string, encoded []byte) error {
	owner, err := normalizeOwner(owner)
	if err != nil {
		return err
	}
	state, err := UnmarshalArchive(encoded)
	if err != nil {
		return err
	}
	authority.mu.Lock()
	authority.players[owner] = state
	authority.mu.Unlock()
	return nil
}
