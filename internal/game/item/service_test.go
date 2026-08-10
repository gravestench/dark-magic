package item

import "testing"

func TestServiceConsumesAuthoredSocketsAndDebitsGoldAtomically(t *testing.T) {
	state, err := NewState(Layout{Gold: GoldBalance{Carried: 500}}, []Item{
		{ID: "sword", Code: "ssd", Width: 1, Height: 3}, {ID: "rune", Code: "r01", Width: 1, Height: 1},
	}, map[string]Placement{
		"sword": {Container: ContainerQuest, Slot: "socket_target"}, "rune": {Container: ContainerQuest, Slot: "socket_material"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rule := ServiceRule{ID: "socket", TargetSlot: "socket_target", ConsumeSlots: []string{"socket_material"}, GoldCost: 125}
	target, err := state.completeService(rule)
	if err != nil || target != "sword" {
		t.Fatalf("service = %q, %v", target, err)
	}
	layout, items, placements := state.Snapshot()
	if layout.Gold.Carried != 375 || len(items) != 1 || len(placements) != 1 {
		t.Fatalf("service snapshot = %#v %#v %#v", layout, items, placements)
	}
	if len(items["sword"].AppliedServices) != 1 || items["sword"].AppliedServices[0] != "socket" {
		t.Fatalf("target = %#v", items["sword"])
	}
}

func TestFailedServiceLeavesInputsAndGoldUntouched(t *testing.T) {
	state, err := NewState(Layout{Gold: GoldBalance{Carried: 10}}, []Item{{ID: "sword", Code: "ssd", Width: 1, Height: 3}}, map[string]Placement{"sword": {Container: ContainerQuest, Slot: "target"}})
	if err != nil {
		t.Fatal(err)
	}
	beforeLayout, beforeItems, beforePlacements := state.Snapshot()
	if _, err := state.completeService(ServiceRule{ID: "imbue", TargetSlot: "target", ConsumeSlots: []string{"material"}, GoldCost: 20}); err == nil {
		t.Fatal("incomplete service was accepted")
	}
	afterLayout, afterItems, afterPlacements := state.Snapshot()
	if afterLayout.Gold != beforeLayout.Gold || len(afterItems) != len(beforeItems) || afterPlacements["sword"] != beforePlacements["sword"] {
		t.Fatal("failed service mutated state")
	}
}

func TestItemSnapshotsCloneAppliedServices(t *testing.T) {
	state, err := NewState(Layout{}, []Item{{ID: "sword", Code: "ssd", Width: 1, Height: 3, AppliedServices: []string{"socket"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, items, _ := state.Snapshot()
	items["sword"].AppliedServices[0] = "corrupt"
	_, again, _ := state.Snapshot()
	if again["sword"].AppliedServices[0] != "socket" {
		t.Fatal("snapshot shared service slice with authority")
	}
}
