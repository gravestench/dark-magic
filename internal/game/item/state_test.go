package item

import "testing"

func TestMoveValidatesEveryContainerWithoutPartialMutation(t *testing.T) {
	items := []Item{
		{ID: "sword", Code: "ssd", Width: 1, Height: 3, BodySlots: []string{"rarm"}},
		{ID: "potion", Code: "hp1", Width: 1, Height: 1, BeltEligible: true},
	}
	state, err := NewState(Layout{Grids: map[Container]Grid{ContainerInventory: {Width: 10, Height: 4}}, BeltCapacity: 4}, items, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Move("sword", Placement{Container: ContainerInventory, X: 0, Y: 0}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("potion", Placement{Container: ContainerInventory, X: 0, Y: 1}); err == nil {
		t.Fatal("overlapping inventory move was accepted")
	}
	if _, found := state.Placement("potion"); found {
		t.Fatal("rejected move changed the potion placement")
	}
	if err := state.Move("potion", Placement{Container: ContainerBelt, BeltSlot: 3}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("sword", Placement{Container: ContainerEquipment, Slot: "head"}); err == nil {
		t.Fatal("incompatible equipment move was accepted")
	}
	if placement, _ := state.Placement("sword"); placement.Container != ContainerInventory {
		t.Fatalf("rejected equip changed placement: %#v", placement)
	}
}

func TestHeldAndBodySlotsAreExclusive(t *testing.T) {
	items := []Item{
		{ID: "first", Code: "rin", Width: 1, Height: 1, BodySlots: []string{"lring", "rring"}},
		{ID: "second", Code: "rin", Width: 1, Height: 1, BodySlots: []string{"lring", "rring"}},
	}
	state, err := NewState(Layout{Grids: map[Container]Grid{ContainerInventory: {Width: 10, Height: 4}}, BeltCapacity: 4}, items, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Move("first", Placement{Container: ContainerHeld}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("second", Placement{Container: ContainerHeld}); err == nil {
		t.Fatal("second cursor item was accepted")
	}
	if err := state.Move("first", Placement{Container: ContainerEquipment, Slot: "lring"}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("second", Placement{Container: ContainerEquipment, Slot: "lring"}); err == nil {
		t.Fatal("occupied body slot was accepted")
	}
}

func TestPlayerAndHirelingEquipmentHaveIndependentOccupancy(t *testing.T) {
	items := []Item{
		{ID: "player-helm", Code: "cap", Width: 2, Height: 2, BodySlots: []string{"head"}},
		{ID: "hireling-helm", Code: "cap", Width: 2, Height: 2, BodySlots: []string{"head"}},
		{ID: "second-hireling-helm", Code: "cap", Width: 2, Height: 2, BodySlots: []string{"head"}},
	}
	state, err := NewState(Layout{}, items, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Move("player-helm", Placement{Container: ContainerEquipment, Slot: "head"}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("hireling-helm", Placement{Container: ContainerHireling, Slot: "head"}); err != nil {
		t.Fatalf("player equipment incorrectly occupied hireling slot: %v", err)
	}
	if err := state.Move("second-hireling-helm", Placement{Container: ContainerHireling, Slot: "head"}); err == nil {
		t.Fatal("occupied hireling slot accepted a second item")
	}
}

func TestPlaceHeldSwapsEquipmentAndBeltSlots(t *testing.T) {
	items := []Item{
		{ID: "held-ring", Code: "rin", Width: 1, Height: 1, BodySlots: []string{"lring"}},
		{ID: "worn-ring", Code: "rin", Width: 1, Height: 1, BodySlots: []string{"lring"}},
		{ID: "held-potion", Code: "hp1", Width: 1, Height: 1, BeltEligible: true},
		{ID: "belt-potion", Code: "mp1", Width: 1, Height: 1, BeltEligible: true},
	}
	state, err := NewState(Layout{BeltCapacity: 4}, items, map[string]Placement{
		"held-ring":   {Container: ContainerHeld},
		"worn-ring":   {Container: ContainerEquipment, Slot: "lring"},
		"belt-potion": {Container: ContainerBelt, BeltSlot: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	displaced, err := state.PlaceHeld("held-ring", Placement{Container: ContainerEquipment, Slot: "lring"})
	if err != nil || displaced != "worn-ring" {
		t.Fatalf("equipment swap = %q, %v", displaced, err)
	}
	if err := state.Move("worn-ring", Placement{Container: ContainerWorld}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("held-potion", Placement{Container: ContainerHeld}); err != nil {
		t.Fatal(err)
	}
	displaced, err = state.PlaceHeld("held-potion", Placement{Container: ContainerBelt, BeltSlot: 2})
	if err != nil || displaced != "belt-potion" {
		t.Fatalf("belt swap = %q, %v", displaced, err)
	}
}

func TestGridsAndServiceEscrowAreDistinctContainers(t *testing.T) {
	items := []Item{
		{ID: "first", Code: "box", Width: 2, Height: 2},
		{ID: "second", Code: "box", Width: 2, Height: 2},
	}
	layout := Layout{Grids: map[Container]Grid{
		ContainerInventory: {Width: 10, Height: 4},
		ContainerStash:     {Width: 6, Height: 8},
		ContainerCube:      {Width: 3, Height: 4},
	}}
	state, err := NewState(layout, items, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Matching coordinates in different grids do not overlap.
	if err := state.Move("first", Placement{Container: ContainerInventory, X: 0, Y: 0}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("second", Placement{Container: ContainerStash, X: 0, Y: 0}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("first", Placement{Container: ContainerQuest, Slot: "socket_input"}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("second", Placement{Container: ContainerQuest, Slot: "socket_input"}); err == nil {
		t.Fatal("occupied quest-service slot was accepted")
	}
}

func TestPlaceHeldSwapsOneOverlapAndRejectsSeveral(t *testing.T) {
	items := []Item{
		{ID: "held", Code: "big", Width: 2, Height: 2},
		{ID: "one", Code: "gem", Width: 1, Height: 1},
		{ID: "two", Code: "gem", Width: 1, Height: 1},
	}
	state, err := NewState(Layout{Grids: map[Container]Grid{ContainerInventory: {Width: 4, Height: 4}}}, items, map[string]Placement{
		"held": {Container: ContainerHeld},
		"one":  {Container: ContainerInventory, X: 1, Y: 1},
		"two":  {Container: ContainerInventory, X: 3, Y: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	displaced, err := state.PlaceHeld("held", Placement{Container: ContainerInventory, X: 0, Y: 0})
	if err != nil || displaced != "one" {
		t.Fatalf("single-overlap swap = %q, %v", displaced, err)
	}
	if placement, _ := state.Placement("one"); placement.Container != ContainerHeld {
		t.Fatalf("displaced item placement = %#v", placement)
	}

	// Put the large item back in hand, with two small items under its next target.
	if err := state.Move("one", Placement{Container: ContainerInventory, X: 2, Y: 2}); err != nil {
		t.Fatal(err)
	}
	if err := state.Move("held", Placement{Container: ContainerHeld}); err != nil {
		t.Fatal(err)
	}
	before, _ := state.Placement("held")
	if _, err := state.PlaceHeld("held", Placement{Container: ContainerInventory, X: 2, Y: 2}); err == nil {
		t.Fatal("multiple-overlap placement was accepted")
	}
	if after, _ := state.Placement("held"); after != before {
		t.Fatalf("rejected placement changed held item: %#v -> %#v", before, after)
	}
}
