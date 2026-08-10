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
	if !validWeaponSet(layout.ActiveWeaponSet) {
		return nil, fmt.Errorf("item: active weapon set must be 0 or 1")
	}
	if layout.VendorGrid.Width < 0 || layout.VendorGrid.Height < 0 ||
		(layout.VendorGrid.Width == 0) != (layout.VendorGrid.Height == 0) {
		return nil, fmt.Errorf("item: vendor grid dimensions must both be positive or both be zero")
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

// SelectWeaponSet atomically changes which hand-slot pair is active. Both sets
// remain equipped in authority; presentation and combat snapshots choose the
// active pair instead of moving item identities between containers.
func (state *State) SelectWeaponSet(set int) error {
	if !validWeaponSet(set) {
		return fmt.Errorf("item: weapon set must be 0 or 1")
	}
	state.layout.ActiveWeaponSet = set
	return nil
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

// PlaceHeld applies Diablo II's cursor-item rule. Grid footprints use overlap
// validation; named equipment, belt, and quest/service sockets use one occupant.
// In either case, replacing exactly one item moves it into the hand atomically.
// Vendor buying and selling are separate transactions: they assign catalog
// positions instead of treating the visible vendor page as a player grid.
func (state *State) PlaceHeld(id string, destination Placement) (string, error) {
	candidate, found := state.items[id]
	if !found {
		return "", fmt.Errorf("item: unknown item %q", id)
	}
	current, found := state.placements[id]
	if !found || current.Container != ContainerHeld {
		return "", fmt.Errorf("item: %q is not held", id)
	}
	if isHeldSlot(destination.Container) {
		return state.placeHeldSlot(candidate, destination)
	}
	if !isGrid(destination.Container) {
		return "", fmt.Errorf("item: held placement requires a grid or named-slot destination")
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

func (state *State) placeHeldSlot(candidate Item, destination Placement) (string, error) {
	// Check what the item is allowed to wear before looking at occupancy. This
	// keeps a rejected drop completely motionless, even when a slot has an item.
	var err error
	if destination.Container == ContainerBelt {
		err = state.validateBeltEligibility(candidate, destination)
	} else if isService(destination.Container) {
		err = validateServiceSlot(destination)
	} else {
		err = state.validateEquipmentEligibility(candidate, destination)
	}
	if err != nil {
		return "", err
	}
	displaced := state.slotOccupant(candidate.ID, destination)
	state.placements[candidate.ID] = destination
	if displaced != "" {
		state.placements[displaced] = Placement{Container: ContainerHeld}
	}
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
	case ContainerVendor:
		return state.validateVendor(candidate, destination)
	case ContainerQuest:
		if err := validateServiceSlot(destination); err != nil {
			return err
		}
		return state.requireEmpty(candidate.ID, destination.Container, destination.Slot)
	default:
		return fmt.Errorf("item: unsupported container %q", destination.Container)
	}
}

func (state *State) validateVendor(candidate Item, destination Placement) error {
	if strings.TrimSpace(destination.Slot) == "" {
		return fmt.Errorf("item: vendor placement requires a category")
	}
	if destination.Page < 0 {
		return fmt.Errorf("item: vendor page cannot be negative")
	}
	grid := state.layout.VendorGrid
	if grid.Width <= 0 || grid.Height <= 0 {
		return fmt.Errorf("item: vendor grid is not available")
	}
	if destination.X < 0 || destination.Y < 0 || destination.X+candidate.Width > grid.Width || destination.Y+candidate.Height > grid.Height {
		return fmt.Errorf("item: %q does not fit vendor category %q page %d at %d,%d", candidate.ID, destination.Slot, destination.Page, destination.X, destination.Y)
	}
	for id, placed := range state.placements {
		if id == candidate.ID || placed.Container != ContainerVendor || placed.Slot != destination.Slot || placed.Page != destination.Page {
			continue
		}
		other := state.items[id]
		if rectanglesOverlap(destination.X, destination.Y, candidate.Width, candidate.Height, placed.X, placed.Y, other.Width, other.Height) {
			return fmt.Errorf("item: %q overlaps vendor item %q", candidate.ID, id)
		}
	}
	return nil
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

func isService(container Container) bool {
	return container == ContainerVendor || container == ContainerQuest
}

func isHeldSlot(container Container) bool {
	return container == ContainerEquipment || container == ContainerHireling || container == ContainerBelt || container == ContainerQuest
}

func validateServiceSlot(destination Placement) error {
	if destination.Slot == "" {
		return fmt.Errorf("item: %s placement requires a service slot", destination.Container)
	}
	return nil
}

func (state *State) validateEquipment(candidate Item, destination Placement) error {
	if err := state.validateEquipmentEligibility(candidate, destination); err != nil {
		return err
	}
	if occupied := state.slotOccupant(candidate.ID, destination); occupied != "" {
		return fmt.Errorf("item: %s slot %q is occupied by %q", destination.Container, destination.Slot, occupied)
	}
	return nil
}

func (state *State) validateEquipmentEligibility(candidate Item, destination Placement) error {
	if destination.Slot == "" || !slices.Contains(candidate.BodySlots, destination.Slot) {
		return fmt.Errorf("item: %q cannot use body slot %q", candidate.ID, destination.Slot)
	}
	if usesWeaponSet(destination.Container, destination.Slot) {
		if !validWeaponSet(destination.WeaponSet) {
			return fmt.Errorf("item: weapon set must be 0 or 1")
		}
	} else if destination.WeaponSet != 0 {
		return fmt.Errorf("item: body slot %q is shared and cannot use weapon set %d", destination.Slot, destination.WeaponSet)
	}
	return nil
}

func (state *State) validateBelt(candidate Item, destination Placement) error {
	if err := state.validateBeltEligibility(candidate, destination); err != nil {
		return err
	}
	return state.requireEmpty(candidate.ID, ContainerBelt, fmt.Sprint(destination.BeltSlot))
}

func (state *State) validateBeltEligibility(candidate Item, destination Placement) error {
	if !candidate.BeltEligible {
		return fmt.Errorf("item: %q is not belt eligible", candidate.ID)
	}
	if destination.BeltSlot < 0 || destination.BeltSlot >= state.layout.BeltCapacity {
		return fmt.Errorf("item: belt slot %d is outside capacity %d", destination.BeltSlot, state.layout.BeltCapacity)
	}
	return nil
}

func (state *State) slotOccupant(candidateID string, destination Placement) string {
	for id, placement := range state.placements {
		if id == candidateID || placement.Container != destination.Container {
			continue
		}
		if destination.Container == ContainerBelt && placement.BeltSlot == destination.BeltSlot {
			return id
		}
		if destination.Container != ContainerBelt && placement.Slot == destination.Slot &&
			(!usesWeaponSet(destination.Container, destination.Slot) || placement.WeaponSet == destination.WeaponSet) {
			return id
		}
	}
	return ""
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

func validWeaponSet(set int) bool { return set == 0 || set == 1 }

func usesWeaponSet(container Container, slot string) bool {
	return container == ContainerEquipment && (slot == "rarm" || slot == "larm")
}

func rectanglesOverlap(ax, ay, aw, ah, bx, by, bw, bh int) bool {
	return ax < bx+bw && bx < ax+aw && ay < by+bh && by < ay+ah
}
