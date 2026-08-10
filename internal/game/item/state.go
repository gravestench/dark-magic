package item

import (
	"fmt"
	"slices"
	"strings"
)

// State is one player's item-container snapshot. Maps are private so callers
// cannot move an item without passing all validation first.
type State struct {
	layout     Layout
	items      map[string]Item
	placements map[string]Placement
}

func NewState(layout Layout, items []Item, placements map[string]Placement) (*State, error) {
	if layout.InventoryWidth <= 0 || layout.InventoryHeight <= 0 || layout.BeltCapacity < 0 {
		return nil, fmt.Errorf("item: inventory dimensions must be positive and belt capacity cannot be negative")
	}
	state := &State{layout: layout, items: make(map[string]Item, len(items)), placements: make(map[string]Placement, len(placements))}
	for _, candidate := range items {
		candidate.ID, candidate.Code = strings.TrimSpace(candidate.ID), strings.TrimSpace(candidate.Code)
		if candidate.ID == "" || candidate.Code == "" || candidate.Width <= 0 || candidate.Height <= 0 {
			return nil, fmt.Errorf("item: identity, code, width, and height are required")
		}
		if _, exists := state.items[candidate.ID]; exists {
			return nil, fmt.Errorf("item: duplicate identity %q", candidate.ID)
		}
		state.items[candidate.ID] = candidate
	}
	// Add placements one at a time. Each item sees everything admitted before it,
	// which catches overlap and duplicate-slot mistakes in an imported snapshot.
	for id, placement := range placements {
		if err := state.Move(id, placement); err != nil {
			return nil, err
		}
	}
	return state, nil
}

// Move validates first and mutates second. A rejected move leaves the old
// placement untouched, like moving a toy only after finding an empty shelf.
func (state *State) Move(id string, destination Placement) error {
	candidate, found := state.items[id]
	if !found {
		return fmt.Errorf("item: unknown item %q", id)
	}
	if err := state.validate(candidate, destination); err != nil {
		return err
	}
	state.placements[id] = destination
	return nil
}

func (state *State) Placement(id string) (Placement, bool) {
	placement, found := state.placements[id]
	return placement, found
}

func (state *State) validate(candidate Item, destination Placement) error {
	switch destination.Container {
	case ContainerWorld:
		return nil
	case ContainerInventory:
		return state.validateInventory(candidate, destination)
	case ContainerEquipment:
		return state.validateEquipment(candidate, destination)
	case ContainerBelt:
		return state.validateBelt(candidate, destination)
	case ContainerCursor:
		return state.requireEmpty(candidate.ID, ContainerCursor, "")
	default:
		return fmt.Errorf("item: unsupported container %q", destination.Container)
	}
}

func (state *State) validateInventory(candidate Item, destination Placement) error {
	if destination.X < 0 || destination.Y < 0 || destination.X+candidate.Width > state.layout.InventoryWidth || destination.Y+candidate.Height > state.layout.InventoryHeight {
		return fmt.Errorf("item: %q does not fit in inventory at %d,%d", candidate.ID, destination.X, destination.Y)
	}
	for id, placed := range state.placements {
		if id == candidate.ID || placed.Container != ContainerInventory {
			continue
		}
		other := state.items[id]
		if rectanglesOverlap(destination.X, destination.Y, candidate.Width, candidate.Height, placed.X, placed.Y, other.Width, other.Height) {
			return fmt.Errorf("item: %q overlaps %q", candidate.ID, id)
		}
	}
	return nil
}

func (state *State) validateEquipment(candidate Item, destination Placement) error {
	if destination.Slot == "" || !slices.Contains(candidate.BodySlots, destination.Slot) {
		return fmt.Errorf("item: %q cannot use body slot %q", candidate.ID, destination.Slot)
	}
	return state.requireEmpty(candidate.ID, ContainerEquipment, destination.Slot)
}

func (state *State) validateBelt(candidate Item, destination Placement) error {
	if !candidate.BeltEligible {
		return fmt.Errorf("item: %q is not belt eligible", candidate.ID)
	}
	if destination.BeltSlot < 0 || destination.BeltSlot >= state.layout.BeltCapacity {
		return fmt.Errorf("item: belt slot %d is outside capacity %d", destination.BeltSlot, state.layout.BeltCapacity)
	}
	return state.requireEmpty(candidate.ID, ContainerBelt, fmt.Sprint(destination.BeltSlot))
}

func (state *State) requireEmpty(candidateID string, container Container, slot string) error {
	for id, placement := range state.placements {
		if id == candidateID || placement.Container != container {
			continue
		}
		occupied := placement.Slot
		if container == ContainerBelt {
			occupied = fmt.Sprint(placement.BeltSlot)
		}
		if occupied == slot {
			return fmt.Errorf("item: %s slot %q is occupied by %q", container, slot, id)
		}
	}
	return nil
}

func rectanglesOverlap(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && bx < ax+aw && ay < by+bh && by < ay+ah
}
