package item

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ArchiveVersion identifies the durable per-owner container-state schema.
const ArchiveVersion = 5

// archiveEnvelope is the durable boundary for one player's item authority.
// The checksum catches truncated or accidentally modified handoff data before
// any live state is replaced. The payload has its own version so future save
// migrations do not need to guess which historical shape they received.
type archiveEnvelope struct {
	Version  int             `json:"version"`
	Checksum string          `json:"checksum"`
	State    json.RawMessage `json:"state"`
}

type archivedState struct {
	Layout     archivedLayout      `json:"layout"`
	Items      []archivedItem      `json:"items"`
	Placements []archivedPlacement `json:"placements"`
}

type archivedLayout struct {
	Grids           []archivedGrid     `json:"grids"`
	BeltCapacity    int                `json:"belt_capacity"`
	ActiveWeaponSet int                `json:"active_weapon_set"`
	VendorGrid      archivedDimensions `json:"vendor_grid"`
	Gold            archivedGold       `json:"gold"`
}

type archivedGold struct {
	Carried int64 `json:"carried"`
	Stashed int64 `json:"stashed"`
}

type archivedDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type archivedGrid struct {
	Container Container `json:"container"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
}

type archivedItem struct {
	ID              string               `json:"id"`
	Code            string               `json:"code"`
	Width           int                  `json:"width"`
	Height          int                  `json:"height"`
	BodySlots       []string             `json:"body_slots,omitempty"`
	BeltEligible    bool                 `json:"belt_eligible,omitempty"`
	BaseCost        int64                `json:"base_cost,omitempty"`
	AppliedServices []string             `json:"applied_services,omitempty"`
	Presentation    archivedPresentation `json:"presentation"`
}

type archivedPresentation struct {
	InventoryDC6  string            `json:"inventory_dc6,omitempty"`
	WorldDC6      string            `json:"world_dc6,omitempty"`
	WorldAnimated bool              `json:"world_animated,omitempty"`
	Composite     map[string]string `json:"composite,omitempty"`
	WeaponClass   string            `json:"weapon_class,omitempty"`
}

type archivedPlacement struct {
	ItemID    string    `json:"item_id"`
	Container Container `json:"container"`
	X         int       `json:"x,omitempty"`
	Y         int       `json:"y,omitempty"`
	Slot      string    `json:"slot,omitempty"`
	BeltSlot  int       `json:"belt_slot,omitempty"`
	WeaponSet int       `json:"weapon_set,omitempty"`
	Page      int       `json:"page,omitempty"`
}

// MarshalArchive makes a deterministic, checksummed value snapshot. It never
// exposes the State maps, and identical authority produces identical bytes.
func MarshalArchive(state *State) ([]byte, error) {
	if state == nil {
		return nil, fmt.Errorf("item: state is required")
	}
	layout, items, placements := state.Snapshot()
	payload := archivedState{Layout: archiveLayout(layout)}
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		candidate := items[id]
		payload.Items = append(payload.Items, archivedItem{
			ID: candidate.ID, Code: candidate.Code, Width: candidate.Width, Height: candidate.Height,
			BodySlots: append([]string(nil), candidate.BodySlots...), BeltEligible: candidate.BeltEligible, BaseCost: candidate.BaseCost,
			AppliedServices: append([]string(nil), candidate.AppliedServices...),
			Presentation:    archivePresentation(candidate.Presentation),
		})
		if placement, found := placements[id]; found {
			payload.Placements = append(payload.Placements, archivePlacement(id, placement))
		}
	}
	encodedState, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("item: encode archive state: %w", err)
	}
	sum := sha256.Sum256(encodedState)
	encoded, err := json.Marshal(archiveEnvelope{Version: ArchiveVersion, Checksum: hex.EncodeToString(sum[:]), State: encodedState})
	if err != nil {
		return nil, fmt.Errorf("item: encode archive envelope: %w", err)
	}
	return encoded, nil
}

// UnmarshalArchive verifies and reconstructs authority through NewState. This
// means malformed grids, overlapping items, illegal equipment, multiple held
// items, and every other normal placement invariant are rechecked on restore.
func UnmarshalArchive(encoded []byte) (*State, error) {
	var envelope archiveEnvelope
	if err := decodeStrict(encoded, &envelope); err != nil {
		return nil, fmt.Errorf("item: decode archive envelope: %w", err)
	}
	if envelope.Version < 1 || envelope.Version > ArchiveVersion {
		return nil, fmt.Errorf("item: unsupported archive version %d", envelope.Version)
	}
	expected, err := hex.DecodeString(envelope.Checksum)
	if err != nil || len(expected) != sha256.Size {
		return nil, fmt.Errorf("item: invalid archive checksum")
	}
	actual := sha256.Sum256(envelope.State)
	if !bytes.Equal(expected, actual[:]) {
		return nil, fmt.Errorf("item: archive checksum mismatch")
	}
	var payload archivedState
	if err := decodeStrict(envelope.State, &payload); err != nil {
		return nil, fmt.Errorf("item: decode archive state: %w", err)
	}
	layout, err := restoreLayout(payload.Layout)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(payload.Items))
	for _, candidate := range payload.Items {
		items = append(items, Item{
			ID: candidate.ID, Code: candidate.Code, Width: candidate.Width, Height: candidate.Height,
			BodySlots: append([]string(nil), candidate.BodySlots...), BeltEligible: candidate.BeltEligible, BaseCost: candidate.BaseCost,
			AppliedServices: append([]string(nil), candidate.AppliedServices...),
			Presentation:    restorePresentation(candidate.Presentation),
		})
	}
	placements := make(map[string]Placement, len(payload.Placements))
	for _, entry := range payload.Placements {
		if _, exists := placements[entry.ItemID]; exists {
			return nil, fmt.Errorf("item: duplicate archived placement for %q", entry.ItemID)
		}
		placements[entry.ItemID] = Placement{Container: entry.Container, X: entry.X, Y: entry.Y, Slot: entry.Slot, BeltSlot: entry.BeltSlot, WeaponSet: entry.WeaponSet, Page: entry.Page}
	}
	state, err := NewState(layout, items, placements)
	if err != nil {
		return nil, fmt.Errorf("item: restore archive: %w", err)
	}
	return state, nil
}

func archiveLayout(layout Layout) archivedLayout {
	result := archivedLayout{BeltCapacity: layout.BeltCapacity, ActiveWeaponSet: layout.ActiveWeaponSet, VendorGrid: archivedDimensions{Width: layout.VendorGrid.Width, Height: layout.VendorGrid.Height}, Gold: archivedGold{Carried: layout.Gold.Carried, Stashed: layout.Gold.Stashed}}
	containers := make([]string, 0, len(layout.Grids))
	for container := range layout.Grids {
		containers = append(containers, string(container))
	}
	sort.Strings(containers)
	for _, name := range containers {
		grid := layout.Grids[Container(name)]
		result.Grids = append(result.Grids, archivedGrid{Container: Container(name), Width: grid.Width, Height: grid.Height})
	}
	return result
}

func restoreLayout(archived archivedLayout) (Layout, error) {
	layout := Layout{Grids: make(map[Container]Grid, len(archived.Grids)), BeltCapacity: archived.BeltCapacity, ActiveWeaponSet: archived.ActiveWeaponSet, VendorGrid: Grid{Width: archived.VendorGrid.Width, Height: archived.VendorGrid.Height}, Gold: GoldBalance{Carried: archived.Gold.Carried, Stashed: archived.Gold.Stashed}}
	for _, entry := range archived.Grids {
		if _, exists := layout.Grids[entry.Container]; exists {
			return Layout{}, fmt.Errorf("item: duplicate archived grid %q", entry.Container)
		}
		layout.Grids[entry.Container] = Grid{Width: entry.Width, Height: entry.Height}
	}
	return layout, nil
}

func archivePlacement(id string, placement Placement) archivedPlacement {
	return archivedPlacement{ItemID: id, Container: placement.Container, X: placement.X, Y: placement.Y, Slot: placement.Slot, BeltSlot: placement.BeltSlot, WeaponSet: placement.WeaponSet, Page: placement.Page}
}

func archivePresentation(presentation Presentation) archivedPresentation {
	return archivedPresentation{InventoryDC6: presentation.InventoryDC6, WorldDC6: presentation.WorldDC6, WorldAnimated: presentation.WorldAnimated, Composite: presentation.Composite, WeaponClass: presentation.WeaponClass}
}

func restorePresentation(presentation archivedPresentation) Presentation {
	return Presentation{InventoryDC6: presentation.InventoryDC6, WorldDC6: presentation.WorldDC6, WorldAnimated: presentation.WorldAnimated, Composite: presentation.Composite, WeaponClass: presentation.WeaponClass}
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

func normalizeOwner(owner string) (string, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return "", fmt.Errorf("item: owner is required")
	}
	return owner, nil
}
