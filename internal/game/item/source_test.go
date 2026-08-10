package item

import "testing"

func TestControllerQueuesMoveForOneFixedTick(t *testing.T) {
	controller := &Controller{}
	if err := controller.Move(MovePayload{ItemID: "potion", Destination: Placement{Container: ContainerHeld}}); err != nil {
		t.Fatal(err)
	}
	source, err := NewSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(7)
	if len(commands) != 1 || commands[0].Tick != 7 || commands[0].Player != "alice" || commands[0].Sequence != 1 {
		t.Fatalf("commands = %#v", commands)
	}
	if commands := source.Commands(8); commands != nil && len(commands) != 0 {
		t.Fatalf("request was emitted twice: %#v", commands)
	}
}

func TestControllerPreservesMoveAndWeaponSelectionOrder(t *testing.T) {
	controller := &Controller{}
	if err := controller.Move(MovePayload{ItemID: "sword", Destination: Placement{Container: ContainerHeld}}); err != nil {
		t.Fatal(err)
	}
	if err := controller.SelectWeaponSet(1); err != nil {
		t.Fatal(err)
	}
	source, err := NewSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(9)
	if len(commands) != 2 || commands[0].Kind != MoveCommand || commands[1].Kind != WeaponSetCommand {
		t.Fatalf("commands = %#v", commands)
	}
	if commands[0].Sequence != 1 || commands[1].Sequence != 2 {
		t.Fatalf("sequences = %d, %d", commands[0].Sequence, commands[1].Sequence)
	}
}
