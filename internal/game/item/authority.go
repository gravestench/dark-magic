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
	mu      sync.RWMutex
	players map[string]*State
}

func NewAuthority() *Authority { return &Authority{players: make(map[string]*State)} }

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
