package item

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

func TestArchiveRoundTripPreservesEveryContainerAndActiveWeapons(t *testing.T) {
	layout := Layout{Grids: map[Container]Grid{
		ContainerInventory: {Width: 10, Height: 4}, ContainerStash: {Width: 6, Height: 8}, ContainerCube: {Width: 3, Height: 4},
	}, BeltCapacity: 4, ActiveWeaponSet: 1, VendorGrid: Grid{Width: 10, Height: 10}, Gold: GoldBalance{Carried: 123, Stashed: 456}}
	items := []Item{
		{ID: "inventory", Code: "box", Width: 1, Height: 1},
		{ID: "stash", Code: "box", Width: 1, Height: 1},
		{ID: "cube", Code: "box", Width: 1, Height: 1},
		{ID: "primary", Code: "ssd", Width: 1, Height: 3, BodySlots: []string{"rarm"}},
		{ID: "alternate", Code: "axe", Width: 2, Height: 3, BodySlots: []string{"rarm"}},
		{ID: "hireling", Code: "cap", Width: 2, Height: 2, BodySlots: []string{"head"}},
		{ID: "belt", Code: "hp1", Width: 1, Height: 1, BeltEligible: true},
		{ID: "held", Code: "rin", Width: 1, Height: 1, BaseCost: 777, Presentation: Presentation{InventoryDC6: "inv.dc6", WorldDC6: "drop.dc6", WorldAnimated: true}},
		{ID: "quest", Code: "gem", Width: 1, Height: 1},
		{ID: "vendor", Code: "wnd", Width: 1, Height: 2},
		{ID: "world", Code: "gld", Width: 1, Height: 1},
	}
	placements := map[string]Placement{
		"inventory": {Container: ContainerInventory, X: 1, Y: 2}, "stash": {Container: ContainerStash, X: 3, Y: 4},
		"cube": {Container: ContainerCube, X: 2, Y: 3}, "primary": {Container: ContainerEquipment, Slot: "rarm", WeaponSet: 0},
		"alternate": {Container: ContainerEquipment, Slot: "rarm", WeaponSet: 1}, "hireling": {Container: ContainerHireling, Slot: "head"},
		"belt": {Container: ContainerBelt, BeltSlot: 3}, "held": {Container: ContainerHeld},
		"quest": {Container: ContainerQuest, Slot: "socket_input"}, "vendor": {Container: ContainerVendor, Slot: "weapons", Page: 2, X: 4, Y: 5},
		"world": {Container: ContainerWorld},
	}
	state, err := NewState(layout, items, placements)
	if err != nil {
		t.Fatal(err)
	}
	first, err := MarshalArchive(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalArchive(first)
	if err != nil {
		t.Fatal(err)
	}
	gotLayout, gotItems, gotPlacements := restored.Snapshot()
	wantLayout, wantItems, wantPlacements := state.Snapshot()
	if !reflect.DeepEqual(gotLayout, wantLayout) || !reflect.DeepEqual(gotItems, wantItems) || !reflect.DeepEqual(gotPlacements, wantPlacements) {
		t.Fatalf("restored state differs\nlayout: %#v\nitems: %#v\nplacements: %#v", gotLayout, gotItems, gotPlacements)
	}
	second, err := MarshalArchive(restored)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("equivalent state did not produce deterministic archive bytes")
	}
}

func TestArchiveRevalidatesPlacementInvariants(t *testing.T) {
	payload := archivedState{
		Layout: archivedLayout{Grids: []archivedGrid{{Container: ContainerInventory, Width: 1, Height: 1}}},
		Items: []archivedItem{
			{ID: "one", Code: "gem", Width: 1, Height: 1},
			{ID: "two", Code: "gem", Width: 1, Height: 1},
		},
		Placements: []archivedPlacement{
			{ItemID: "one", Container: ContainerInventory},
			{ItemID: "two", Container: ContainerInventory},
		},
	}
	encodedState, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(encodedState)
	encoded, err := json.Marshal(archiveEnvelope{Version: ArchiveVersion, Checksum: hex.EncodeToString(sum[:]), State: encodedState})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := UnmarshalArchive(encoded); err == nil {
		t.Fatal("checksummed overlapping placements were accepted")
	}
}

func TestArchiveRejectsTamperingAndUnknownVersions(t *testing.T) {
	state, err := NewState(Layout{}, []Item{{ID: "held", Code: "rin", Width: 1, Height: 1}}, map[string]Placement{"held": {Container: ContainerHeld}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := MarshalArchive(state)
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatal(err)
	}
	envelope["checksum"] = "0000000000000000000000000000000000000000000000000000000000000000"
	tampered, _ := json.Marshal(envelope)
	if _, err := UnmarshalArchive(tampered); err == nil {
		t.Fatal("tampered archive was accepted")
	}
	envelope["version"] = float64(ArchiveVersion + 1)
	unknown, _ := json.Marshal(envelope)
	if _, err := UnmarshalArchive(unknown); err == nil {
		t.Fatal("unknown archive version was accepted")
	}
}

func TestAuthorityRestoreIsAtomicAndSupportsRealmHandoff(t *testing.T) {
	source := NewAuthority()
	state, err := NewState(Layout{ActiveWeaponSet: 1}, []Item{{ID: "held", Code: "rin", Width: 1, Height: 1}}, map[string]Placement{"held": {Container: ContainerHeld}})
	if err != nil {
		t.Fatal(err)
	}
	if err := source.Register("alice", state); err != nil {
		t.Fatal(err)
	}
	encoded, err := source.Export("alice")
	if err != nil {
		t.Fatal(err)
	}
	destination := NewAuthority()
	if err := destination.Restore("alice", encoded); err != nil {
		t.Fatal(err)
	}
	layout, _, placements, err := destination.Snapshot("alice")
	if err != nil || layout.ActiveWeaponSet != 1 || placements["held"].Container != ContainerHeld {
		t.Fatalf("handoff snapshot = %#v, %#v, %v", layout, placements, err)
	}
	if err := destination.Restore("alice", []byte("not json")); err == nil {
		t.Fatal("invalid restore was accepted")
	}
	layout, _, placements, err = destination.Snapshot("alice")
	if err != nil || layout.ActiveWeaponSet != 1 || placements["held"].Container != ContainerHeld {
		t.Fatalf("failed restore mutated authority = %#v, %#v, %v", layout, placements, err)
	}
}
