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
	if layout.BeltCapacity < 0 {
		return nil, fmt.Errorf("item: belt capacity cannot be negative")
	}
	for container, grid := range layout.Grids {
		if !isGrid(container) || grid.Width <= 0 || grid.Height <= 0 {
			return nil, fmt.Errorf("item: %q requires positive grid dimensions", container)
		}
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
	placementIDs := make([]string, 0, len(placements))
	for id := range placements {
		placementIDs = append(placementIDs, id)
	}
	slices.Sort(placementIDs)
	for _, id := range placementIDs {
		placement := placements[id]
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

// Snapshot returns copies suitable for UI, persistence, or network snapshots.
// Callers may edit the returned maps without changing authority.
func (state *State) Snapshot() (Layout, map[string]Item, map[string]Placement) {
	layout := state.layout
	layout.Grids = make(map[Container]Grid, len(state.layout.Grids))
	for container, grid := range state.layout.Grids {
		layout.Grids[container] = grid
	}
	items := make(map[string]Item, len(state.items))
	for id, candidate := range state.items {
		candidate.BodySlots = slices.Clone(candidate.BodySlots)
		items[id] = candidate
	}
	placements := make(map[string]Placement, len(state.placements))
	for id, placement := range state.placements {
		placements[id] = placement
	}
	return layout, items, placements
}

// PlaceHeld applies Diablo II's grid-drop rule. An empty footprint accepts the
// held item. A footprint touching exactly one item swaps that item into the
// hand. Touching two or more items is ambiguous, so nothing moves.
func (state *State) PlaceHeld(id string, destination Placement) (string, error) {
	candidate, found := state.items[id]
	if !found {
		return "", fmt.Errorf("item: unknown item %q", id)
	}
	current, found := state.placements[id]
	if !found || current.Container != ContainerHeld {
		return "", fmt.Errorf("item: %q is not held", id)
	}
	if !isGrid(destination.Container) {
		return "", fmt.Errorf("item: held grid placement requires inventory, stash, or cube")
	}
	if err := state.validateGridBounds(candidate, destination); err != nil {
		return "", err
	}
	overlaps := state.gridOverlaps(candidate, destination)
	if len(overlaps) > 1 {
		return "", fmt.Errorf("item: %q overlaps multiple items", id)
	}
	if len(overlaps) == 0 {
		state.placements[id] = destination
		return "", nil
	}
	displaced := overlaps[0]
	// Both assignments happen only after every check passes. Observers can never
	// see two held items or a half-completed swap.
	state.placements[id] = destination
	state.placements[displaced] = Placement{Container: ContainerHeld}
	return displaced, nil
}

func (state *State) validate(candidate Item, destination Placement) error {
	switch destination.Container {
	case ContainerWorld:
		return nil
	case ContainerInventory, ContainerStash, ContainerCube:
		return state.validateGrid(candidate, destination)
	case ContainerEquipment, ContainerHireling:
		return state.validateEquipment(candidate, destination)
	case ContainerBelt:
		return state.validateBelt(candidate, destination)
	case ContainerHeld:
		return state.requireEmpty(candidate.ID, ContainerHeld, "")
	case ContainerVendor, ContainerQuest:
		if destination.Slot == "" {
			return fmt.Errorf("item: %s placement requires a service slot", destination.Container)
		}
		return state.requireEmpty(candidate.ID, destination.Container, destination.Slot)
	default:
		return fmt.Errorf("item: unsupported container %q", destination.Container)
	}
}

func (state *State) validateGrid(candidate Item, destination Placement) error {
	if err := state.validateGridBounds(candidate, destination); err != nil {
		return err
	}
	overlaps := state.gridOverlaps(candidate, destination)
	if len(overlaps) > 0 {
		return fmt.Errorf("item: %q overlaps %q", candidate.ID, overlaps[0])
	}
	return nil
}

func (state *State) validateGridBounds(candidate Item, destination Placement) error {
	grid, configured := state.layout.Grids[destination.Container]
	if !configured {
		return fmt.Errorf("item: %s grid is not available", destination.Container)
	}
	if destination.X < 0 || destination.Y < 0 || destination.X+candidate.Width > grid.Width || destination.Y+candidate.Height > grid.Height {
		return fmt.Errorf("item: %q does not fit in %s at %d,%d", candidate.ID, destination.Container, destination.X, destination.Y)
	}
	return nil
}

func (state *State) gridOverlaps(candidate Item, destination Placement) []string {
	overlaps := make([]string, 0, 2)
	for id, placed := range state.placements {
		if id == candidate.ID || placed.Container != destination.Container {
			continue
		}
		other := state.items[id]
		if rectanglesOverlap(destination.X, destination.Y, candidate.Width, candidate.Height, placed.X, placed.Y, other.Width, other.Height) {
			overlaps = append(overlaps, id)
		}
	}
	slices.Sort(overlaps)
	return overlaps
}

func isGrid(container Container) bool {
	return container == ContainerInventory || container == ContainerStash || container == ContainerCube
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
