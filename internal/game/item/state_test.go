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

func TestCursorAndBodySlotsAreExclusive(t *testing.T) {
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
