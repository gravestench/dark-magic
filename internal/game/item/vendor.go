package item

import (
	"fmt"
	"sort"
	"strings"
)

// SellHeld moves a cursor item into an authority-chosen vendor page position.
// The client names only the semantic category; it cannot forge page coordinates.
func (state *State) sellHeld(id, category string) (Placement, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return Placement{}, fmt.Errorf("item: vendor category is required")
	}
	if strings.Contains(category, "/") {
		return Placement{}, fmt.Errorf("item: vendor category cannot contain a slash")
	}
	placement, found := state.placements[id]
	if !found || placement.Container != ContainerHeld {
		return Placement{}, fmt.Errorf("item: %q is not held", id)
	}
	arranged, err := state.arrangeVendor(category, id)
	if err != nil {
		return Placement{}, err
	}
	// Nothing changes until every footprint has a legal position.
	for itemID, destination := range arranged {
		state.placements[itemID] = destination
	}
	return arranged[id], nil
}

// BuyToHeld removes one vendor item from its catalog and puts it in the single
// authoritative hand. Pricing/ledger authority wraps this transfer separately.
func (state *State) buyToHeld(id string) error {
	placement, found := state.placements[id]
	if !found || placement.Container != ContainerVendor {
		return fmt.Errorf("item: %q is not vendor stock", id)
	}
	if err := state.requireEmpty(id, ContainerHeld, ""); err != nil {
		return err
	}
	arranged, err := state.arrangeVendor(placement.Slot, "", id)
	if err != nil {
		return err
	}
	for itemID, destination := range arranged {
		state.placements[itemID] = destination
	}
	state.placements[id] = Placement{Container: ContainerHeld}
	return nil
}

func (state *State) arrangeVendor(category, addedID string, excludedIDs ...string) (map[string]Placement, error) {
	grid := state.layout.VendorGrid
	if grid.Width <= 0 || grid.Height <= 0 {
		return nil, fmt.Errorf("item: vendor grid is not available")
	}
	excluded := make(map[string]struct{}, len(excludedIDs))
	for _, id := range excludedIDs {
		excluded[id] = struct{}{}
	}
	ids := make([]string, 0)
	if addedID != "" {
		ids = append(ids, addedID)
	}
	for id, placement := range state.placements {
		_, skip := excluded[id]
		if id != addedID && !skip && placement.Container == ContainerVendor && placement.Slot == category {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(left, right int) bool {
		leftItem, rightItem := state.items[ids[left]], state.items[ids[right]]
		if leftItem.Code != rightItem.Code {
			return leftItem.Code < rightItem.Code
		}
		return ids[left] < ids[right]
	})
	arranged := make(map[string]Placement, len(ids))
	for _, id := range ids {
		candidate := state.items[id]
		if candidate.Width > grid.Width || candidate.Height > grid.Height {
			return nil, fmt.Errorf("item: %q cannot fit a vendor page", id)
		}
		arranged[id] = firstVendorPosition(category, candidate, arranged, state.items, grid)
	}
	return arranged, nil
}

// Riiablo's recovered VendorGrid walks columns, then rows. Reusing that order
// makes equal stock produce stable pages without presentation choosing cells.
func firstVendorPosition(category string, candidate Item, placed map[string]Placement, items map[string]Item, grid Grid) Placement {
	for page := 0; ; page++ {
		for x := 0; x <= grid.Width-candidate.Width; x++ {
			for y := 0; y <= grid.Height-candidate.Height; y++ {
				destination := Placement{Container: ContainerVendor, Slot: category, Page: page, X: x, Y: y}
				if vendorPositionFree(candidate, destination, placed, items) {
					return destination
				}
			}
		}
	}
}

func vendorPositionFree(candidate Item, destination Placement, placed map[string]Placement, items map[string]Item) bool {
	for id, otherPlacement := range placed {
		if otherPlacement.Page != destination.Page {
			continue
		}
		other := items[id]
		if rectanglesOverlap(destination.X, destination.Y, candidate.Width, candidate.Height, otherPlacement.X, otherPlacement.Y, other.Width, other.Height) {
			return false
		}
	}
	return true
}
